package httpprefix

import "testing"

func TestRouteURL(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"", "/kumbuka", "/tools/wiki", "/pages"} {
		t.Run(prefix, func(t *testing.T) {
			for _, target := range []string{
				"/", "/pages/foo", "/assets/v-test/app.js", "/pages/a%20b",
				"/search?q=a%20b&next=%2Fpages%2Ffoo#results", "/pages/a%2Fb#c%20d",
			} {
				equal(t, prefix+target, RouteURL(prefix, target))
			}
			for _, target := range []string{
				"//evil.example", "///evil.example", "/../escape", "/./foo", "/pages/..",
				"/pages/%2e%2e/escape", "/%2E/foo", "/a%2f..%2fb", "/%5cevil.example",
				"/a\\b", "/%2fexample", "/bad%zz", "/a\r\n", "/a%0d%0ab",
			} {
				equal(t, prefix+"/", RouteURL(prefix, target))
			}
			for _, target := range []string{"", "https://example.test/a", "mailto:a@example.test", "#section", "?q=a", "relative.png"} {
				equal(t, target, RouteURL(prefix, target))
			}
		})
	}
}

func TestRouteURLNormalizesPrefix(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"app", "/app/", "https://example.test/app/"} {
		equal(t, "/app/pages/foo", RouteURL(prefix, "/pages/foo"))
	}
	for _, prefix := range []string{"", "/"} {
		equal(t, "/pages/foo", RouteURL(prefix, "/pages/foo"))
	}
}
