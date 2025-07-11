package lit

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter init new
func NewRouter(appCtx context.Context, opts ...RouterOption) Router {
	// Init gin engine
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.ContextWithFallback = true

	// Init router
	r := newRouter(engine)

	// Execute all options
	for _, opt := range opts {
		opt(r)
	}

	// Setup root middleware
	// Includes logging, tracing, panic recovery
	r.Use(rootMiddleware(appCtx))

	return r
}

// router implements Router interface and wrap gin.IRouter
type router struct {
	route
	ginRouter gin.IRouter
	engine    *gin.Engine
}

func newRouter(engine *gin.Engine) *router {
	return &router{
		route: route{
			routes: engine,
		},
		ginRouter: engine,
		engine:    engine,
	}
}

func (r *router) routing(prefix string, middlewares ...gin.HandlerFunc) *router {
	gr := r.ginRouter.Group(prefix, middlewares...)

	return &router{
		route: route{
			routes: gr,
		},
		ginRouter: gr,
		engine:    r.engine,
	}
}

func (r *router) Group(prefix string, routerFunc func(Router), middlewares ...HandlerFunc) Router {
	mws := make([]gin.HandlerFunc, len(middlewares))
	for idx := 0; idx < len(middlewares); idx++ {
		mws[idx] = toGinHandler(middlewares[idx])
	}

	routerFunc(r.routing(prefix, mws...))
	return r
}

func (r *router) Route(prefix string, middlewares ...HandlerFunc) Router {
	mws := make([]gin.HandlerFunc, len(middlewares))
	for idx := 0; idx < len(middlewares); idx++ {
		mws[idx] = toGinHandler(middlewares[idx])
	}

	return r.routing(prefix, mws...)
}

func (r *router) Handler() http.Handler {
	return r.engine.Handler()
}

type route struct {
	routes gin.IRoutes
}

func (r *route) Use(middlewares ...HandlerFunc) {
	mws := make([]gin.HandlerFunc, len(middlewares))
	for idx := 0; idx < len(middlewares); idx++ {
		mws[idx] = toGinHandler(middlewares[idx])
	}

	// Put all middleware to gin middleware
	r.routes.Use(mws...)
}

func (r *route) Handle(method string, relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return &route{
		routes: r.routes.Handle(
			method, relativePath,
			buildGinHandlers(handler, middleware)...,
		),
	}
}

func (r *route) Match(methods []string, relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	if len(methods) == 0 {
		return r // No need to match, just return self
	}

	ginHandlers := buildGinHandlers(handler, middleware)
	if len(methods) == 1 && methods[0] == "*" {
		return &route{
			routes: r.routes.Any(relativePath, ginHandlers...),
		}
	}

	return &route{
		routes: r.routes.Match(methods, relativePath, ginHandlers...),
	}
}

func (r *route) Get(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodGet, relativePath, handler, middleware...)
}

func (r *route) Post(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodPost, relativePath, handler, middleware...)
}

func (r *route) Delete(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodDelete, relativePath, handler, middleware...)
}

func (r *route) Patch(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodPatch, relativePath, handler, middleware...)
}

func (r *route) Put(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodPut, relativePath, handler, middleware...)
}

func (r *route) Options(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodOptions, relativePath, handler, middleware...)
}

func (r *route) Head(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route {
	return r.Handle(http.MethodHead, relativePath, handler, middleware...)
}

func (r *route) StaticFileFS(relativePath string, filepath string, fs http.FileSystem) Route {
	return &route{
		routes: r.routes.StaticFileFS(relativePath, filepath, fs),
	}
}

func (r *route) StaticFS(relativePath string, fs http.FileSystem) Route {
	return &route{
		routes: r.routes.StaticFS(relativePath, fs),
	}
}

func buildGinHandlers(
	handler HandlerFunc,
	middlewares []HandlerFunc,
) []gin.HandlerFunc {
	if len(middlewares) == 0 {
		return []gin.HandlerFunc{toGinHandler(handler)}
	}

	// Append handler to the last of middlewares, base on gin.Context.Handle docs
	middlewares = append(middlewares, handler)

	out := make([]gin.HandlerFunc, len(middlewares))
	for idx := 0; idx < len(middlewares); idx++ {
		m := middlewares[idx]
		out[idx] = toGinHandler(m)
	}
	return out
}

func toGinHandler(h HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		litCtx := newContext(ctx)
		if err := h(litCtx); err != nil {
			litCtx.Abort() // To skip all middleware after
			litCtx.Error(err)
		}
	}
}
