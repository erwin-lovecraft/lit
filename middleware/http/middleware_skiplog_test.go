package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit"
)

func TestSkipLoggingResponseBodyMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	route, c, hdlRequest := lit.NewRouterForTest(w)
	route.Use(SkipLoggingResponseBodyMiddleware())
	route.Handle(http.MethodGet, "/test", func(context lit.Context) error {
		return nil
	})

	c.SetRequest(httptest.NewRequest(http.MethodGet, "/test", nil))

	hdlRequest()

	rs, exist := c.Get(lit.SkipLoggingResponseBodyKey)
	require.True(t, exist)
	require.Equal(t, true, rs)
}
