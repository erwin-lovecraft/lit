package lit

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/monitoring"
)

func BenchmarkOneRoute(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Get("/ping", func(c Context) error { return nil })
	runRequest(b, r.Handler(), http.MethodGet, "/ping")
}

func BenchmarkLogger(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = os.Stdout
	r := gin.New()
	r.Use(gin.Logger())
	r.GET("/ping", func(c *gin.Context) {})
	runRequest(b, r, http.MethodGet, "/ping")
}

func BenchmarkManyHandlers(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Use(func(context Context) error { return nil })
	r.Use(func(context Context) error { return nil })
	r.Get("/ping", func(c Context) error { return nil })
	runRequest(b, r.Handler(), http.MethodGet, "/ping")
}

func BenchmarkOneRouteWithMultipleMiddleware(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Get("/ping", func(c Context) error { return nil },
		func(c Context) error { return nil },
		func(c Context) error { return nil },
	)
	runRequest(b, r.Handler(), http.MethodGet, "/ping")
}

func Benchmark5Params(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Use(func(context Context) error { return nil })
	r.Get("/param/:param1/:params2/:param3/:param4/:param5", func(c Context) error { return nil })
	runRequest(b, r.Handler(), http.MethodGet, "/param/:param1/:params2/:param3/:param4/:param5")
}

func BenchmarkOneRouteJSON(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	data := struct {
		Status string `json:"status"`
	}{"ok"}

	r := newRouter(gin.New())
	r.Get("/json", func(c Context) error {
		return c.JSON(http.StatusOK, data)
	})
	runRequest(b, r.Handler(), http.MethodGet, "/json")
}

func BenchmarkOneRouteString(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Get("/string", func(c Context) error {
		return c.String(http.StatusOK, "this is a plain text")
	})
	runRequest(b, r.Handler(), http.MethodGet, "/text")
}

func BenchmarkRouterGroup(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	emptyHandler := func(c Context) error { return nil }

	r.Group("/test", func(r Router) {
		r.Get("/endpoint", func(c Context) error {
			return nil
		}, emptyHandler)

		r.Route("/sub", emptyHandler).Get("/endpoint", emptyHandler)
	})
	runRequest(b, r.Handler(), http.MethodGet, "/test/endpoint")
}

func BenchmarkMiddlewareRoot(b *testing.B) {
	monitor, err := monitoring.New(monitoring.Config{
		ServerName: "lit", Environment: "dev", Writer: os.Stdout,
	})
	require.NoError(b, err)
	ctx := monitoring.SetInContext(context.Background(), monitor)

	gin.SetMode(gin.ReleaseMode)
	r := newRouter(gin.New())
	r.Use(rootMiddleware(ctx))
	r.Get("/ping", func(c Context) error { return nil })
	runRequest(b, r.Handler(), http.MethodGet, "/ping")
}

type mockWriter struct {
	headers http.Header
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		http.Header{},
	}
}

func (m *mockWriter) Header() (h http.Header) {
	return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

func runRequest(b *testing.B, hdl http.Handler, method, path string) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	w := newMockWriter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hdl.ServeHTTP(w, req)
	}
}
