package guard

import (
	"errors"
	"net/http"

	"github.com/viebiz/lit"
)

const (
	unAuthorizedKey = "unauthorized"
	forbiddenKey    = "forbidden"
)

var (
	errUserProfileNotInCtx = errors.New("user profile not in context")
	errM2MProfileNotInCtx  = errors.New("m2m profile not in context")
	errMissingAccessToken  = &lit.HTTPError{Status: http.StatusUnauthorized, Code: unAuthorizedKey, Desc: "Access token is required"}
	errForbidden           = &lit.HTTPError{Status: http.StatusForbidden, Code: forbiddenKey, Desc: "Permission denied"}
)

func unauthorizedErr(err error) *lit.HTTPError {
	return &lit.HTTPError{Status: http.StatusUnauthorized, Code: unAuthorizedKey, Desc: err.Error()}
}
