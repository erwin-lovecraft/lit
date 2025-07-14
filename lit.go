package lit

import (
	"net/http"
)

// HandlerFunc defines a function to serve HTTP requests or process middleware
type HandlerFunc func(Context) error

// Router controls handler registration
type Router interface {
	Route

	// Group creates a new router group, all the routes inside will have common middleware and prefix
	// Note: Route Cannot have another route group (Router) inside Route
	//
	//
	// Usage:
	//
	//	func main() {
	//		r := new(Route)
	//		r.Group("/api/v1", func(v1 Router) {
	//			v1.Get("/ping", pingHandler)
	//			v1.Put("/something/update", putHandler)
	//		}, loggerMiddleware(), notifyMiddleware()).
	//		// Register 2 route as below and shared loggerMiddleware(), and notifyMiddleware()
	//		// /api/v1/ping -> pingHandler
	//		// /api/v1/something/update -> putHandler
	//	}
	Group(prefix string, routerFunc func(Router), middleware ...HandlerFunc) Router

	// Route creates a new router group, all the routes inside will have common middleware and prefix
	// Note: Route Cannot have another route group (Router) inside Route
	//
	//
	// Usage:
	//
	//	func main() {
	//		r := new(Route)
	//		r.Route("/api/v1", loggerMiddleware(), notifyMiddleware()).
	//		  	Get("/ping", pingHandler).
	//		  	Put("/something/update", putHandler)
	//		// Register 2 route as below and shared loggerMiddleware(), and notifyMiddleware()
	//		// /api/v1/ping -> pingHandler
	//		// /api/v1/something/update -> putHandler
	//	}
	Route(prefix string, middleware ...HandlerFunc) Router

	Routes() RoutesInfo

	// Handler returns http standard handler
	Handler() http.Handler
}

// Route defines a route interface
//
// Usage:
//
//	func main() {
//		r := new(Route)
//		r.Use(loggerMiddleware()).
//		  	Get("/ping", pingHandler).
//		  	Put("/something/update", putHandler)
//		// Register 2 route as below and shared loggerMiddleware()
//		// /ping -> pingHandler
//		// /something/update -> putHandler
//	}
type Route interface {
	// Use adds middleware to router
	Use(middlewares ...HandlerFunc)

	// Handle registers a new request handle
	// NOTICE: The last handler should be the real handler,
	// other ones should be middleware
	Handle(method string, relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Match registers a route that matches the specified declared methods
	// To match all methods, use []string{"*"}
	Match(methods []string, relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Get registers a GET request handler
	Get(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Post registers a POST request handler
	Post(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Delete registers a DELETE request handler
	Delete(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Patch registers a PATCH request handler
	Patch(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Put registers a PUT request handler
	Put(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Options register an OPTIONS request handler
	Options(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// Head registers a HEAD request handler
	Head(relativePath string, handler HandlerFunc, middleware ...HandlerFunc) Route

	// StaticFileFS registers a route to serve a single file with custom `http.FileSystem`
	StaticFileFS(relativePath string, filepath string, fs http.FileSystem) Route

	// StaticFS serves files from the given file system root
	// If got not found, return http.NotFound with Router's NotFound handler
	StaticFS(relativePath string, fs http.FileSystem) Route
}

// ResponseWriter copy from gin.ResponseWriter
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}

// RouterOption options to construct router
type RouterOption func(Router)
