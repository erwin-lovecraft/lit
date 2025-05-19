package guard

import (
	"errors"

	"github.com/viebiz/lit"
	"github.com/viebiz/lit/iam"
)

func (guard AuthGuard) RolePermissionHandler(handler lit.HandlerFunc, resource string, permissions Action) lit.HandlerFunc {
	return func(c lit.Context) error {
		req := c.Request()
		ctx := req.Context()

		// 1. Get user profile from request context
		profile := iam.GetUserProfileFromContext(ctx)
		if profile.ID() == "" {
			return errForbidden
		}

		// TODO: Multiple role not supported yet, use the first role
		var r string
		for _, role := range profile.GetRoles() {
			r = role
			break
		}

		// 2. Check if profile permitted to do this action
		if err := guard.enforcer.Enforce(r, resource, permissions.String()); err != nil {
			if errors.Is(err, iam.ErrActionIsNotAllowed) {
				return errForbidden
			}

			return err
		}

		return handler(c)
	}
}
