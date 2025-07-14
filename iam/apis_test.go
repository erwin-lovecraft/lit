package iam

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/jwt"
	"github.com/viebiz/lit/postgres"
)

func TestNewM2MProfile(t *testing.T) {
	tests := map[string]struct {
		id     string
		scopes []string
		wantID string
		want   map[string]bool
	}{
		"no scopes": {
			id:     "alice",
			scopes: nil,
			wantID: "alice",
			want:   map[string]bool{},
		},
		"with scopes": {
			id:     "bob",
			scopes: []string{"read", "write", "read"},
			wantID: "bob",
			want:   map[string]bool{"read": true, "write": true},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			id := tc.id
			sc := tc.scopes

			// WHEN
			prof := NewM2MProfile(id, sc)

			// THEN
			require.Equal(t, tc.wantID, prof.id)
			require.Equal(t, tc.want, prof.scopes)
		})
	}
}

func TestNewUserProfile(t *testing.T) {
	tests := map[string]struct {
		id          string
		roles       []string
		permissions []string
		wantID      string
		wantR       []string
		wantP       []string
	}{
		"empty": {
			id:          "",
			roles:       nil,
			permissions: nil,
			wantID:      "",
			wantR:       nil,
			wantP:       nil,
		},
		"non-empty": {
			id:          "carol",
			roles:       []string{"admin", "user"},
			permissions: []string{"perm1"},
			wantID:      "carol",
			wantR:       []string{"admin", "user"},
			wantP:       []string{"perm1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			id := tc.id
			r := tc.roles
			p := tc.permissions

			// WHEN
			up := NewUserProfile(id, r, p)

			// THEN
			require.Equal(t, tc.wantID, up.id)
			require.Equal(t, tc.wantR, up.roles)
			require.Equal(t, tc.wantP, up.permissions)
		})
	}
}

func TestExtractM2MProfileFromClaims(t *testing.T) {
	tests := map[string]struct {
		claims    Claims
		expResult M2MProfile
		expErr    error
	}{
		"success": {
			claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: "u1"},
				ExtraClaims: map[string]interface{}{
					"scope": "read write",
				},
			},
			expResult: M2MProfile{
				id:     "u1",
				scopes: map[string]bool{"read": true, "write": true},
			},
		},
		"no scopes": {
			claims: Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "x"}},
			expErr: ErrMissingRequiredClaim,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			cl := tc.claims

			// WHEN
			prof, err := ExtractM2MProfileFromClaims(cl)

			// THEN
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResult, prof)
			}
		})
	}
}

func TestExtractUserProfileFromClaims(t *testing.T) {
	tests := map[string]struct {
		claims    Claims
		expResult UserProfile
		expErr    error
	}{
		"success": {
			claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: "u1"},
				ExtraClaims: map[string]interface{}{
					"roles": "admin,user",
				},
			},
			expResult: UserProfile{
				id:    "u1",
				roles: []string{"admin", "user"},
			},
		},
		"no scopes": {
			claims: Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "x"}},
			expErr: ErrMissingRequiredClaim,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			cl := tc.claims

			// WHEN
			prof, err := ExtractUserProfileFromClaims(cl)

			// THEN
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResult, prof)
			}
		})
	}
}

func TestNewRFC9068Validator_Error(t *testing.T) {
	mockClient := new(mockHTTPClient)
	mockClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusInternalServerError,
	}, errors.New("fail"))

	tests := map[string]struct {
		issuer   string
		audience string
		client   HTTPClient
		wantErr  string
	}{
		"download fail": {
			issuer:   "https://example.com",
			audience: "aud",
			client:   mockClient,
			wantErr:  "fail",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			iss := tc.issuer
			aud := tc.audience
			cli := tc.client

			// WHEN
			v, err := NewRFC9068Validator(iss, aud, cli)

			// THEN
			require.Nil(t, v)
			require.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestNewEnforcer(t *testing.T) {
	tests := map[string]struct {
		cfg     EnforcerConfig
		wantErr bool
	}{
		"success": {
			cfg: EnforcerConfig{
				DBConn: func() postgres.ContextExecutor {
					pool, err := postgres.NewPool(context.Background(), getPGURL(),
						1, 1,
					)
					require.NoError(t, err)
					require.NotNil(t, pool)

					return pool
				}(),
			},
		},
		"nil DBConn": {
			cfg:     EnforcerConfig{DBConn: nil},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			cfg := tc.cfg

			// WHEN
			e, err := NewEnforcer(context.Background(), cfg)

			// THEN
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, e)
			} else {
				require.NoError(t, err)
				require.NotNil(t, e)
			}
		})
	}
}

func getPGURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgresql://lit:lit@localhost:54321/master?sslmode=disable"
}
