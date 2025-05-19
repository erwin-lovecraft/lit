package cors

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
)

// Config holds the CORS configuration
type Config struct {
	underlying cors.Config
}

// New initializes and returns a Config with predefined defaults.
//
// Default configurations:
//   - Allowed methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
//   - Allowed headers:
//   - Standard: "Accept", "Origin", "Content-Type", "Content-Length", "Authorization"
//   - OpenTelemetry: "traceparent", "tracestate", "baggage"
//   - Allow credentials: true (supports cookies and authorization headers)
//   - Max age: 300 seconds (caches preflight response for 5 minutes)
func New(origins []string) Config {
	return Config{
		underlying: cors.Config{
			AllowOrigins: origins,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodHead,
				http.MethodOptions,
			},
			AllowHeaders: []string{
				"Accept", "Origin", "Content-Type", "Content-Length", "Authorization", // Basic headers
				"traceparent", "tracestate", "baggage", // OpenTelemetry headers
			},
			ExposeHeaders:    []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		},
	}
}

func (corsCfg *Config) SetAllowMethods(methods ...string) {
	corsCfg.underlying.AllowMethods = methods
}

func (corsCfg *Config) SetAllowHeaders(headers ...string) {
	corsCfg.underlying.AllowHeaders = headers
}

func (corsCfg *Config) SetExposeHeaders(headers ...string) {
	corsCfg.underlying.ExposeHeaders = headers
}

func (corsCfg *Config) DisableCredentials() {
	corsCfg.underlying.AllowCredentials = false
}

func (corsCfg *Config) SetMaxAge(maxAge time.Duration) {
	corsCfg.underlying.MaxAge = maxAge
}
