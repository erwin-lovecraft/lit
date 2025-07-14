package iam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestM2MProfileMethods(t *testing.T) {
	type args struct {
		p           M2MProfile
		expID       string
		expScopes   []string
		hasScope    string
		expHasScope bool
		anyScopes   []string
		expHasAny   bool
	}

	tests := map[string]args{
		"empty": {
			p:           M2MProfile{id: "", scopes: map[string]bool{}},
			expID:       "",
			expScopes:   []string{},
			hasScope:    "foo",
			expHasScope: false,
			anyScopes:   []string{"a", "b"},
			expHasAny:   false,
		},
		"single": {
			p:           M2MProfile{id: "m1", scopes: map[string]bool{"read": true}},
			expID:       "m1",
			expScopes:   []string{"read"},
			hasScope:    "read",
			expHasScope: true,
			anyScopes:   []string{"write", "read"},
			expHasAny:   true,
		},
		"multiple": {
			p:           M2MProfile{id: "m2", scopes: map[string]bool{"a": true, "b": true}},
			expID:       "m2",
			expScopes:   []string{"a", "b"},
			hasScope:    "c",
			expHasScope: false,
			anyScopes:   []string{"x", "y", "b"},
			expHasAny:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			profile := tc.p

			// WHEN
			gotID := profile.ID()
			gotScopes := profile.GetScopes()
			gotHasScope := profile.HasScope(tc.hasScope)
			gotHasAny := profile.HasAnyScope(tc.anyScopes...)

			// THEN
			require.Equal(t, tc.expID, gotID)
			require.ElementsMatch(t, tc.expScopes, gotScopes)
			require.Equal(t, tc.expHasScope, gotHasScope)
			require.Equal(t, tc.expHasAny, gotHasAny)
		})
	}
}

func TestExtractScopeFromClaims(t *testing.T) {
	type args struct {
		claims Claims
		expMap map[string]bool
		expErr error
	}

	tests := map[string]args{
		"missing scope claim": {
			claims: Claims{ExtraClaims: map[string]interface{}{}},
			expMap: nil,
			expErr: ErrMissingRequiredClaim,
		},
		"single scope": {
			claims: Claims{ExtraClaims: map[string]interface{}{scopeClaimKey: "read"}},
			expMap: map[string]bool{"read": true},
			expErr: nil,
		},
		"multiple scopes": {
			claims: Claims{ExtraClaims: map[string]interface{}{scopeClaimKey: "a b c"}},
			expMap: map[string]bool{"a": true, "b": true, "c": true},
			expErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			claims := tc.claims

			// WHEN
			got, err := extractScopeFromClaims(claims)

			// THEN
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expMap, got)
			}
		})
	}
}
