package lit

import (
	"net/http/pprof"
)

// WithLivenessEndpoint setup liveness endpoint, that not captured by monitoring
func WithLivenessEndpoint(endpoint string) RouterOption {
	return func(r Router) {
		r.Get(endpoint, func(c Context) error {
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.Header("X-Content-Type-Options", "nosniff")
			return c.String(200, "OK")
		})
	}
}

func WithProfiling() RouterOption {
	return func(r Router) {
		const prefix = "/_/profile"
		r.Route(prefix).
			Get("/", WrapF(pprof.Index)).
			Get("/cmdline", WrapF(pprof.Cmdline)).
			Get("/profile", WrapF(pprof.Profile)).
			Post("/symbol", WrapF(pprof.Symbol)).
			Get("/symbol", WrapF(pprof.Symbol)).
			Get("/trace", WrapF(pprof.Trace)).
			Get("/allocs", WrapH(pprof.Handler("allocs"))).
			Get("/block", WrapH(pprof.Handler("block"))).
			Get("/goroutine", WrapH(pprof.Handler("goroutine"))).
			Get("/heap", WrapH(pprof.Handler("heap"))).
			Get("/mutex", WrapH(pprof.Handler("mutex"))).
			Get("/threadcreate", WrapH(pprof.Handler("threadcreate")))
	}
}
