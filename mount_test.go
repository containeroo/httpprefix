package httpprefix

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMountUnderPrefix(t *testing.T) {
	t.Parallel()

	inner := newTestHandler()

	t.Run("empty prefix returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, ""))
	})

	t.Run("whitespace prefix returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, "   "))
	})

	t.Run("root prefix returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, "/"))
	})

	t.Run("root like prefix returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, "///"))
	})

	t.Run("URL without path returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, "https://example.com"))
	})

	t.Run("URL with root path returns original handler", func(t *testing.T) {
		t.Parallel()
		require.Same(t, inner, MountUnderPrefix(inner, "https://example.com/"))
	})

	t.Run("root prefix serves root routes", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/foo", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foo", rec.Body.String())
	})

	t.Run("prefix without leading slash is normalized", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/foo", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foo", rec.Body.String())
	})

	t.Run("prefix with trailing slash is normalized", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua/")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/foo", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foo", rec.Body.String())
	})

	t.Run("full URL prefix is normalized", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "https://example.com/tambua/")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/foo", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foo", rec.Body.String())
	})

	t.Run("bare prefix GET redirects permanently", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
		assert.Equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("bare prefix HEAD redirects permanently", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/tambua", nil))
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
		assert.Equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("bare prefix POST redirects temporarily", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("prefixed route is stripped", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/foo", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "foo", rec.Body.String())
	})

	t.Run("nested prefixed route is stripped", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/api/v1/ok", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("prefix subtree root serves inner root", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua/", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "root", rec.Body.String())
	})

	t.Run("non prefixed path is not found", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefix(inner, "/tambua")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/foo", nil))
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestMountUnderPrefixWithOptions(t *testing.T) {
	t.Parallel()

	inner := newTestHandler()

	t.Run("custom GET redirect code is applied", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithGetHeadRedirectCode(http.StatusMovedPermanently))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("custom POST redirect code is applied", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithOtherRedirectCode(http.StatusFound))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("WithOptions sets GET redirect code", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithOptions(Options{
			GetHeadRedirectCode: http.StatusPermanentRedirect,
			OtherRedirectCode:   http.StatusSeeOther,
		}))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
	})

	t.Run("WithOptions sets POST redirect code", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithOptions(Options{
			GetHeadRedirectCode: http.StatusPermanentRedirect,
			OtherRedirectCode:   http.StatusSeeOther,
		}))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})

	t.Run("invalid GET redirect code falls back to default", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithGetHeadRedirectCode(http.StatusOK))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
	})

	t.Run("invalid POST redirect code falls back to default", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", WithOtherRedirectCode(http.StatusBadRequest))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	})

	t.Run("nil option is ignored", func(t *testing.T) {
		t.Parallel()
		h := MountUnderPrefixWithOptions(inner, "/tambua", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		assert.Equal(t, http.StatusPermanentRedirect, rec.Code)
	})
}

func TestIsRedirectStatusCode(t *testing.T) {
	t.Parallel()

	t.Run("301", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRedirectStatusCode(http.StatusMovedPermanently))
	})

	t.Run("302", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRedirectStatusCode(http.StatusFound))
	})

	t.Run("303", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRedirectStatusCode(http.StatusSeeOther))
	})

	t.Run("307", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRedirectStatusCode(http.StatusTemporaryRedirect))
	})

	t.Run("308", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isRedirectStatusCode(http.StatusPermanentRedirect))
	})

	t.Run("non redirect status", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isRedirectStatusCode(http.StatusOK))
	})
}

func TestRedirectCodeForMethod(t *testing.T) {
	t.Parallel()

	options := Options{
		GetHeadRedirectCode: http.StatusMovedPermanently,
		OtherRedirectCode:   http.StatusSeeOther,
	}

	t.Run("GET uses GET HEAD code", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, http.StatusMovedPermanently, redirectCodeForMethod(http.MethodGet, options))
	})

	t.Run("HEAD uses GET HEAD code", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, http.StatusMovedPermanently, redirectCodeForMethod(http.MethodHead, options))
	})

	t.Run("POST uses other code", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, http.StatusSeeOther, redirectCodeForMethod(http.MethodPost, options))
	})
}

func newTestHandler() *http.ServeMux {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "root")
	})
	inner.HandleFunc("GET /foo", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "foo")
	})
	inner.HandleFunc("GET /api/v1/ok", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})
	return inner
}
