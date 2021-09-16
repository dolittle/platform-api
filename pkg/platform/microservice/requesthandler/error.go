package requesthandler

import (
	"fmt"
	"net/http"
)

type Error struct {
	StatusCode int
	Err        error
}

func NewInternalError(err error) *Error {
	return &Error{http.StatusInternalServerError, err}
}
func NewBadRequest(err error) *Error {
	return &Error{http.StatusBadRequest, err}
}
func NewForbidden(err error) *Error {
	return &Error{http.StatusForbidden, err}
}

func (e *Error) Error() string {
	return fmt.Sprintf("status %d: Err %v", e.StatusCode, e.Err)
}
