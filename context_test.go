package lit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	type mockWriter struct {
		useMock  bool
		inStatus int
		outErr   error
	}

	tcs := map[string]struct {
		inErr          error
		mockWriter     mockWriter
		expectedStatus int
		expectedBody   string
	}{
		"common error": {
			inErr:          errors.New("simulated error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func() string {
				buf := bytes.NewBuffer(nil)
				require.NoError(t, json.NewEncoder(buf).Encode(&HTTPError{
					Status: http.StatusInternalServerError,
					Code:   http.StatusText(http.StatusInternalServerError),
					Desc:   "Internal Server Error",
				}))
				return buf.String()
			}(),
		},
		"expected error": {
			inErr: testError{
				Code: http.StatusBadRequest,
				Msg:  "bad request",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func() string {
				buf := bytes.NewBuffer(nil)
				require.NoError(t, json.NewEncoder(buf).Encode(testError{Code: http.StatusBadRequest, Msg: "bad request"}))
				return buf.String()
			}(),
		},
		"error when marshal": {
			inErr:          testErrorMarshal{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "", // Because marshal error
		},
		"error when write": {
			inErr: testError{
				Code: http.StatusBadRequest,
				Msg:  "bad request",
			},
			mockWriter: mockWriter{
				useMock:  true,
				inStatus: http.StatusBadRequest,
				outErr:   errors.New("simulated error"),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			c := CreateTestContext(recorder)
			c.SetRequest(
				httptest.NewRequest(http.MethodGet, "/", nil),
			)

			if tc.mockWriter.useMock {
				mockResponseWriter := NewMockResponseWriter(t)
				mockResponseWriter.On("Header").Return(recorder.Header())
				mockResponseWriter.On("WriteHeader", tc.mockWriter.inStatus).Run(func(args mock.Arguments) {
					recorder.WriteHeader(tc.mockWriter.inStatus)
				})
				mockResponseWriter.EXPECT().WriteHeaderNow()
				mockResponseWriter.EXPECT().Write(mock.AnythingOfType("[]uint8")).Return(0, tc.mockWriter.outErr)

				// Override the response writer
				c.SetWriter(mockResponseWriter)
			}

			// When
			c.Error(tc.inErr)

			// Then
			require.Equal(t, tc.expectedStatus, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			require.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}

type testError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e testError) Error() string {
	return e.Msg
}

func (e testError) StatusCode() int {
	return e.Code
}

type testErrorMarshal struct{}

func (e testErrorMarshal) Error() string {
	return "marshal error"
}

func (e testErrorMarshal) StatusCode() int {
	return http.StatusBadRequest
}

func (e testErrorMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

func TestParam(t *testing.T) {
	tcs := map[string]struct {
		key            string
		paramValue     string
		expectedResult string
	}{
		"param exists": {
			key:            "id",
			paramValue:     "123",
			expectedResult: "123",
		},
		"param does not exist": {
			key:            "missing",
			paramValue:     "",
			expectedResult: "",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/api/v1/test/123", nil))
			handleContext()

			// When
			result := c.Param(tc.key)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestParamWithDefault(t *testing.T) {
	tcs := map[string]struct {
		key            string
		defaultValue   string
		paramValue     string
		expectedResult string
	}{
		"param exists": {
			key:            "id",
			defaultValue:   "default",
			paramValue:     "123",
			expectedResult: "123",
		},
		"param does not exist": {
			key:            "missing",
			defaultValue:   "default",
			paramValue:     "",
			expectedResult: "default",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/api/v1/test/123", nil))
			handleContext()

			// When
			result := c.ParamWithDefault(tc.key, tc.defaultValue)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestParamWithCallback(t *testing.T) {
	tcs := map[string]struct {
		key            string
		callback       func() string
		paramValue     string
		expectedResult string
	}{
		"param exists": {
			key:            "id",
			callback:       func() string { return "callback-value" },
			paramValue:     "123",
			expectedResult: "123",
		},
		"param does not exist": {
			key:            "missing",
			callback:       func() string { return "callback-value" },
			paramValue:     "",
			expectedResult: "callback-value",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/api/v1/test/123", nil))
			handleContext()

			// When
			result := c.ParamWithCallback(tc.key, tc.callback)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestQuery(t *testing.T) {
	tcs := map[string]struct {
		key            string
		query          string
		expectedResult string
	}{
		"query exists": {
			key:            "id",
			query:          "id=123",
			expectedResult: "123",
		},
		"query does not exist": {
			key:            "missing",
			query:          "id=123",
			expectedResult: "",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/?"+tc.query, nil))
			handleContext()

			// When
			result := c.Query(tc.key)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestQueryWithDefault(t *testing.T) {
	tcs := map[string]struct {
		key            string
		defaultValue   string
		query          string
		expectedResult string
	}{
		"query exists": {
			key:            "id",
			defaultValue:   "default",
			query:          "id=123",
			expectedResult: "123",
		},
		"query does not exist": {
			key:            "missing",
			defaultValue:   "default",
			query:          "id=123",
			expectedResult: "default",
		},
		"query exists but empty": {
			key:            "empty",
			defaultValue:   "default",
			query:          "id=123&empty=",
			expectedResult: "default",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/?"+tc.query, nil))
			handleContext()

			// When
			result := c.QueryWithDefault(tc.key, tc.defaultValue)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestQueryWithCallback(t *testing.T) {
	tcs := map[string]struct {
		key            string
		callback       func() string
		query          string
		expectedResult string
	}{
		"query exists": {
			key:            "id",
			callback:       func() string { return "callback-value" },
			query:          "id=123",
			expectedResult: "123",
		},
		"query does not exist": {
			key:            "missing",
			callback:       func() string { return "callback-value" },
			query:          "id=123",
			expectedResult: "callback-value",
		},
		"query exists but empty": {
			key:            "empty",
			callback:       func() string { return "callback-value" },
			query:          "id=123&empty=",
			expectedResult: "callback-value",
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			r, c, handleContext := NewRouterForTest(recorder)

			r.Get("/api/v1/test/:id", func(c Context) error { return nil })
			c.SetRequest(httptest.NewRequest(http.MethodGet, "/?"+tc.query, nil))
			handleContext()

			// When
			result := c.QueryWithCallback(tc.key, tc.callback)

			// Then
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
