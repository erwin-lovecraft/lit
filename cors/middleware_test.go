package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit"
)

func TestMiddleware_Preflight(t *testing.T) {
	origins := []string{"http://example.com"}
	cfg := New(origins)
	handler := Middleware(cfg)

	// Preflight OPTIONS request
	w := httptest.NewRecorder()
	ctx := lit.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", origins[0])
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	ctx.SetRequest(req)

	// Execute middleware
	err := handler(ctx)
	require.NoError(t, err)
}
