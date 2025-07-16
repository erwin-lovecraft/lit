package lit

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/i18n"
	"github.com/viebiz/lit/testutil"
)

func TestLitContext_BindWithValidation(t *testing.T) {
	type complexStruct struct {
		ID               int `form:"id" json:"id" binding:"required"`
		Equal            int `form:"equal" json:"equal" binding:"eq=100"`
		NotEqual         int `form:"notequal" json:"notequal" binding:"ne=0"`
		LessThan         int `form:"lessthan" json:"lessthan" binding:"lt=50"`
		LessThanEqual    int `form:"lessthanequal" json:"lessthanequal" binding:"lte=50"`
		GreaterThan      int `form:"greaterthan" json:"greaterthan" binding:"gt=10"`
		GreaterThanEqual int `form:"greaterthanequal" json:"greaterthanequal" binding:"gte=10"`
		Multi            int `form:"multi" json:"multi" binding:"min=10,max=20,required"`
	}

	tcs := map[string]struct {
		givenContentType string
		givenRequestBody string
		expectedErr      error
	}{
		"success json": {
			givenContentType: "application/json",
			givenRequestBody: `{
                "id":1,
                "equal":100,
                "notequal":5,
                "lessthan":30,
                "lessthanequal":40,
                "greaterthan":20,
                "greaterthanequal":10,
                "multi":15
            }`,
			expectedErr: nil,
		},
		"got error json": {
			givenContentType: "application/json",
			givenRequestBody: `{
                "equal":99,
                "notequal":0,
                "lessthan":60,
                "lessthanequal":60,
                "greaterthan":10,
                "greaterthanequal":5,
                "multi":30
            }`,
			expectedErr: ValidationError{
				"ID":               "The ID field is required",
				"Equal":            "The Equal field must be 100",
				"NotEqual":         "ne",
				"LessThan":         "lt",
				"LessThanEqual":    "lte",
				"GreaterThan":      "gt",
				"GreaterThanEqual": "gte",
				"Multi":            "The Multi field must be at most 20 but got 30",
			},
		},
		"success form": {
			givenContentType: "application/x-www-form-urlencoded",
			givenRequestBody: func() string {
				f := url.Values{}
				f.Add("id", "1")
				f.Add("equal", "100")
				f.Add("notequal", "5")
				f.Add("lessthan", "30")
				f.Add("lessthanequal", "40")
				f.Add("greaterthan", "20")
				f.Add("greaterthanequal", "10")
				f.Add("multi", "20")
				return f.Encode()
			}(),
			expectedErr: nil,
		},
		"got error form": {
			givenContentType: "application/x-www-form-urlencoded",
			givenRequestBody: func() string {
				f := url.Values{}
				f.Add("equal", "9")
				f.Add("notequal", "0")
				f.Add("lessthan", "60")
				f.Add("lessthanequal", "60")
				f.Add("greaterthan", "10")
				f.Add("greaterthanequal", "5")
				f.Add("multi", "3")
				return f.Encode()
			}(),
			expectedErr: ValidationError{
				"ID":               "The ID field is required",
				"Equal":            "The Equal field must be 100",
				"NotEqual":         "ne",
				"LessThan":         "lt",
				"LessThanEqual":    "lte",
				"GreaterThan":      "gt",
				"GreaterThanEqual": "gte",
				"Multi":            "The Multi field must be at least 10 but got 3",
			},
		},
		"got unexpected error": {
			givenContentType: "application/json",
			givenRequestBody: "invalid json",
			expectedErr:      errors.New("invalid character 'i' looking for beginning of value"),
		},
	}

	for scenario, tc := range tcs {
		tc := tc
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			w := httptest.NewRecorder()
			ctx := CreateTestContext(w)
			req := httptest.NewRequest(http.MethodPost, "/dummy", bytes.NewBufferString(tc.givenRequestBody))
			req.Header.Set("Content-Type", tc.givenContentType)
			ctx.SetRequest(req)

			var compObj complexStruct

			// Initialize localization
			langBundle := i18n.Init(context.Background(), i18n.BundleConfig{
				SourcePath: "i18n/testdata",
			})
			lc := langBundle.GetLocalize("en")
			ctx.SetRequestContext(i18n.SetInContext(ctx, lc))

			// WHEN
			err := ctx.Bind(&compObj)

			// THEN
			if tc.expectedErr != nil {
				var validationErr ValidationError
				if errors.As(tc.expectedErr, &validationErr) {
					testutil.Equal(t, tc.expectedErr, err)
					testutil.Equal(t, parseValidateErrorMessage(tc.expectedErr.Error()), parseValidateErrorMessage(err.Error()))
				} else {
					require.EqualError(t, err, tc.expectedErr.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func parseValidateErrorMessage(errMessage string) map[string]string {
	rs := map[string]string{}
	for _, field := range strings.Split(errMessage, "\n") {
		keyValue := strings.Split(field, ":")
		if len(keyValue) != 2 {
			continue
		}
		rs[keyValue[0]] = keyValue[1]
	}
	return rs
}

func TestValidationError_StatusCode(t *testing.T) {
	validateErr := ValidationError{}

	require.Equal(t, http.StatusBadRequest, validateErr.StatusCode())
}

func TestListContext_Bind(t *testing.T) {
	type givenArgs struct {
		Method  string
		URI     string
		Body    []byte
		Headers http.Header
	}
	type objectStruct struct {
		ID   int    `uri:"id" form:"id" json:"id"`
		Name string `uri:"name" form:"name" json:"name"`
		Role string `uri:"role" form:"role" json:"role"`
	}

	tcs := map[string]struct {
		rawPath   string
		given     givenArgs
		expResult objectStruct
		expErr    error
	}{
		"success - mix type": {
			rawPath: "/spacemarines/:id",
			given: givenArgs{
				Method: "PUT",
				URI:    "/spacemarines/1604",
				Body:   []byte(`{"name":"Erwin","role":"primarch"}`),
				Headers: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			expResult: objectStruct{
				ID:   1604,
				Name: "Erwin",
				Role: "primarch",
			},
		},
		"success - uri only": {
			rawPath: "/spacemarines/:id/:name/:role",
			given: givenArgs{
				Method: "GET",
				URI:    "/spacemarines/1/John/admin",
				Body:   nil,
			},
			expResult: objectStruct{ID: 1, Name: "John", Role: "admin"},
		},
		"success - query only": {
			rawPath: "/spacemarines",
			given: givenArgs{
				Method: "GET",
				URI:    "/spacemarines?id=1&name=John&role=admin",
				Body:   nil,
			},
			expResult: objectStruct{ID: 1, Name: "John", Role: "admin"},
		},

		"success - form only": {
			rawPath: "/spacemarines/:id",
			given: givenArgs{
				Method:  "POST",
				URI:     "/spacemarines/42",
				Body:    []byte("name=Luther&role=tech"),
				Headers: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
			},
			expResult: objectStruct{ID: 42, Name: "Luther", Role: "tech"},
		},

		"error - invalid id": {
			rawPath: "/spacemarines/:id",
			given: givenArgs{
				Method: "GET",
				URI:    "/spacemarines/abc",
				Body:   nil,
			},
			expErr: errors.New("strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},

		"error - invalid JSON": {
			rawPath: "/spacemarines/:id",
			given: givenArgs{
				Method:  "PUT",
				URI:     "/spacemarines/7",
				Body:    []byte(`{"name":`),
				Headers: http.Header{"Content-Type": []string{"application/json"}},
			},
			expErr: errors.New("unexpected EOF"),
		},
	}

	for scenario, tc := range tcs {
		tc := tc
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()

			// Given
			recorder := httptest.NewRecorder()
			mockReq := httptest.NewRequest(tc.given.Method, tc.given.URI, bytes.NewBuffer(tc.given.Body))
			if tc.given.Headers != nil {
				mockReq.Header = tc.given.Headers
			}

			var (
				actualResult objectStruct
				actualErr    error
			)

			r, c, handleContext := NewRouterForTest(recorder)
			r.Match([]string{tc.given.Method}, tc.rawPath, func(c Context) error {
				actualErr = c.Bind(&actualResult)
				return nil
			})
			c.SetRequest(mockReq)

			// When
			handleContext()

			// Then
			if tc.expErr == nil {
				require.NoError(t, actualErr)
				require.Equal(t, tc.expResult, actualResult)
			} else {
				require.EqualError(t, tc.expErr, actualErr.Error())
			}
		})
	}
}
