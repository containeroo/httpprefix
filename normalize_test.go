package httpprefix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRoutePrefix(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix(""))
	})

	t.Run("whitespace input", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix("   "))
	})

	t.Run("root input", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix("/"))
	})

	t.Run("root like input", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix("///"))
	})

	t.Run("path without leading slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("tambua"))
	})

	t.Run("path with leading slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("/tambua"))
	})

	t.Run("path with trailing slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("/tambua/"))
	})

	t.Run("path with surrounding whitespace", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("   /tambua/   "))
	})

	t.Run("multiple trailing slashes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/api", NormalizeRoutePrefix("/api///"))
	})

	t.Run("full URL with path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("https://example.com/tambua"))
	})

	t.Run("full URL with path and trailing slash", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/tambua", NormalizeRoutePrefix("https://example.com/tambua/"))
	})

	t.Run("full URL with no path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix("https://example.com"))
	})

	t.Run("full URL with root path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", NormalizeRoutePrefix("https://example.com/"))
	})

	t.Run("malformed URL is treated as path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/://bad-url", NormalizeRoutePrefix("://bad-url"))
	})
}

func TestAbsoluteURLPath(t *testing.T) {
	t.Parallel()

	t.Run("absolute URL", func(t *testing.T) {
		t.Parallel()
		path, ok := absoluteURLPath("https://example.com/app/")
		require.True(t, ok)
		assert.Equal(t, "/app/", path)
	})

	t.Run("relative path", func(t *testing.T) {
		t.Parallel()
		path, ok := absoluteURLPath("/app")
		assert.False(t, ok)
		assert.Empty(t, path)
	})

	t.Run("malformed URL", func(t *testing.T) {
		t.Parallel()
		path, ok := absoluteURLPath("%")
		assert.False(t, ok)
		assert.Empty(t, path)
	})
}
