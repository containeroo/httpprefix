package httpprefix

import (
	"net/url"
	"strings"
)

// NormalizeRoutePrefix converts user input into a canonical route prefix.
//
// It accepts either a raw path (for example, "api", "/api", "/api///")
// or a full URL (for example, "https://example.com/api/").
//
// Rules:
//   - Empty, whitespace-only, root-like values ("/", "///") return "".
//   - Trailing slashes are removed.
//   - A leading slash is added when missing.
//   - For full URLs, only the URL path is used.
//
// The returned value is always either "" or a string beginning with "/".
func NormalizeRoutePrefix(input string) string {
	s := strings.TrimSpace(input)
	if s == "" || s == "/" {
		return ""
	}

	// Attempt URL parse. Only treat it as a URL if a scheme is present.
	if u, err := url.Parse(s); err == nil && u.Scheme != "" {
		s = u.Path
	}

	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "/")

	if s == "" || s == "/" {
		return ""
	}

	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}

	return s
}
