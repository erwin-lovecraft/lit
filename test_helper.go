package lit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouterForTest(w http.ResponseWriter) (Router, Context, func()) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.ContextWithFallback = true
	rtr := &router{
		engine: engine,
		route: route{
			routes: &engine.RouterGroup,
		},
	}

	ginCtx := gin.CreateTestContextOnly(w, engine)
	ctx := litContext{
		Context: ginCtx,
	}

	return rtr, ctx, func() {
		engine.HandleContext(ginCtx)
	}
}

func CreateTestContext(w http.ResponseWriter) Context {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.ContextWithFallback = true
	ginCtx := gin.CreateTestContextOnly(w, engine)

	return newContext(ginCtx)
}
