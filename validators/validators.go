package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ivanmartinez/boocat/database"
)

// Validator is an interface to validate values
type Validator interface {
	// Validate validates a value with the validator
	Validate(context.Context, interface{}) bool
}

// regExpValidator is a implementation of Validator based on a regular
// expresion
type regExpValidator struct {
	regExp *regexp.Regexp
}

// NewRegExpValidator returns a new regExpValidator created from the expresion
// passed as parameter
func NewRegExpValidator(regExpString string) (Validator, error) {
	if regExp, err := regexp.Compile(regExpString); err == nil {
		return &regExpValidator{regExp: regExp}, nil
	} else {
		return nil, err
	}
}

// Validate validates a value with the regExpValidator
func (val *regExpValidator) Validate(_ctx context.Context,
	value interface{}) bool {

	stringValue := fmt.Sprintf("%v", value)
	return val.regExp.MatchString(stringValue)
}

// referenceValidator is a implementation of Validator to validate a reference
// to another record
type referenceValidator struct {
	format   string
	database database.DB
}

// NewReferenceValidator returns a referenceValidator to validate references
// to records of the passed format
func NewReferenceValidator(db database.DB, fmt string) Validator {
	return &referenceValidator{
		format:   fmt,
		database: db,
	}
}

// Validate validates a value with the referenceValidator
func (val *referenceValidator) Validate(ctx context.Context,
	value interface{}) bool {

	stringValue := fmt.Sprintf("%v", value)
	_, error := val.database.GetRecord(ctx, val.format, stringValue)
	return error == nil
}
