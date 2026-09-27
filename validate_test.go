package httpprefix

import "testing"

func TestValidateRoutePrefix(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "/", "/app", "/app/", "/tools/app/", "/a-Z_09.~", "/.well-known", "/~user", "/..."} {
		t.Run("valid "+value, func(t *testing.T) {
			if err := ValidateRoutePrefix(value); err != nil {
				t.Fatalf("ValidateRoutePrefix(%q): %v", value, err)
			}
		})
	}
	for _, value := range []string{
		"app", "https://example.com/app", "//app", "//", "///", "/app//", "/a//b",
		"/.", "/..", "/a/../b", "/a/./b", "/a/../", "/app?q=x", "/app#top",
		"/a%2fb", "/%2e%2e", "/bad%zz", "/a\\b", "/a b", " /a", "/a/ ",
		"/a\n", "/a\r", "/a\t", "/a\x00", "/café", "/a:b", "/{path}", "/a*b",
	} {
		t.Run("invalid "+value, func(t *testing.T) {
			if err := ValidateRoutePrefix(value); err == nil {
				t.Fatalf("ValidateRoutePrefix(%q) accepted malformed input", value)
			}
		})
	}
}
