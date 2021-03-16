package errors

import "fmt"

type FormatNotFoundError struct {
	message string
}

type InternalServerError struct{}

type RecordDoesntHaveIDError struct{}

type RecordHasIDError struct{}

type RecordNotFoundError struct {
	message string
}

type ValidationFailedError struct{}

func NewFormatNotFoundError(formatName string) FormatNotFoundError {
	return FormatNotFoundError{message: fmt.Sprintf("format %v not found", formatName)}
}

func NewInternalServerError() InternalServerError {
	return InternalServerError{}
}

func NewRecordDoesntHaveIDError() RecordDoesntHaveIDError {
	return RecordDoesntHaveIDError{}
}

func NewRecordHasIDError() RecordHasIDError {
	return RecordHasIDError{}
}

func NewRecordNotFoundError(formatName, id string) RecordNotFoundError {
	return RecordNotFoundError{message: fmt.Sprintf("record of format %v with ID %v not found", formatName, id)}
}

func NewValidationFailedError() ValidationFailedError {
	return ValidationFailedError{}
}

func (e FormatNotFoundError) Error() string {
	return e.message
}

func (e InternalServerError) Error() string {
	return "internal server error"
}

func (e RecordDoesntHaveIDError) Error() string {
	return "record doesn't have ID"
}

func (e RecordHasIDError) Error() string {
	return "record has ID"
}

func (e RecordNotFoundError) Error() string {
	return "record not found"
}

func (e ValidationFailedError) Error() string {
	return "validation failed"
}
