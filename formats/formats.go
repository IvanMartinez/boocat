package formats

import (
	"github.com/ivanmartinez/boocat/validators"
)

var formats map[string]*Format

// Format defines a form to create or update records (authors, books...),
// including the allowed values for the fields. As a consequence, it defines
// the fields and allowed values of the records.
type Format struct {
	// ID
	// @TODO: Re-think this field, ID or label?
	name string
	// Fields of the format that identify a record. This should be the minimum
	// set of fields whose values differentiate a record from the rest.
	idFields []formatField
	// Other non-ID fields of the record
	otherFields []formatField
	// idFields + otherFields
	allFields []formatField
	// fields of the record as []TemplateField
	tplFields []TemplateField
}

// formatField defines defines a form field, together with the
// allowed values
type formatField struct {
	Name        string               // ID
	Label       string               // Display name
	Description string               // Text description
	Validator   validators.Validator // Regular expression to validate the values
}

// TemplateField contains the data to generate a field of a form with
// a HTML template
type TemplateField struct {
	Name             string // ID
	Label            string // Display name
	Description      string // Description
	Value            string // Value
	ValidationFailed bool   // Whether the value failed validation
}

func Initialize() {
	formats = make(map[string]*Format)

	nameValidator, _ := validators.NewRegExpValidator(
		"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	yearValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")

	name := formatField{
		Name:        "name",
		Label:       "Name",
		Description: "A-Z,a-z",
		Validator:   nameValidator,
	}
	birthdate := formatField{
		Name:        "birthdate",
		Label:       "Year of birth",
		Description: "A year",
		Validator:   yearValidator,
	}
	year := formatField{
		Name:        "year",
		Label:       "Year",
		Description: "A year",
		Validator:   yearValidator,
	}
	bio := formatField{
		Name:        "biography",
		Label:       "Biography",
		Description: "Free text",
		Validator:   nil,
	}
	synopsis := formatField{
		Name:        "synopsis",
		Label:       "Synopsis",
		Description: "Free text",
		Validator:   nil,
	}

	formats["author"] = &Format{
		name:        "author",
		idFields:    []formatField{name},
		otherFields: []formatField{birthdate, bio},
		allFields:   []formatField{name, birthdate, bio},
	}
	formats["book"] = &Format{
		name:        "book",
		idFields:    []formatField{name, year},
		otherFields: []formatField{synopsis},
		allFields:   []formatField{name, year, synopsis},
	}
}

// GetFormat returns a format
func Get(name string) (*Format, bool) {
	format, found := formats[name]
	return format, found
}

func (f *Format) Name() string {
	return f.name
}

// templateFields takes a format and returns a slice of TemplateField
func (f *Format) TemplateFields() []TemplateField {

	fieldsWithValue := make([]TemplateField, len(f.allFields),
		len(f.allFields))
	for index, field := range f.allFields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		fieldsWithValue[index].Value = ""
	}
	return fieldsWithValue
}

// ValidatedFieldsWithValue takes a format and a slice of field values and
// returns a slice of TemplateField with the values and results of validation
func (f *Format) ValidatedFieldsWithValue(
	fieldValues map[string]string) (tplFields []TemplateField,
	valFailed bool) {

	tplFields = make([]TemplateField, len(f.allFields),
		len(f.allFields))
	valFailed = false

	for index, field := range f.allFields {
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

// LabelValues takes a format and a slice of field values and returns a map of
// the field labels with the values
func (f *Format) LabelValues(
	fieldValues map[string]string) map[string]string {

	labelValues := make(map[string]string)
	for _, field := range f.allFields {
		labelValues[field.Label] = fieldValues[field.Name]
	}
	return labelValues
}
