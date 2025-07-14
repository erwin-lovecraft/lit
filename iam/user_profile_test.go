package iam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserProfileMethods(t *testing.T) {
	type args struct {
		p             UserProfile
		expID         string
		expRoles      []string
		expPerms      []string
		expRoleString string
	}

	tests := map[string]args{
		"empty": {
			p:             UserProfile{id: "", roles: []string{}, permissions: []string{}},
			expID:         "",
			expRoles:      []string{},
			expPerms:      []string{},
			expRoleString: "",
		},
		"normal": {
			p:             UserProfile{id: "u1", roles: []string{"admin", "user"}, permissions: []string{"read", "write"}},
			expID:         "u1",
			expRoles:      []string{"admin", "user"},
			expPerms:      []string{"read", "write"},
			expRoleString: "admin,user",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			p := tc.p

			// WHEN
			gotID := p.ID()
			gotRoles := p.GetRoles()
			gotPerms := p.GetPermission()
			gotRoleString := p.GetRoleString()

			// THEN
			require.Equal(t, tc.expID, gotID)
			require.ElementsMatch(t, tc.expRoles, gotRoles)
			require.ElementsMatch(t, tc.expPerms, gotPerms)
			require.Equal(t, tc.expRoleString, gotRoleString)
		})
	}
}

func TestExtractRolesFromClaims(t *testing.T) {
	type args struct {
		claims   Claims
		expRoles []string
		expErr   error
	}

	tests := map[string]args{
		"missing roles claim": {
			claims:   Claims{ExtraClaims: map[string]interface{}{}},
			expRoles: nil,
			expErr:   ErrMissingRequiredClaim,
		},
		"roles as comma string": {
			claims:   Claims{ExtraClaims: map[string]interface{}{roleClaimKey: "admin,owner"}},
			expRoles: []string{"admin", "owner"},
			expErr:   nil,
		},
		"roles as []string": {
			claims:   Claims{ExtraClaims: map[string]interface{}{roleClaimKey: []string{"admin", "owner"}}},
			expRoles: []string{"admin", "owner"},
			expErr:   nil,
		},
		"roles as []interface{} with mixed types": {
			claims: Claims{ExtraClaims: map[string]interface{}{
				roleClaimKey: []interface{}{"admin", "owner", 123},
			}},
			expRoles: []string{"admin", "owner", "123"},
			expErr:   nil,
		},
		"roles wrong type": {
			claims:   Claims{ExtraClaims: map[string]interface{}{roleClaimKey: 123}},
			expRoles: nil,
			expErr:   ErrInvalidToken,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			claims := tc.claims

			// WHEN
			got, err := extractRolesFromClaims(claims)

			// THEN
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tc.expRoles, got)
			}
		})
	}
}
