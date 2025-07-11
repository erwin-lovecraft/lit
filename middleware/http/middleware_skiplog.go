package http

import (
	"github.com/viebiz/lit"
)

func SkipLoggingResponseBodyMiddleware() lit.HandlerFunc {
	return func(c lit.Context) error {
		c.Set(lit.SkipLoggingResponseBodyKey, true)

		// Continue handle request
		c.Next()

		return nil
	}
}
