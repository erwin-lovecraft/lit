package cors

import (
	"github.com/gin-contrib/cors"
	"github.com/viebiz/lit"
)

func Middleware(cfg Config) lit.HandlerFunc {
	return func(c lit.Context) error {
		return lit.AdaptGinHandler(
			cors.New(cfg.underlying),
		)(c)
	}
}
