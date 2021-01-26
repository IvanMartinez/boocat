package formats

//@TODO: Merge formats and validators?

import (
	"context"
	"strings"

	"github.com/ivanmartinez/boocat/validators"
)

// Format definition
type Format struct {
	// Name of the format
	Name string
	// Field names and validators
	Fields map[string]validators.Validator
	// Name of the searchable fields
	Searchable map[string]struct{}
}

// Map of available formats. A format is a map of field names and validators.
var Formats map[string]Format

// Initialize initializes the formats
func Initialize() {
	Formats = make(map[string]Format)

	nameValidator, _ := validators.NewRegExpValidator(
		"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	yearValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")

	Formats["author"] = Format{
		Name: "author",
		Fields: map[string]validators.Validator{
			"name":      nameValidator,
			"birthdate": yearValidator,
			"biography": validators.NewNilValidator(),
		},
		Searchable: map[string]struct{}{"name": struct{}{}},
	}

	Formats["book"] = Format{
		Name: "book",
		Fields: map[string]validators.Validator{
			"name":     nameValidator,
			"year":     yearValidator,
			"author":   validators.NewNilValidator(),
			"synopsis": validators.NewNilValidator(),
		},
		Searchable: map[string]struct{}{"name": struct{}{}},
	}
}

// FormatForTemplate returns the format whose name matches the ending of the template name, and a boolean indicating
// if the format was found
func FormatForTemplate(templateName string) (Format, bool) {
	for formatName, format := range Formats {
		if strings.HasSuffix(templateName, formatName) {
			return format, true
		}
	}
	return Format{}, false
}

// IncompleteRecord tells if the record doesn't have all the fields of the format
func (f Format) IncompleteRecord(record map[string]string) bool {
	for name := range f.Fields {
		if _, found := record[name]; !found {
			return true
		}
	}
	return false
}

// Merge returns a record with the fields and values of pRecord and sRecord that exist in the format plus id.
// If a field exists in both records then the field and value is taken from pRecord.
func (f Format) Merge(pRecord, sRecord map[string]string) (mRecord map[string]string) {
	mRecord = make(map[string]string)
	for name := range f.Fields {
		if value, found := pRecord[name]; found {
			mRecord[name] = value
		} else if value, found := sRecord[name]; found {
			mRecord[name] = value
		}
	}
	// id is not in the format
	if value, found := pRecord["id"]; found {
		mRecord["id"] = value
	} else if value, found := sRecord["id"]; found {
		mRecord["id"] = value
	}
	return mRecord
}

// Validate takes a record and returns a map with the fields that failed the validations of the format
func (f Format) Validate(ctx context.Context, record map[string]string) (failed map[string]string) {
	failed = make(map[string]string)
	for name, value := range record {
		if name != "id" {
			if validator, found := f.Fields[name]; found {
				if !validator.Validate(ctx, value) {
					// Underscore value because empty string is empty pipeline in the template
					failed["_"+name+"_fail"] = "_"
				}
			} else {
				// Underscore value because empty string is empty pipeline in the template
				failed["_"+name+"_fail"] = "_"
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
	for field, _ := range fields {
		if _, found := f.Searchable[field]; !found {
			return false
		}
	}
	return true
}
