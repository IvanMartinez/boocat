package formats

import (
	"github.com/ivanmartinez/boocat/validators"
)

var formats map[string]*Format

// Format defines a form to create or update records (authors, books...),
// including the allowed values for the fields. As a consequence, it defines
// the fields and allowed values of the records.
type Format struct {
	// Display name
	Label string
	// Fields of the format
	Fields []Field
	// Fields[:idSliceLimit] is the minimum set of fields whose values diferentiate
	// a record from the rest
	idSliceLimit int
	// validators for the values of the fields
	validators map[string]validators.Validator
}

// Field contains the data to generate a field of a form with
// a HTML template
type Field struct {
	Name             string // ID
	Label            string // Display name
	Description      string // Description
	Value            string // Value
	ValidationFailed bool   // Whether the value failed validation
}

func Initialize() {
	formats = make(map[string]*Format)

	name := Field{
		Name:        "name",
		Label:       "Name",
		Description: "A-Z,a-z",
	}
	birthdate := Field{
		Name:        "birthdate",
		Label:       "Year of birth",
		Description: "A year",
	}
	year := Field{
		Name:        "year",
		Label:       "Year",
		Description: "A year",
	}
	bio := Field{
		Name:        "biography",
		Label:       "Biography",
		Description: "Free text",
	}
	synopsis := Field{
		Name:        "synopsis",
		Label:       "Synopsis",
		Description: "Free text",
	}

	nameValidator, _ := validators.NewRegExpValidator(
		"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	yearValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")

	formats["author"] = &Format{
		Label:        "Author",
		Fields:       []Field{name, birthdate, bio},
		idSliceLimit: 1,
		validators: map[string]validators.Validator{
			"name":      nameValidator,
			"birthdate": yearValidator,
		},
	}

	formats["book"] = &Format{
		Label:        "Book",
		Fields:       []Field{name, year, synopsis},
		idSliceLimit: 2,
		validators: map[string]validators.Validator{
			"name": nameValidator,
			"year": yearValidator,
		},
	}
}

// GetFormat returns a format
func Get(name string) (*Format, bool) {
	format, found := formats[name]
	return format, found
}

// ValidatedFieldsWithValue takes a format and a slice of field values and
// returns a slice of Field with the values and results of validation
func (f *Format) ValidatedFieldsWithValue(
	fieldValues map[string]string) (tplFields []Field,
	valFailed bool) {

	tplFieldsWithValue := make([]Field, len(f.Fields),
		len(f.Fields))
	valFailed = false

	for index, field := range f.Fields {
		tplFieldsWithValue[index].Name = field.Name
		tplFieldsWithValue[index].Label = field.Label
		tplFieldsWithValue[index].Description = field.Description
		tplFieldsWithValue[index].ValidationFailed = false
		// If there is a value for the field
		if value, found := fieldValues[field.Name]; found {
			tplFieldsWithValue[index].Value = value
			// If there is no validator or validation passed
			if validator, found := f.validators[field.Name]; found {
				if !validator.Validate(value) {
					// Validation failed
					tplFieldsWithValue[index].ValidationFailed = true
					valFailed = true
				}
			}
		}
	}

	return tplFieldsWithValue, valFailed
}

// LabelValues takes a slice of field values and returns a map of the field
// labels with the values
func (f *Format) LabelValues(
	fieldValues map[string]string) map[string]string {

	labelValues := make(map[string]string)
	for _, field := range f.Fields {
		labelValues[field.Label] = fieldValues[field.Name]
	}
	return labelValues
}
