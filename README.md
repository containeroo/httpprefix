# httpprefix

`httpprefix` is a small Go package for mounting `net/http` handlers under a configurable route prefix.

Use it when the same service should work both:

- at root (`/`) in local/dev
- under a sub-path (for example `/app`) in staging/production

## Install

```bash
go get github.com/containeroo/httpprefix
```

## Quick Start

```go
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/containeroo/httpprefix"
)

func main() {
	// Could come from env/config/flag: "", "/app", "app", or full URL.
	prefix := httpprefix.NormalizeRoutePrefix("https://example.com/app/")

	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})
	inner.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "healthy")
	})

	h := httpprefix.MountUnderPrefix(inner, prefix)
	log.Fatal(http.ListenAndServe(":8080", h))
}
```

With `prefix == "/app"`:

- `GET /app` -> `308 Location: /app/`
- `POST /app` -> `307 Location: /app/`
- `GET /app/health` -> inner `GET /health`

## Custom Redirect Codes

```go
h := httpprefix.MountUnderPrefixWithOptions(
	inner,
	"/app",
	httpprefix.WithGetHeadRedirectCode(http.StatusMovedPermanently), // 301
	httpprefix.WithOtherRedirectCode(http.StatusFound),              // 302
)
```

`WithOptions(httpprefix.Options{...})` is also available when you want to set both values at once.

## API

### `ValidateRoutePrefix(prefix string) error`

Optionally validate strict deployment configuration **before** normalization:

```go
input := "/app/" // from a flag or environment variable
if err := httpprefix.ValidateRoutePrefix(input); err != nil {
    return err
}
prefix := httpprefix.NormalizeRoutePrefix(input) // "/app"
```

Accepts empty or `/` for root deployment, absolute paths with ASCII unreserved
characters (`A-Z`, `a-z`, `0-9`, `-`, `.`, `_`, `~`), and one trailing slash.
Rejects dot segments, empty segments, URLs, percent escapes, whitespace,
backslashes, query strings, and fragments. Validation uses no regular expressions.
Existing normalization and mounting remain permissive for compatibility.

### `RouteURL(prefix, target string) string`

Generate links, form actions, asset URLs, or redirect destinations using the same
prefix as the mount:

```go
httpprefix.RouteURL("/app", "/")                 // "/app/"
httpprefix.RouteURL("/app", "/pages/foo?q=1#top") // "/app/pages/foo?q=1#top"
httpprefix.RouteURL("", "/pages/foo")            // "/pages/foo"
```

The prefix is normalized. Pass **prefix-free application paths**; do not call
this helper again on a generated URL. A prefix can match an internal route name:
`RouteURL("/pages", "/pages/foo")` correctly returns `/pages/pages/foo`.

Local paths retain escaping, query strings, and fragments. Malformed local paths,
network-path references (`//host`), backslashes, line breaks, and decoded dot
segments fall back to the deployment root. External URLs and relative references
remain unchanged. This is not a general URL sanitizer or a redirect authorization
check; apply your application's destination policy separately.

### `NormalizeRoutePrefix(input string) string`

Normalizes configuration input into a canonical prefix:

- returns `""` for empty/root-like values (`""`, `"   "`, `"/"`, `"///"`)
- trims trailing slashes (`"/app///"` -> `"/app"`)
- adds leading slash when needed (`"app"` -> `"/app"`)
- accepts full URLs and uses only path (`"https://x.io/app/"` -> `"/app"`)

### `MountUnderPrefix(h http.Handler, prefix string) http.Handler`

Returns a handler that:

- serves `h` under `prefix + "/"` via `http.StripPrefix`
- redirects bare `prefix` to `prefix + "/"`
- returns `h` unchanged when `prefix` normalizes to `""`

Redirect status codes:

- `GET`, `HEAD` -> `308 Permanent Redirect`
- all others -> `307 Temporary Redirect`

Normalization inside mount matches `NormalizeRoutePrefix`:

- empty, whitespace-only, and root-like values return `h` unchanged
- full URLs use only their path
- full URLs without a path return `h` unchanged
- missing leading slash is added
- trailing slashes are removed

### `MountUnderPrefixWithOptions(h http.Handler, prefix string, opts ...Option) http.Handler`

Same behavior as `MountUnderPrefix`, but lets you override redirect status codes.

### `Options` and `Option`

- `Options.GetHeadRedirectCode`: code for `GET` and `HEAD` redirects
- `Options.OtherRedirectCode`: code for non-`GET`/`HEAD` redirects
- `WithGetHeadRedirectCode(code int)`
- `WithOtherRedirectCode(code int)`
- `WithOptions(opts Options)` to replace both values in one call

Allowed redirect codes are: `301`, `302`, `303`, `307`, `308`.
If you pass anything else, defaults are used (`308` for `GET`/`HEAD`, `307` otherwise).

## Behavior Notes

- If the normalized prefix is empty, the original handler is returned unchanged.
- The package does not modify query strings when redirecting.
- Redirection is path-based and method-aware to preserve semantics for non-GET requests.
- Mounting does not rewrite redirects emitted by the inner handler or `ServeMux`; applications must account for those separately.
- Routing behavior relies on `net/http` `ServeMux` path patterns.

## Versioning

Follow semantic versioning (`vMAJOR.MINOR.PATCH`).

- Breaking API change -> major bump
- Backward-compatible feature -> minor bump
- Fix/documentation-only -> patch

## Development

```bash
go test ./...
```

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
