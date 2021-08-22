package errors

// common errors

import (
	"errors"
	"fmt"
)

type UnexpectedError struct {
	err error
}

var (
	ErrFormatNotFound     = errors.New("format not found")
	ErrRecordNotFound     = errors.New("record not found")
	ErrRecordHasID        = errors.New("record has ID")
	ErrRecordDoesntHaveID = errors.New("record doesn't have ID")
)

type ValidationFailedError struct {
	Failed map[string]string
}

func NewUnexpectedError(err error) UnexpectedError {
	return UnexpectedError{err: err}
}

func (e UnexpectedError) Error() string {
	return fmt.Sprintf("internal error: %v", e.err)
}

func (e UnexpectedError) Unwrap() error {
	return e.err
}

func (e ValidationFailedError) Error() string {
	return "validation failed"
}
