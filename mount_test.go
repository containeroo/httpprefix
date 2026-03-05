package httpprefix

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMountUnderPrefix(t *testing.T) {
	t.Parallel()

	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "root")
	})
	inner.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "foo")
	})
	inner.HandleFunc("GET /api/v1/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})

	t.Run("prefix '/' behaves like root", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusOK, rec.Code)
		equal(t, "foo", rec.Body.String())
	})

	t.Run("prefix without leading slash is normalized", func(t *testing.T) {
		h := MountUnderPrefix(inner, "tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tambua/foo", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusOK, rec.Code)
		equal(t, "foo", rec.Body.String())
	})

	t.Run("prefix with trailing slash is normalized", func(t *testing.T) {
		h := MountUnderPrefix(inner, "/tambua/")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tambua/foo", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusOK, rec.Code)
		equal(t, "foo", rec.Body.String())
	})

	t.Run("empty prefix returns original handler (serves at root)", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		equal(t, http.StatusOK, rec.Code)
		equal(t, "root", rec.Body.String())

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/tambua/foo", nil)
		h.ServeHTTP(rec2, req2)
		equal(t, http.StatusOK, rec2.Code)
		equal(t, "root", rec2.Body.String())
	})

	t.Run("bare prefix GET redirects with 308", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tambua", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusPermanentRedirect, rec.Code)
		equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("bare prefix HEAD redirects with 308", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodHead, "/tambua", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusPermanentRedirect, rec.Code)
		equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("POST to bare prefix redirects with 307 (method preserved)", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tambua", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusTemporaryRedirect, rec.Code)
		equal(t, "/tambua/", rec.Header().Get("Location"))
	})

	t.Run("prefixed paths are stripped and routed to inner handler", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tambua/foo", nil)
		h.ServeHTTP(rec, req)
		equal(t, http.StatusOK, rec.Code)
		equal(t, "foo", rec.Body.String())

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/tambua/api/v1/ok", nil)
		h.ServeHTTP(rec2, req2)
		equal(t, http.StatusOK, rec2.Code)
		equal(t, "ok", rec2.Body.String())
	})

	t.Run("prefix with trailing slash serves inner root", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tambua/", nil)
		h.ServeHTTP(rec, req)

		equal(t, http.StatusOK, rec.Code)
		equal(t, "root", rec.Body.String())
	})

	t.Run("non-prefixed paths 404 when mounted under a prefix", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefix(inner, "/tambua")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		h.ServeHTTP(rec, req)
		equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestMountUnderPrefixWithOptions(t *testing.T) {
	t.Parallel()

	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "root")
	})

	t.Run("custom redirect codes are applied", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefixWithOptions(
			inner,
			"/tambua",
			WithGetHeadRedirectCode(http.StatusMovedPermanently), // 301
			WithOtherRedirectCode(http.StatusFound),              // 302
		)

		getRec := httptest.NewRecorder()
		h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		equal(t, http.StatusMovedPermanently, getRec.Code)
		equal(t, "/tambua/", getRec.Header().Get("Location"))

		postRec := httptest.NewRecorder()
		h.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		equal(t, http.StatusFound, postRec.Code)
		equal(t, "/tambua/", postRec.Header().Get("Location"))
	})

	t.Run("WithOptions can set both redirect codes", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefixWithOptions(
			inner,
			"/tambua",
			WithOptions(Options{
				GetHeadRedirectCode: http.StatusPermanentRedirect, // 308
				OtherRedirectCode:   http.StatusSeeOther,          // 303
			}),
		)

		getRec := httptest.NewRecorder()
		h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		equal(t, http.StatusPermanentRedirect, getRec.Code)

		postRec := httptest.NewRecorder()
		h.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		equal(t, http.StatusSeeOther, postRec.Code)
	})

	t.Run("invalid redirect codes fall back to defaults", func(t *testing.T) {
		t.Parallel()

		h := MountUnderPrefixWithOptions(
			inner,
			"/tambua",
			WithGetHeadRedirectCode(http.StatusOK), // invalid redirect status
			WithOtherRedirectCode(http.StatusBadRequest), // invalid redirect status
		)

		getRec := httptest.NewRecorder()
		h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/tambua", nil))
		equal(t, http.StatusPermanentRedirect, getRec.Code) // default 308

		postRec := httptest.NewRecorder()
		h.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/tambua", nil))
		equal(t, http.StatusTemporaryRedirect, postRec.Code) // default 307
	})
}
