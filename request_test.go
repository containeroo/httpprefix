package httpprefix

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestMountedRedirects(t *testing.T) {
	for _, tt := range []struct {
		name, prefix, target, want string
		rewrite, explicit          bool
	}{
		{"default unchanged", "/app", "/login", "/login", false, false},
		{"login", "/app", "/login?next=%2Ffoo#top", "/app/login?next=%2Ffoo#top", true, false},
		{"external", "/app", "https://example.org/login", "https://example.org/login", true, false},
		{"relative", "/app", "../login?q=1", "../login?q=1", true, false},
		{"escaped", "/app", "/files/a%20b?q=1#top", "/app/files/a%20b?q=1#top", true, false},
		{"unsafe", "/app", "//example.org/login", "/app/", true, false},
		{"overlapping path", "/pages", "/pages/foo", "/pages/pages/foo", true, false},
		{"explicit", "/app", "/login", "/app/login", true, true},
		{"explicit without rewriting", "/app", "/login", "/app/login", false, true},
		{"explicit overlapping path", "/pages", "/pages/foo", "/pages/pages/foo", true, true},
		{"root", "", "/login", "/login", true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.explicit {
					Redirect(w, r, tt.target, http.StatusSeeOther)
				} else {
					w.Header().Set("Location", tt.target)
					w.WriteHeader(http.StatusSeeOther)
				}
			})
			mounted := MountUnderPrefixWithOptions(h, tt.prefix, WithOptions(Options{RewriteRedirects: tt.rewrite}))
			r := httptest.NewRecorder()
			mounted.ServeHTTP(r, httptest.NewRequest(http.MethodGet, tt.prefix+"/", nil))
			if got := r.Header().Get("Location"); got != tt.want {
				t.Fatalf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileServerRedirect(t *testing.T) {
	h := MountUnderPrefixWithOptions(http.FileServer(http.FS(fstest.MapFS{
		"docs/file.txt": &fstest.MapFile{Data: []byte("hello")},
	})), "/app", WithRedirectRewriting())
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/app/docs", nil))
	if r.Code != http.StatusMovedPermanently || r.Header().Get("Location") != "docs/" {
		t.Fatalf("relative file redirect: status=%d location=%q", r.Code, r.Header().Get("Location"))
	}
}

func TestRedirectResponseSemantics(t *testing.T) {
	for _, status := range []int{201, 301, 302, 303, 304, 307, 308} {
		h := MountUnderPrefixWithOptions(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", "/target")
			w.WriteHeader(status)
			w.WriteHeader(status)
		}), "/app", WithRedirectRewriting())
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/app/", nil))
		want := "/app/target"
		if status == 201 || status == 304 {
			want = "/target"
		}
		if got := r.Header().Get("Location"); got != want {
			t.Fatalf("status %d: Location=%q, want %q", status, got, want)
		}
	}
}

func TestMountedResponseController(t *testing.T) {
	h := MountUnderPrefixWithOptions(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Fatal(err)
		}
	}), "/app", WithRedirectRewriting())
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/app/", nil))
	if !r.Flushed {
		t.Fatal("response controller did not reach the underlying writer")
	}
}

func TestMountURLAndMuxRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(URLForRequest(r, "/login")))
	})
	h := MountUnderPrefixWithOptions(mux, "/app", WithRedirectRewriting())
	for _, path := range []string{"/app/assets", "/app/assets/"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if path == "/app/assets" && r.Header().Get("Location") != "/app/assets/" {
			t.Fatalf("incorrect mux redirect: %s", r.Header().Get("Location"))
		}
		if path == "/app/assets/" && r.Body.String() != "/app/login" {
			t.Fatalf("incorrect request URL: %s", r.Body.String())
		}
	}
}
