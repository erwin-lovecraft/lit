package lit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_Defaults(t *testing.T) {
	r := NewRouter(context.Background())
	routes := r.Routes()
	if len(routes) != 0 {
		t.Errorf("Expected no routes initially, got %d", len(routes))
	}
}

func TestNewRouter_WithOptions(t *testing.T) {
	r := NewRouter(context.Background(),
		WithLivenessEndpoint("/healthz"),
		WithProfiling(),
	)

	require.ElementsMatch(t, r.Routes(), RoutesInfo{
		{Method: "GET", Path: "/_/profile/"},
		{Method: "GET", Path: "/_/profile/trace"},
		{Method: "GET", Path: "/_/profile/threadcreate"},
		{Method: "GET", Path: "/_/profile/cmdline"},
		{Method: "GET", Path: "/_/profile/profile"},
		{Method: "GET", Path: "/_/profile/symbol"},
		{Method: "GET", Path: "/_/profile/allocs"},
		{Method: "GET", Path: "/_/profile/block"},
		{Method: "GET", Path: "/_/profile/goroutine"},
		{Method: "GET", Path: "/_/profile/heap"},
		{Method: "GET", Path: "/_/profile/mutex"},
		{Method: "GET", Path: "/healthz"},
		{Method: "POST", Path: "/_/profile/symbol"},
	})
}

func TestGetRouteRegisters(t *testing.T) {
	r := NewRouter(context.Background())
	handlerCalled := false
	r.Route("/api").Get("/ping", func(c Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "pong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/ping", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if got := w.Body.String(); got != "pong" {
		t.Errorf("Expected body 'pong', got '%s'", got)
	}
}

func TestPostRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Post("/p", func(c Context) error {
		return c.String(http.StatusCreated, "post")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/p", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if got := w.Body.String(); got != "post" {
		t.Errorf("Expected body 'post', got '%s'", got)
	}
}

func TestHeadRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Head("/h", func(c Context) error {
		return c.String(http.StatusNoContent, "")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/h", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestDeleteRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Delete("/d", func(c Context) error {
		return c.String(http.StatusOK, "deleted")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/d", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Body.String(); got != "deleted" {
		t.Errorf("Expected body 'deleted', got '%s'", got)
	}
}

func TestPatchRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Patch("/pt", func(c Context) error {
		return c.String(http.StatusAccepted, "patched")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pt", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}
	if got := w.Body.String(); got != "patched" {
		t.Errorf("Expected body 'patched', got '%s'", got)
	}
}

func TestPutRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Put("/pu", func(c Context) error {
		return c.String(http.StatusOK, "put")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/pu", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Body.String(); got != "put" {
		t.Errorf("Expected body 'put', got '%s'", got)
	}
}

func TestOptionsRoute(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/").Options("/op", func(c Context) error {
		return c.String(http.StatusNoContent, "")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/op", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestMatchEmptyMethods(t *testing.T) {
	r := NewRouter(context.Background())
	initial := len(r.Routes())
	r.Route("/").Match([]string{}, "/none", func(c Context) error { return nil })
	after := len(r.Routes())
	if initial != after {
		t.Errorf("Expected no new routes, got %d", after-initial)
	}
}

func TestMatchAny(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/any").Match([]string{"*"}, "/test", func(c Context) error {
		return c.String(http.StatusOK, "any")
	})

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	for _, m := range methods {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(m, "/any/test", nil)
		r.Handler().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for %s, got %d", m, w.Code)
		}
	}
}

func TestMatchSpecificMethods(t *testing.T) {
	r := NewRouter(context.Background())
	r.Route("/m").Match([]string{"GET", "POST"}, "/mp", func(c Context) error {
		return c.String(http.StatusOK, "mp")
	})

	cases := []struct {
		method   string
		wantCode int
		wantBody string
	}{
		{"GET", http.StatusOK, "mp"},
		{"POST", http.StatusOK, "mp"},
		{"PUT", http.StatusNotFound, "404 page not found"},    // not registered
		{"DELETE", http.StatusNotFound, "404 page not found"}, // not registered
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tc.method, "/m/mp", nil)
		r.Handler().ServeHTTP(w, req)
		if w.Code != tc.wantCode {
			t.Errorf("Method %s: expected code %d, got %d", tc.method, tc.wantCode, w.Code)
		}
		if got := w.Body.String(); got != tc.wantBody {
			t.Errorf("Method %s: expected body '%s', got '%s'", tc.method, tc.wantBody, got)
		}
	}
}

func TestRouteAndMiddleware(t *testing.T) {
	r := NewRouter(context.Background())
	mw := func(c Context) error {
		c.Writer().Header().Set("X-Test", "123")
		return nil
	}

	r.Route("/grp", mw).Get("/hello", func(c Context) error {
		return c.String(http.StatusOK, "hi")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/grp/hello", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Test"); got != "123" {
		t.Error("Group middleware not applied")
	}
	if body := w.Body.String(); body != "hi" {
		t.Errorf("Expected 'hi', got '%s'", body)
	}
}

func TestGroupAndMiddleware(t *testing.T) {
	r := NewRouter(context.Background())
	mw := func(c Context) error {
		c.Writer().Header().Set("X-Test", "123")
		return nil
	}

	r.Group("/grp", func(r Router) {
		r.Get("/hello", func(c Context) error {
			return c.String(http.StatusOK, "hi")
		})
	}, mw)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/grp/hello", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Test"); got != "123" {
		t.Error("Group middleware not applied")
	}
	if body := w.Body.String(); body != "hi" {
		t.Errorf("Expected 'hi', got '%s'", body)
	}
}

func TestUseMiddleware(t *testing.T) {
	r := NewRouter(context.Background())
	r.Use(func(c Context) error {
		c.Writer().Header().Set("X-Global", "yes")
		return nil
	})
	r.Route("/").Get("/u", func(c Context) error {
		return c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/u", nil)
	r.Handler().ServeHTTP(w, req)
	if got := w.Header().Get("X-Global"); got != "yes" {
		t.Error("Global middleware not applied")
	}
}

func TestBuildGinHandlers_Order(t *testing.T) {
	h := func(c Context) error { return nil }
	m1 := func(c Context) error { return nil }
	out := buildGinHandlers(h, []HandlerFunc{m1})
	if len(out) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(out))
	}
}

func TestToGinHandler_ErrorAborts(t *testing.T) {
	engine := gin.New()
	h := toGinHandler(func(c Context) error {
		return errors.New("fail")
	})
	called := false
	next := func(c *gin.Context) {
		called = true
	}
	engine.GET("/err", h, next)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/err", nil)
	engine.ServeHTTP(w, req)
	if called {
		t.Error("Handlers after error should not run")
	}
}

func TestStaticFileFS(t *testing.T) {
	r := NewRouter(context.Background())
	fs := http.Dir(".")
	r.Route("/").StaticFileFS("/favicon.ico", "favicon.ico", fs)
	routes := r.Routes()
	found := false
	for _, ri := range routes {
		if ri.Path == "/favicon.ico" && ri.Method == "GET" {
			found = true
			break
		}
	}
	if !found {
		t.Error("StaticFileFS route not registered")
	}
}

func TestStaticFS(t *testing.T) {
	r := NewRouter(context.Background())
	fs := http.Dir(".")
	r.Route("/").StaticFS("/static", fs)
	routes := r.Routes()
	found := false
	for _, ri := range routes {
		if ri.Path == "/static/*filepath" && ri.Method == "GET" {
			found = true
			break
		}
	}
	if !found {
		t.Error("StaticFS route not registered")
	}
}

func TestNewRouter_LivenessOption(t *testing.T) {
	r := NewRouter(context.Background(),
		WithLivenessEndpoint("/healthz"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "OK" {
		t.Errorf("Expected 'OK', got '%s'", body)
	}
}
