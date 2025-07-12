package cors

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConfigDefaults(t *testing.T) {
	origins := []string{"http://example.com", "https://foo.com"}
	cfg := New(origins)

	// Origins
	require.Equal(t, origins, cfg.underlying.AllowOrigins)

	// Allowed methods
	expectedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}
	require.Equal(t, expectedMethods, cfg.underlying.AllowMethods)

	// Allowed headers
	expectedHeaders := []string{
		"Accept", "Origin", "Content-Type", "Content-Length", "Authorization",
		"traceparent", "tracestate", "baggage",
	}
	require.Equal(t, expectedHeaders, cfg.underlying.AllowHeaders)

	// Expose headers
	require.Equal(t, []string{"Link"}, cfg.underlying.ExposeHeaders)

	// Credentials
	require.True(t, cfg.underlying.AllowCredentials)

	// MaxAge
	require.Equal(t, time.Duration(300), cfg.underlying.MaxAge)
}

func TestSetAllowMethods(t *testing.T) {
	cfg := New(nil)
	methods := []string{"A", "B", "C"}
	cfg.SetAllowMethods(methods...)
	require.Equal(t, methods, cfg.underlying.AllowMethods)
}

func TestSetAllowHeaders(t *testing.T) {
	cfg := New(nil)
	headers := []string{"X-Custom", "Y-Custom"}
	cfg.SetAllowHeaders(headers...)
	require.Equal(t, headers, cfg.underlying.AllowHeaders)
}

func TestSetExposeHeaders(t *testing.T) {
	cfg := New(nil)
	expose := []string{"X-Expose-1", "X-Expose-2"}
	cfg.SetExposeHeaders(expose...)
	require.Equal(t, expose, cfg.underlying.ExposeHeaders)
}

func TestDisableCredentials(t *testing.T) {
	cfg := New(nil)
	cfg.DisableCredentials()
	require.False(t, cfg.underlying.AllowCredentials)
}

func TestSetMaxAge(t *testing.T) {
	cfg := New(nil)
	newAge := 42 * time.Second
	cfg.SetMaxAge(newAge)
	require.Equal(t, newAge, cfg.underlying.MaxAge)
}
