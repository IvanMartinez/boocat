package boocat

// Implements the definition of a format

import (
	"context"
	"fmt"
)

// Format definition
type Format struct {
	// Name of the format
	Name string
	// Field names and validators
	Fields map[string]Validate
	// Names of the searchable fields
	Searchable map[string]struct{}
}

// Signature of validation functions. If validation succeeds, they return the empty string. Otherwise they return a
// human readable explanation of why it failed.
type Validate func(ctx context.Context, value interface{}) string

// Validate takes a record and returns a map with the result of the validation of every field
func (f Format) Validate(ctx context.Context, record map[string]string) map[string]string {
	failed := make(map[string]string, len(record))
	for name, value := range record {
		if name != "id" {
			if validateFunc, found := f.Fields[name]; found {
				if validateFunc != nil {
					if fail := validateFunc(ctx, value); fail != "" {
						failed[name] = fail
					}
				}
			} else {
				failed[name] = fmt.Sprintf("not a field of format '%s'", f.Name)
			}
		}
	}
	return failed
}

// SearchableAre returns if the searchable fields are the same as the ones passed as parameters
func (f Format) SearchableAre(fields map[string]struct{}) bool {
	if len(f.Searchable) != len(fields) {
		return false
	}
	for field := range fields {
		if _, found := f.Searchable[field]; !found {
			return false
		}
	}
	return true
}
