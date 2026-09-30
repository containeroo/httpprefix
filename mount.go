package httpprefix

import "net/http"

const (
	defaultGetHeadRedirectCode = http.StatusPermanentRedirect // 308
	defaultOtherRedirectCode   = http.StatusTemporaryRedirect // 307
)

// Options configures mounting and redirects used by MountUnderPrefixWithOptions.
type Options struct {
	// GetHeadRedirectCode is used for redirects on GET and HEAD requests.
	GetHeadRedirectCode int
	// OtherRedirectCode is used for redirects on non-GET/HEAD requests.
	OtherRedirectCode int
	// RewriteRedirects prefixes application-local Location headers on redirects.
	RewriteRedirects bool
}

// Option mutates Options used by MountUnderPrefixWithOptions.
type Option func(*Options)

// WithRedirectRewriting prefixes application-local redirect destinations.
// Raw Location paths must be prefix-free; use Redirect for explicit redirects.
// External URLs and relative references are preserved using RouteURL's rules.
func WithRedirectRewriting() Option {
	return func(options *Options) { options.RewriteRedirects = true }
}

// WithOptions overwrites all options used by MountUnderPrefixWithOptions.
func WithOptions(opts Options) Option {
	return func(target *Options) {
		*target = opts
	}
}

// WithGetHeadRedirectCode sets the redirect status code for GET and HEAD requests.
func WithGetHeadRedirectCode(code int) Option {
	return func(target *Options) {
		target.GetHeadRedirectCode = code
	}
}

// WithOtherRedirectCode sets the redirect status code for non-GET/HEAD requests.
func WithOtherRedirectCode(code int) Option {
	return func(target *Options) {
		target.OtherRedirectCode = code
	}
}

// defaultOptions returns the default redirect configuration.
func defaultOptions() Options {
	return Options{
		GetHeadRedirectCode: defaultGetHeadRedirectCode,
		OtherRedirectCode:   defaultOtherRedirectCode,
	}
}

// optionsFrom resolves options from defaults and applies Option overrides.
func optionsFrom(opts ...Option) Options {
	options := defaultOptions()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&options)
	}
	options.GetHeadRedirectCode = sanitizeRedirectCode(options.GetHeadRedirectCode, defaultGetHeadRedirectCode)
	options.OtherRedirectCode = sanitizeRedirectCode(options.OtherRedirectCode, defaultOtherRedirectCode)
	return options
}

// sanitizeRedirectCode returns code when it is a supported redirect status, otherwise fallback.
func sanitizeRedirectCode(code, fallback int) int {
	if isRedirectStatusCode(code) {
		return code
	}
	return fallback
}

// isRedirectStatusCode reports whether code is supported for prefix redirects.
func isRedirectStatusCode(code int) bool {
	switch code {
	case http.StatusMovedPermanently, // 301
		http.StatusFound,             // 302
		http.StatusSeeOther,          // 303
		http.StatusTemporaryRedirect, // 307
		http.StatusPermanentRedirect: // 308
		return true
	default:
		return false
	}
}

// prefixRedirectHandler handles exact bare-prefix requests (for example "/app")
// and issues a redirect to the slash-suffixed path ("/app/").
type prefixRedirectHandler struct {
	// prefix is the normalized mount prefix without a trailing slash.
	prefix string
	// options controls which redirect status code is used by request method.
	options Options
}

// ServeHTTP delegates bare-prefix redirect handling for a fixed prefix.
func (h prefixRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlePrefixRedirect(w, r, h.prefix, h.options)
}

// MountUnderPrefix mounts h under route prefix and returns a handler that serves:
//   - prefix + "/" subtree via http.StripPrefix(prefix, h)
//   - bare prefix redirect to prefix + "/"
//
// If prefix normalizes to "", MountUnderPrefix returns h unchanged. This
// includes empty, whitespace-only, root-like values, and full URLs without a
// path.
//
// Redirect status is:
//   - 308 for GET and HEAD
//   - 307 for all other methods
//
// Prefix normalization inside this function matches NormalizeRoutePrefix:
//   - full URLs use only their path
//   - leading slash is added when missing
//   - trailing slashes are removed
//
// This function uses default redirect status codes (GET/HEAD: 308, others: 307).
// Mounted requests carry context for URLForRequest and Redirect. Handler
// Location headers are unchanged unless WithRedirectRewriting is enabled via
// MountUnderPrefixWithOptions, which also supports overriding redirect codes.
func MountUnderPrefix(h http.Handler, prefix string) http.Handler {
	return MountUnderPrefixWithOptions(h, prefix)
}

// MountUnderPrefixWithOptions behaves like MountUnderPrefix and accepts optional
// redirect status code overrides and opt-in handler redirect rewriting.
//
// If prefix normalizes to "", h is returned unchanged.
//
// Allowed redirect codes are 301, 302, 303, 307, and 308.
// Invalid codes are replaced with defaults (GET/HEAD: 308, others: 307).
func MountUnderPrefixWithOptions(h http.Handler, prefix string, opts ...Option) http.Handler {
	options := optionsFrom(opts...)

	prefix = NormalizeRoutePrefix(prefix)
	if prefix == "" {
		return h
	}

	mux := http.NewServeMux()

	// Redirect bare "/tambua" -> "/tambua/" so subtree handlers match.
	// Use a path-only pattern so behavior stays consistent across Go versions.
	mux.Handle(prefix, prefixRedirectHandler{prefix: prefix, options: options})

	// Mount everything under prefix and strip it so internal routes live at "/".
	mux.Handle(prefix+"/", http.StripPrefix(prefix, withMount(h, prefix, options.RewriteRedirects)))

	return mux
}

// handlePrefixRedirect redirects bare prefix requests to the slash-suffixed path.
// It uses configurable status codes, defaulting to 308 for GET/HEAD and 307 for
// other methods.
func handlePrefixRedirect(w http.ResponseWriter, r *http.Request, prefix string, options Options) {
	http.Redirect(w, r, prefix+"/", redirectCodeForMethod(r.Method, options))
}

// redirectCodeForMethod returns the configured redirect status for the request method.
func redirectCodeForMethod(method string, options Options) int {
	if usesGetHeadRedirectCode(method) {
		return options.GetHeadRedirectCode
	}
	return options.OtherRedirectCode
}

// usesGetHeadRedirectCode reports whether method uses the GET/HEAD redirect policy.
func usesGetHeadRedirectCode(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead:
		return true
	default:
		return false
	}
}
