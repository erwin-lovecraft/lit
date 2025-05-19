package guard

import (
	"github.com/viebiz/lit"
	"github.com/viebiz/lit/iam"
)

func (guard AuthGuard) RequiredM2MScopeMiddleware(scopes ...string) lit.HandlerFunc {
	return func(c lit.Context) error {
		ctx := c.Request().Context()

		// 1. Get M2M profile from request context
		profile := iam.GetM2MProfileFromContext(ctx)
		if profile.ID() == "" {
			return errForbidden
		}

		// 2. Check if profile has any required scopes
		if !profile.HasAnyScope(scopes...) {
			return errForbidden
		}

		// 3. Continue with the next handler
		c.Next()
		
		return nil
	}
}
