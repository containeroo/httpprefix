package httpprefix

import (
	"context"
	"net/http"
)

type mountKey struct{}

type mountState struct {
	prefix           string
	explicitRedirect bool
}

// URLForRequest generates an external URL for a prefix-free application path.
// It uses the mount installed by MountUnderPrefix or MountUnderPrefixWithOptions.
// Unmounted requests use an empty prefix. Target handling follows RouteURL.
func URLForRequest(r *http.Request, target string) string {
	if mount, ok := r.Context().Value(mountKey{}).(*mountState); ok {
		return RouteURL(mount.prefix, target)
	}
	return RouteURL("", target)
}

// Redirect redirects to a prefix-free application path using the request's mount.
// It avoids a second prefix when redirect rewriting is enabled. Do not pass an
// already-generated URL; external URLs follow RouteURL's rules.
func Redirect(w http.ResponseWriter, r *http.Request, target string, status int) {
	if mount, ok := r.Context().Value(mountKey{}).(*mountState); ok {
		mount.explicitRedirect = true
	}
	http.Redirect(w, r, URLForRequest(r, target), status)
}

func withMount(h http.Handler, prefix string, rewrite bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mount := &mountState{prefix: prefix}
		r = r.WithContext(context.WithValue(r.Context(), mountKey{}, mount))
		if rewrite {
			w = &redirectWriter{ResponseWriter: w, mount: mount}
		}
		h.ServeHTTP(w, r)
	})
}

type redirectWriter struct {
	http.ResponseWriter
	mount       *mountState
	wroteHeader bool
}

// Unwrap preserves access to transport capabilities through ResponseController.
func (w *redirectWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *redirectWriter) WriteHeader(status int) {
	if !w.wroteHeader && isRedirectStatusCode(status) && !w.mount.explicitRedirect {
		if target := w.Header().Get("Location"); target != "" {
			w.Header().Set("Location", RouteURL(w.mount.prefix, target))
		}
	}
	if status >= 200 {
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *redirectWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
