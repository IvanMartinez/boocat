package errors

import "errors"

type InternalServerError struct{}

var (
	ErrFormatNotFound     = errors.New("format not found")
	ErrRecordNotFound     = errors.New("record not found")
	ErrRecordHasID        = errors.New("record has ID")
	ErrRecordDoesntHaveID = errors.New("record doesn't have ID")
)

type ValidationFailedError struct {
	Failed map[string]string
}

func (e InternalServerError) Error() string {
	return "internal server error"
}

func (e ValidationFailedError) Error() string {
	return "validation failed"
}
