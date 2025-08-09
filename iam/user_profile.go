package iam

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

const (
	defaultNamespace   string = "https://lightning.app"
	roleClaimKey              = defaultNamespace + "/roles"
	permissionClaimKey        = defaultNamespace + "/permissions"
	profileClaimKey           = defaultNamespace + "/profile"
)

type UserProfile struct {
	id          string
	roles       []string
	permissions []string
	profile     map[string]any // To store user data
}

func (p UserProfile) ID() string {
	return p.id
}

func (p UserProfile) GetRoles() []string {
	return slices.Clone(p.roles)
}

func (p UserProfile) GetPermission() []string {
	return slices.Clone(p.permissions)
}

func (p UserProfile) GetRoleString() string {
	return strings.Join(p.roles, ",")
}

func (p UserProfile) GetProfile() map[string]any {
	return maps.Clone(p.profile)
}

func extractRolesFromClaims(claims Claims) ([]string, error) {
	rolesClaim, exists := claims.ExtraClaims[roleClaimKey]
	if !exists {
		return nil, ErrMissingRequiredClaim
	}

	switch v := rolesClaim.(type) {
	case string:
		return strings.Split(v, ","), nil
	case []string:
		return v, nil
	case []interface{}:
		rs := make([]string, len(v))
		for idx, item := range v {
			role, ok := item.(string)
			if !ok {
				role = fmt.Sprintf("%v", item)
			}

			rs[idx] = role
		}

		return rs, nil
	default:
		return nil, ErrInvalidToken
	}
}

func extractUserProfileFromClaims(claims Claims) map[string]any {
	pf, exists := claims.ExtraClaims[profileClaimKey]
	if !exists {
		return nil
	}

	pfMap, ok := pf.(map[string]any)
	if !ok {
		return nil
	}

	return pfMap
}
