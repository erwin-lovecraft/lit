package monitoring

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_initSentry(t *testing.T) {
	type args struct {
		givenCfg    sentryConfig
		clientExist bool
		expErr      error
	}
	tcs := map[string]args{
		"success": {
			givenCfg: sentryConfig{
				DSN:         "https://whatever@example.com/16042000",
				ServerName:  "lightning",
				Environment: "dev",
				Version:     "1.0.0",
			},
			clientExist: true,
		},
		"success - skip if DSN empty": {
			givenCfg: sentryConfig{},
		},
	}

	for scenario, tc := range tcs {
		tc := tc
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			// Given

			// When
			c, err := initSentry(tc.givenCfg, &Monitor{logger: zap.NewNop()})

			// Then
			if tc.expErr != nil {
				require.ErrorContains(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
			}
			if tc.clientExist {
				require.NotNil(t, c)
			}
		})
	}
}
