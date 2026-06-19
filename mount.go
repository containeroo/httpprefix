package httpprefix

import "net/http"

const (
	defaultGetHeadRedirectCode = http.StatusPermanentRedirect // 308
	defaultOtherRedirectCode   = http.StatusTemporaryRedirect // 307
)

// Options configures redirect status codes used by MountUnderPrefixWithOptions.
type Options struct {
	// GetHeadRedirectCode is used for redirects on GET and HEAD requests.
	GetHeadRedirectCode int
	// OtherRedirectCode is used for redirects on non-GET/HEAD requests.
	OtherRedirectCode int
}

// Option mutates Options used by MountUnderPrefixWithOptions.
type Option func(*Options)

// WithOptions overwrites all redirect options used by MountUnderPrefixWithOptions.
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

// sanitizeRedirectCode ensures redirect status codes stay within safe redirect values.
func sanitizeRedirectCode(code, fallback int) int {
	switch code {
	case http.StatusMovedPermanently, // 301
		http.StatusFound,             // 302
		http.StatusSeeOther,          // 303
		http.StatusTemporaryRedirect, // 307
		http.StatusPermanentRedirect: // 308
		return code
	default:
		return fallback
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
// Use MountUnderPrefixWithOptions to override redirect codes.
func MountUnderPrefix(h http.Handler, prefix string) http.Handler {
	return MountUnderPrefixWithOptions(h, prefix)
}

// MountUnderPrefixWithOptions behaves like MountUnderPrefix and accepts optional
// redirect status code overrides.
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
	mux.Handle(prefix+"/", http.StripPrefix(prefix, h))

	return mux
}

// handlePrefixRedirect redirects bare prefix requests to the slash-suffixed path.
// It uses configurable status codes, defaulting to 308 for GET/HEAD and 307 for
// other methods.
func handlePrefixRedirect(w http.ResponseWriter, r *http.Request, prefix string, options Options) {
	status := options.OtherRedirectCode
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		status = options.GetHeadRedirectCode
	}
	http.Redirect(w, r, prefix+"/", status)
}
