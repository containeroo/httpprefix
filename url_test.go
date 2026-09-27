package httpprefix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteURL(t *testing.T) {
	t.Parallel()

	t.Run("root target without prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/", RouteURL("", "/"))
	})

	t.Run("local path without prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/pages/foo", RouteURL("", "/pages/foo"))
	})

	t.Run("local path with prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/pages/foo", RouteURL("/kumbuka", "/pages/foo"))
	})

	t.Run("nested prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tools/wiki/pages/foo", RouteURL("/tools/wiki", "/pages/foo"))
	})

	t.Run("target matching prefix name is still prefixed", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/pages/pages/foo", RouteURL("/pages", "/pages/foo"))
	})

	t.Run("asset path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/assets/v-test/app.js", RouteURL("/kumbuka", "/assets/v-test/app.js"))
	})

	t.Run("escaped space", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/pages/a%20b", RouteURL("/kumbuka", "/pages/a%20b"))
	})

	t.Run("query and fragment are preserved", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/search?q=a%20b&next=%2Fpages%2Ffoo#results", RouteURL("/kumbuka", "/search?q=a%20b&next=%2Fpages%2Ffoo#results"))
	})

	t.Run("escaped slash in ordinary segment is preserved", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/pages/a%2Fb#c%20d", RouteURL("/kumbuka", "/pages/a%2Fb#c%20d"))
	})

	t.Run("network path falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "//evil.example"))
	})

	t.Run("triple slash network path falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "///evil.example"))
	})

	t.Run("parent segment falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/../escape"))
	})

	t.Run("current segment falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/./foo"))
	})

	t.Run("trailing parent segment falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/pages/.."))
	})

	t.Run("escaped parent segment falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/pages/%2e%2e/escape"))
	})

	t.Run("escaped current segment falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/%2E/foo"))
	})

	t.Run("escaped separators exposing parent segment fall back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/a%2f..%2fb"))
	})

	t.Run("escaped backslash falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/%5cevil.example"))
	})

	t.Run("literal backslash falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/a\\b"))
	})

	t.Run("escaped leading slash falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/%2fexample"))
	})

	t.Run("malformed escape falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/bad%zz"))
	})

	t.Run("literal line break falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/a\r\n"))
	})

	t.Run("escaped line break falls back to root", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/kumbuka/", RouteURL("/kumbuka", "/a%0d%0ab"))
	})

	t.Run("empty target remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", RouteURL("/kumbuka", ""))
	})

	t.Run("external URL remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "https://example.test/a", RouteURL("/kumbuka", "https://example.test/a"))
	})

	t.Run("mailto URL remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "mailto:a@example.test", RouteURL("/kumbuka", "mailto:a@example.test"))
	})

	t.Run("fragment reference remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "#section", RouteURL("/kumbuka", "#section"))
	})

	t.Run("query reference remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "?q=a", RouteURL("/kumbuka", "?q=a"))
	})

	t.Run("relative path remains unchanged", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "relative.png", RouteURL("/kumbuka", "relative.png"))
	})
}

func TestRouteURLNormalizesPrefix(t *testing.T) {
	t.Parallel()

	t.Run("prefix without leading slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/app/pages/foo", RouteURL("app", "/pages/foo"))
	})

	t.Run("prefix with trailing slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/app/pages/foo", RouteURL("/app/", "/pages/foo"))
	})

	t.Run("URL prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/app/pages/foo", RouteURL("https://example.test/app/", "/pages/foo"))
	})

	t.Run("empty prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/pages/foo", RouteURL("", "/pages/foo"))
	})

	t.Run("root prefix", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/pages/foo", RouteURL("/", "/pages/foo"))
	})
}

func TestIsAbsolutePathReference(t *testing.T) {
	t.Parallel()

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isAbsolutePathReference("/foo"))
	})

	t.Run("relative path", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isAbsolutePathReference("foo"))
	})
}

func TestIsUnsafePath(t *testing.T) {
	t.Parallel()

	t.Run("network path", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isUnsafePath("//example.com"))
	})

	t.Run("backslash", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isUnsafePath("/a\\b"))
	})

	t.Run("carriage return", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isUnsafePath("/a\rb"))
	})

	t.Run("newline", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isUnsafePath("/a\nb"))
	})

	t.Run("ordinary path", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isUnsafePath("/a/b"))
	})
}

func TestHasDotSegment(t *testing.T) {
	t.Parallel()

	t.Run("current directory segment", func(t *testing.T) {
		t.Parallel()
		assert.True(t, hasDotSegment("/a/./b"))
	})

	t.Run("parent directory segment", func(t *testing.T) {
		t.Parallel()
		assert.True(t, hasDotSegment("/a/../b"))
	})

	t.Run("dots inside segment", func(t *testing.T) {
		t.Parallel()
		assert.False(t, hasDotSegment("/a/.../b"))
	})

	t.Run("ordinary path", func(t *testing.T) {
		t.Parallel()
		assert.False(t, hasDotSegment("/a/b"))
	})
}
