package instrumentgrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/grpcclient/testdata"
	"github.com/viebiz/lit/testutil"
)

func TestExtractFullMethod(t *testing.T) {
	tests := map[string]struct {
		fullMethod  string
		wantService string
		wantMethod  string
	}{
		"valid full method": {
			fullMethod:  "/weather.WeatherService/GetWeatherInfo",
			wantService: "weather.WeatherService",
			wantMethod:  "GetWeatherInfo",
		},
		"missing leading slash": {
			fullMethod:  "weather.WeatherService/GetWeatherInfo",
			wantService: "weather.WeatherService",
			wantMethod:  "GetWeatherInfo",
		},
		"invalid format": {
			fullMethod:  "invalidFormat",
			wantService: "invalidFormat",
			wantMethod:  "",
		},
		"empty method": {
			fullMethod:  "",
			wantService: "",
			wantMethod:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			fullMethod := tt.fullMethod

			// WHEN
			gotService, gotMethod := extractFullMethod(fullMethod)

			// THEN
			require.Equal(t, tt.wantService, gotService)
			require.Equal(t, tt.wantMethod, gotMethod)
		})
	}
}

func TestSerializeProtoMessage(t *testing.T) {
	tests := map[string]struct {
		req        any
		wantOutput string
	}{
		"valid proto message": {
			req:        &testdata.WeatherRequest{Date: "2023-01-15"},
			wantOutput: `{"date":"2023-01-15"}`,
		},
		"empty proto message": {
			req:        &testdata.WeatherRequest{},
			wantOutput: `{}`,
		},
		"non-proto message": {
			req:        "not a proto message",
			wantOutput: "",
		},
		"nil request": {
			req:        nil,
			wantOutput: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			req := tt.req

			// WHEN
			result := serializeProtoMessage(req)

			// THEN
			if tt.wantOutput == "" {
				require.Nil(t, result)
			} else {
				var expMaps map[string]interface{}
				require.NoError(t, json.Unmarshal([]byte(tt.wantOutput), &expMaps))
				var actMaps map[string]interface{}
				require.NoError(t, json.Unmarshal(result, &actMaps))
				testutil.Equal(t, expMaps, actMaps)
			}
		})
	}
}
