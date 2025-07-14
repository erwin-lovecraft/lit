package lit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdaptGinHandler(t *testing.T) {
	r := NewRouter(context.Background())
	// gin handler writes custom status and body
	ginH := func(gc *gin.Context) {
		gc.String(http.StatusTeapot, "gin")
	}
	r.Route("/").Get("/adapt", AdaptGinHandler(ginH))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/adapt", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusTeapot {
		t.Errorf("AdaptGinHandler status: want %d got %d", http.StatusTeapot, w.Code)
	}
	if body := w.Body.String(); body != "gin" {
		t.Errorf("AdaptGinHandler body: want %q got %q", "gin", body)
	}
}

func TestAdaptGinHandler_WithWrongContext(t *testing.T) {
	// gin handler writes custom status and body
	ginH := func(gc *gin.Context) {
		gc.String(http.StatusTeapot, "gin")
	}

	expErr := HTTPError{
		Status: http.StatusInternalServerError,
		Code:   http.StatusText(http.StatusInternalServerError),
		Desc:   "Internal server error",
	}

	err := AdaptGinHandler(ginH)(litContext{})
	require.Equal(t, expErr, err)
}

func TestWrapF(t *testing.T) {
	r := NewRouter(context.Background())
	hf := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("wrapf"))
	}
	r.Route("/").Get("/wrapf", WrapF(hf))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/wrapf", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("WrapF status: want %d got %d", http.StatusForbidden, w.Code)
	}
	if body := w.Body.String(); body != "wrapf" {
		t.Errorf("WrapF body: want %q got %q", "wrapf", body)
	}
}

func TestWrapH(t *testing.T) {
	r := NewRouter(context.Background())
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("wraph"))
	})
	r.Route("/").Get("/wraph", WrapH(handler))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/wraph", nil)
	r.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("WrapH status: want %d got %d", http.StatusAccepted, w.Code)
	}
	if body := w.Body.String(); body != "wraph" {
		t.Errorf("WrapH body: want %q got %q", "wraph", body)
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := HTTPError{
		Status: http.StatusInternalServerError,
		Code:   http.StatusText(http.StatusInternalServerError),
		Desc:   "Internal server error",
	}

	require.Equal(t, "Status: [500], Code: [Internal Server Error], Desc: [Internal server error]", err.Error())
}
