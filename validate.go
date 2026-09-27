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
	if isRootRoutePrefix(prefix) {
		return nil
	}

	if !strings.HasPrefix(prefix, "/") {
		return errors.New("route prefix must be an absolute path")
	}

	path := strings.TrimSuffix(prefix[1:], "/")
	for _, segment := range strings.Split(path, "/") {
		if isInvalidRoutePrefixSegment(segment) {
			return errors.New("route prefix must not contain empty or dot segments")
		}
		if !hasOnlyRoutePrefixSegmentChars(segment) {
			return errors.New("route prefix segments must contain only ASCII letters, digits, '-', '.', '_', or '~'")
		}
	}

	return nil
}

// isRootRoutePrefix reports whether prefix represents an unprefixed deployment.
func isRootRoutePrefix(prefix string) bool {
	return prefix == "" || prefix == "/"
}

// isInvalidRoutePrefixSegment reports whether segment is empty or a dot segment.
func isInvalidRoutePrefixSegment(segment string) bool {
	return segment == "" || isDotSegment(segment)
}

// hasOnlyRoutePrefixSegmentChars reports whether every byte is allowed in a route prefix segment.
func hasOnlyRoutePrefixSegmentChars(segment string) bool {
	for i := 0; i < len(segment); i++ {
		if !isRoutePrefixSegmentChar(segment[i]) {
			return false
		}
	}
	return true
}

// isRoutePrefixSegmentChar reports whether c is allowed in a route prefix segment.
func isRoutePrefixSegmentChar(c byte) bool {
	return isAlphaNum(c) || c == '-' || c == '.' || c == '_' || c == '~'
}

// isAlphaNum reports whether c is an ASCII letter or digit.
func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
