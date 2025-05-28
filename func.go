package lit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdaptGinHandler allows using gin middleware
func AdaptGinHandler(ginHandler gin.HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		litCtx, ok := ctx.(*litContext)
		if !ok {
			return HTTPError{
				Status: http.StatusInternalServerError,
				Code:   http.StatusText(http.StatusInternalServerError),
				Desc:   "Internal server error",
			}
		}

		ginHandler(litCtx.Context)
		return nil
	}
}

func WrapF(f http.HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		f(ctx.Writer(), ctx.Request())
		return nil
	}
}

func WrapH(h http.Handler) HandlerFunc {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.Writer(), ctx.Request())
		return nil
	}
}
