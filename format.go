package boocat

import (
	"context"
	"errors"

	"github.com/ivanmartinez/boocat/validators"
)

var formats map[string]format

// Format defines a form to create or update records (authors, books...),
// including the allowed values for the fields. As a consequence, it defines
// the fields and allowed values of the records.
type format struct {
	// ID
	// @TODO: Re-think this field, ID or label?
	Name string
	// Fields of the format that identify a record. This should be the minimum
	// set of fields whose values differentiate a record from the rest.
	fields []formatField
	// Other fields of the format.
	otherFields []formatField
}

// formatField defines defines a form field, together with the
// allowed values
type formatField struct {
	Name        string               // ID
	Label       string               // Display name
	Description string               // Text description
	Validator   validators.Validator // Regular expression to validate the values
}

// getFormat returns a format
func getFormat(ctx context.Context, id string) (*format, error) {
	if id == "author" {
		nameValidator, _ := validators.NewRegExpValidator(
			"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
		nameField := formatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameValidator,
		}
		birthdateValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")
		birthdateField := formatField{
			Name:        "birthdate",
			Label:       "Year of birth",
			Description: "A year",
			Validator:   birthdateValidator,
		}

		return &format{
			Name:   "author",
			fields: []formatField{nameField, birthdateField},
		}, nil

	} else if id == "book" {
		nameValidator, _ := validators.NewRegExpValidator(
			"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
		nameField := formatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameValidator,
		}
		yearValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")
		yearField := formatField{
			Name:        "year",
			Label:       "Year",
			Description: "A year",
			Validator:   yearValidator,
		}

		return &format{
			Name:   "book",
			fields: []formatField{nameField, yearField},
		}, nil

	} else {
		return nil, errors.New("format not found")
	}
}

// templateFields takes a format and returns a slice of TemplateField
func (f *format) templateFields() []TemplateField {

	fieldsWithValue := make([]TemplateField, len(f.fields),
		len(f.fields))
	for index, field := range f.fields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		fieldsWithValue[index].Value = ""
	}
	return fieldsWithValue
}

// validatedFieldsWithValue takes a format and a slice of field values and
// returns a slice of TemplateField with the values and results of validation
func (f *format) validatedFieldsWithValue(
	fieldValues map[string]string) (tplFields []TemplateField,
	valFailed bool) {

	tplFields = make([]TemplateField, len(f.fields),
		len(f.fields))
	valFailed = false

	for index, field := range f.fields {
		tplFields[index].Name = field.Name
		tplFields[index].Label = field.Label
		tplFields[index].Description = field.Description
		// If there is a value for the field
		if value, found := fieldValues[field.Name]; found {
			tplFields[index].Value = value
			// If there is no validator or validation passed
			if field.Validator == nil ||
				field.Validator.Validate(value) {

				tplFields[index].ValidationFailed = false
			} else {
				// Validation failed
				tplFields[index].ValidationFailed = true
				valFailed = true
			}
		}
	}
	return tplFields, valFailed
}

// labelValues takes a format and a slice of field values and returns a map of
// the field labels with the values
func (f *format) labelValues(
	fieldValues map[string]string) map[string]string {

	labelValues := make(map[string]string)
	for _, field := range f.fields {
		labelValues[field.Label] = fieldValues[field.Name]
	}
	return labelValues
}
