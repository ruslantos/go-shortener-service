package errors

import (
	"errors"
)

type ClientError struct {
	err error
}

func NewClientError(err error) ClientError {
	return ClientError{
		err: err,
	}
}

func (e ClientError) Error() string {
	return e.err.Error()
}

func IsClientError(err error) bool {
	return errors.As(err, &ClientError{})
}
