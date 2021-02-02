package validators

import (
	"context"
	"fmt"
	"regexp"
)

// Validator is an interface to validate values
type Validator interface {
	// Validate validates a value with the validator
	Validate(context.Context, interface{}) bool
}

type Validate func(ctx context.Context, value interface{}) bool

// nilValidator accepts anything
type nilValidator struct {
}

// regExpValidator validates accordingly to a regular expresion
type regExpValidator struct {
	regExp *regexp.Regexp
}

// referenceValidator validates a reference to another record
type referenceValidator struct {
	format string
	//database database.DB
}

// NewNilValidator returns a new nilValidator
func NewNilValidator() Validator {
	return &nilValidator{}
}

// Validate validates a value with the nilValidator
func (val *nilValidator) Validate(_ctx context.Context, _value interface{}) bool {
	return true
}

// NewRegExpValidator returns a new regExpValidator created from the expresion passed as parameter
func NewRegExpValidator(regExpString string) (Validator, error) {
	if regExp, err := regexp.Compile(regExpString); err == nil {
		return &regExpValidator{regExp: regExp}, nil
	} else {
		return nil, err
	}
}

// Validate validates a value with the regExpValidator
func (val *regExpValidator) Validate(_ctx context.Context, value interface{}) bool {
	stringValue := fmt.Sprintf("%v", value)
	return val.regExp.MatchString(stringValue)
}

/*
// NewReferenceValidator returns a referenceValidator to validate references to records of the passed format name
func NewReferenceValidator(db database.DB, fmt string) Validator {
	return &referenceValidator{
		format:   fmt,
		database: db,
	}
}

// Validate validates a value with the referenceValidator
func (val *referenceValidator) Validate(ctx context.Context, value interface{}) bool {
	stringValue := fmt.Sprintf("%v", value)
	_, error := val.database.GetRecord(ctx, val.format, stringValue)
	return error == nil
}
*/
