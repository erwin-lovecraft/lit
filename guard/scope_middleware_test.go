package guard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/viebiz/lit"
	"github.com/viebiz/lit/iam"
)

func TestRequiredM2MScopeMiddleware(t *testing.T) {
	tcs := map[string]struct {
		in         []string
		m2mProfile iam.M2MProfile
		expErr     error
	}{
		"success": {
			in:         []string{"weaponry"},
			m2mProfile: iam.NewM2MProfile("imperium|ultra_marine", []string{"squad", "armory", "weaponry"}),
		},
		"error - profile not exists": {
			in:     []string{"armory"},
			expErr: errForbidden,
		},
		"error - missing required scopes": {
			in:         []string{"armory"},
			m2mProfile: iam.NewM2MProfile("imperium|dark_angel", []string{"squad", "relics", "reinforcements"}),
			expErr:     errForbidden,
		},
	}

	for scenario, tc := range tcs {
		tc := tc
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()

			// Given
			_, ctx, _ := lit.NewRouterForTest(httptest.NewRecorder())

			reqCtx := context.Background()
			req := httptest.NewRequestWithContext(reqCtx, http.MethodGet, "/", nil)
			req = req.WithContext(iam.SetM2MProfileInContext(req.Context(), tc.m2mProfile))
			ctx.SetRequest(req)

			// When
			guard := New(nil, nil)
			hdl := guard.RequiredM2MScopeMiddleware(tc.in...)
			err := hdl(ctx)

			// Then
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
			}
		})
	}
}
