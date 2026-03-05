package httpprefix

import "testing"

func TestNormalizeRoutePrefix(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		equal(t, "", NormalizeRoutePrefix(""))
		equal(t, "", NormalizeRoutePrefix("   "))
		equal(t, "", NormalizeRoutePrefix("/"))
	})

	t.Run("simple paths", func(t *testing.T) {
		t.Parallel()
		equal(t, "/tambua", NormalizeRoutePrefix("tambua"))
		equal(t, "/tambua", NormalizeRoutePrefix("/tambua"))
		equal(t, "/tambua", NormalizeRoutePrefix("/tambua/"))
		equal(t, "/tambua", NormalizeRoutePrefix("   /tambua/   "))
	})

	t.Run("multiple trailing slashes", func(t *testing.T) {
		t.Parallel()
		equal(t, "/api", NormalizeRoutePrefix("/api///"))
	})

	t.Run("full URL with path", func(t *testing.T) {
		t.Parallel()
		equal(t, "/tambua", NormalizeRoutePrefix("https://example.com/tambua"))
		equal(t, "/tambua", NormalizeRoutePrefix("https://example.com/tambua/"))
	})

	t.Run("full URL with no path", func(t *testing.T) {
		t.Parallel()
		equal(t, "", NormalizeRoutePrefix("https://example.com"))
		equal(t, "", NormalizeRoutePrefix("https://example.com/"))
	})

	t.Run("malformed URL treated as path", func(t *testing.T) {
		t.Parallel()
		equal(t, "/://bad-url", NormalizeRoutePrefix("://bad-url"))
	})

	t.Run("root like input", func(t *testing.T) {
		t.Parallel()
		equal(t, "", NormalizeRoutePrefix("///"))
	})
}
