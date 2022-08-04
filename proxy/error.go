package proxy

import "fmt"

type InvalidRequestError struct {
	reason error
}

func NewInvalidRequestError(reason error) error {
	return &InvalidRequestError{
		reason: reason,
	}
}

func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("invalid request: %s", e.reason.Error())
}

func (e *InvalidRequestError) Unwrap() error {
	return e.reason
}
