package errorutil

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	Error   = "error"
	Message = "message"
)

type CustomError struct {
	ErrorType     error
	OriginalError error
}

func NewCustomError(errorType error, originalError error) *CustomError {
	return &CustomError{
		ErrorType:     errorType,
		OriginalError: originalError,
	}
}

func GetErrorType(err error) error {
	if e, ok := err.(*CustomError); ok {
		return e.ErrorType
	}

	return err
}

func GetOriginalError(err error) error {
	if e, ok := err.(*CustomError); ok {
		return e.OriginalError
	}
	return err
}

func (c *CustomError) Error() string {
	return fmt.Sprintf("%s: %s", c.ErrorType.Error(), c.OriginalError.Error())
}

var (
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
)

func CombineHTTPErrorMessage(httpStatusCode int, err error) string {
	return fmt.Sprintf("%s: %s", http.StatusText(httpStatusCode), err.Error())
}
