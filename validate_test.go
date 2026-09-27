package httpprefix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRoutePrefix(t *testing.T) {
	t.Parallel()

	t.Run("empty prefix is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix(""))
	})

	t.Run("root prefix is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/"))
	})

	t.Run("single segment is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/app"))
	})

	t.Run("trailing slash is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/app/"))
	})

	t.Run("multiple segments are valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/tools/app/"))
	})

	t.Run("allowed punctuation is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/a-Z_09.~"))
	})

	t.Run("well known segment is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/.well-known"))
	})

	t.Run("tilde segment is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/~user"))
	})

	t.Run("multiple dots are valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, ValidateRoutePrefix("/..."))
	})

	t.Run("relative path is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("app"))
	})

	t.Run("URL is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("https://example.com/app"))
	})

	t.Run("network path is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("//app"))
	})

	t.Run("double slash root is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("//"))
	})

	t.Run("triple slash root is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("///"))
	})

	t.Run("repeated trailing separator is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/app//"))
	})

	t.Run("empty middle segment is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a//b"))
	})

	t.Run("single dot segment is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/."))
	})

	t.Run("double dot segment is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/.."))
	})

	t.Run("parent segment in middle is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a/../b"))
	})

	t.Run("current segment in middle is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a/./b"))
	})

	t.Run("parent segment before trailing slash is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a/../"))
	})

	t.Run("query is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/app?q=x"))
	})

	t.Run("fragment is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/app#top"))
	})

	t.Run("escaped slash is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a%2fb"))
	})

	t.Run("escaped dot segment is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/%2e%2e"))
	})

	t.Run("malformed escape is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/bad%zz"))
	})

	t.Run("backslash is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a\\b"))
	})

	t.Run("space is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a b"))
	})

	t.Run("leading space is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix(" /a"))
	})

	t.Run("trailing segment space is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a/ "))
	})

	t.Run("newline is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a\n"))
	})

	t.Run("carriage return is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a\r"))
	})

	t.Run("tab is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a\t"))
	})

	t.Run("NUL is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a\x00"))
	})

	t.Run("non ASCII character is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/café"))
	})

	t.Run("colon is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a:b"))
	})

	t.Run("brace is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/{path}"))
	})

	t.Run("asterisk is invalid", func(t *testing.T) {
		t.Parallel()
		assert.Error(t, ValidateRoutePrefix("/a*b"))
	})
}

func TestIsAlphaNum(t *testing.T) {
	t.Parallel()

	t.Run("lowercase letter", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isAlphaNum('a'))
	})

	t.Run("uppercase letter", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isAlphaNum('Z'))
	})

	t.Run("digit", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isAlphaNum('7'))
	})

	t.Run("punctuation", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isAlphaNum('-'))
	})

	t.Run("non ASCII byte", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isAlphaNum(0xff))
	})
}

func TestIsRoutePrefixSegmentChar(t *testing.T) {
	t.Parallel()

	t.Run("alphanumeric", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRoutePrefixSegmentChar('a'))
	})

	t.Run("hyphen", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRoutePrefixSegmentChar('-'))
	})

	t.Run("dot", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRoutePrefixSegmentChar('.'))
	})

	t.Run("underscore", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRoutePrefixSegmentChar('_'))
	})

	t.Run("tilde", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRoutePrefixSegmentChar('~'))
	})

	t.Run("slash", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isRoutePrefixSegmentChar('/'))
	})
}
