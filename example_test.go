package httpprefix_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/containeroo/httpprefix"
)

func ExampleNormalizeRoutePrefix() {
	fmt.Println(httpprefix.NormalizeRoutePrefix(""))
	fmt.Println(httpprefix.NormalizeRoutePrefix("/"))
	fmt.Println(httpprefix.NormalizeRoutePrefix("app"))
	fmt.Println(httpprefix.NormalizeRoutePrefix("https://example.com/app/"))
	// Output:
	//
	//
	// /app
	// /app
}

func ExampleMountUnderPrefix() {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "root")
	})
	inner.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})

	h := httpprefix.MountUnderPrefix(inner, "/app")

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/app", nil))
	fmt.Println(rec1.Code, rec1.Header().Get("Location"))

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/app/health", nil))
	fmt.Println(rec2.Code, rec2.Body.String())

	// Output:
	// 308 /app/
	// 200 ok
}

func ExampleMountUnderPrefixWithOptions() {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "root")
	})

	h := httpprefix.MountUnderPrefixWithOptions(
		inner,
		"/app",
		httpprefix.WithGetHeadRedirectCode(http.StatusMovedPermanently), // 301
		httpprefix.WithOtherRedirectCode(http.StatusFound),              // 302
	)

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/app", nil))
	fmt.Println(rec1.Code, rec1.Header().Get("Location"))

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/app", nil))
	fmt.Println(rec2.Code, rec2.Header().Get("Location"))

	// Output:
	// 301 /app/
	// 302 /app/
}

func ExampleValidateRoutePrefix() {
	input := "/app/"
	if err := httpprefix.ValidateRoutePrefix(input); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(httpprefix.NormalizeRoutePrefix(input))
	fmt.Println(httpprefix.ValidateRoutePrefix("/app/../other") != nil)
	// Output:
	// /app
	// true
}

func ExampleRouteURL() {
	fmt.Println(httpprefix.RouteURL("/app", "/"))
	fmt.Println(httpprefix.RouteURL("/app", "/pages/foo?q=bar#top"))
	fmt.Println(httpprefix.RouteURL("", "/pages/foo"))
	fmt.Println(httpprefix.RouteURL("/pages", "/pages/foo"))
	// Output:
	// /app/
	// /app/pages/foo?q=bar#top
	// /pages/foo
	// /pages/pages/foo
}
