package lit

import (
	"fmt"
)

// Error represents standard error of lit framework
type Error interface {
	error

	StatusCode() int // Suppose return status code
}

// HTTPError represents an expected error from HTTP request
type HTTPError struct {
	Status int    `json:"-"`
	Code   string `json:"error"`
	Desc   string `json:"error_description"`
}

func (e HTTPError) StatusCode() int {
	return e.Status
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("Status: [%d], Code: [%s], Desc: [%s]", e.Status, e.Code, e.Desc)
}
