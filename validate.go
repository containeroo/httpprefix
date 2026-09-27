package httpprefix

import (
	"errors"
	"strings"
)

// ValidateRoutePrefix validates strict path-only deployment configuration.
// It accepts an empty string or "/" for root deployment, and absolute paths
// containing non-empty segments of ASCII letters, digits, '-', '.', '_', or '~'.
// One trailing slash is allowed. Dot segments ("." and "..") are rejected.
// URLs, escapes, whitespace, queries, fragments, and repeated slashes are rejected.
//
// Validate before calling NormalizeRoutePrefix when malformed configuration
// should be rejected rather than normalized permissively. Validation does not
// change the input, and the existing normalizer and mounting helpers remain
// permissive for backward compatibility.
func ValidateRoutePrefix(prefix string) error {
	if prefix == "" || prefix == "/" {
		return nil
	}
	if !strings.HasPrefix(prefix, "/") {
		return errors.New("route prefix must be an absolute path")
	}

	path := strings.TrimSuffix(prefix[1:], "/")
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return errors.New("route prefix must not contain empty or dot segments")
		}
		for i := 0; i < len(segment); i++ {
			c := segment[i]
			if isAlphaNum(c) ||
				c == '-' ||
				c == '.' ||
				c == '_' ||
				c == '~' {
				continue
			}
			return errors.New("route prefix segments must contain only ASCII letters, digits, '-', '.', '_', or '~'")
		}
	}
	return nil
}

// isAlphaNum reports whether c is an ASCII letter or digit.
func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
