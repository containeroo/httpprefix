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
	prefix := strings.TrimSpace(input)
	if isRootRoutePrefix(prefix) {
		return ""
	}

	if path, ok := absoluteURLPath(prefix); ok {
		prefix = path
	}

	prefix = strings.TrimSpace(prefix)
	prefix = strings.TrimRight(prefix, "/")
	if isRootRoutePrefix(prefix) {
		return ""
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return prefix
}

// absoluteURLPath returns the path when value is an absolute URL with a scheme.
func absoluteURLPath(value string) (string, bool) {
	u, err := url.Parse(value)
	if err != nil {
		return "", false
	}
	if u.Scheme == "" {
		return "", false
	}
	return u.Path, true
}
