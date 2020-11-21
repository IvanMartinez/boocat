package validators

import (
	"fmt"
	"regexp"
)

// Validator is an interface to validate values
type Validator interface {
	// Validate validates a value with the validator
	Validate(interface{}) bool
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

// Validate validates a value with the regExtpValidator
func (val *regExpValidator) Validate(value interface{}) bool {
	stringValue := fmt.Sprintf("%v", value)
	return val.regExp.MatchString(stringValue)
}
