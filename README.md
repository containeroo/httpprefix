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

Redirect status codes:

- `GET`, `HEAD` -> `308 Permanent Redirect`
- all others -> `307 Temporary Redirect`

Normalization inside mount:

- `""` and `"/"` return `h` unchanged
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

- The package does not modify query strings when redirecting.
- Redirection is path-based and method-aware to preserve semantics for non-GET requests.
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
