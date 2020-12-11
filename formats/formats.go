package formats

//@TODO: Merge formats and validators?

import (
	"context"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/validators"
)

// Map of available formats. A format is a map of field names and validators.
var formats map[string]map[string]validators.Validator

// Initialize initializes the formats
func Initialize(db database.DB) {
	formats = make(map[string]map[string]validators.Validator)

	nameValidator, _ := validators.NewRegExpValidator(
		"^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$")
	yearValidator, _ := validators.NewRegExpValidator("^[1|2][0-9]{3}$")
	authorValidator := validators.NewReferenceValidator(db, "author")

	formats["author"] = map[string]validators.Validator{
		"name":      nameValidator,
		"birthdate": yearValidator,
		"biography": validators.NewNilValidator(),
	}

	formats["book"] = map[string]validators.Validator{
		"name":     nameValidator,
		"year":     yearValidator,
		"author":   authorValidator,
		"synopsis": validators.NewNilValidator(),
	}
}

// Get returns a format
func Get(name string) (format map[string]validators.Validator, found bool) {
	format, found = formats[name]
	return format, found
}

// IncompleteRecord tells if the record doesn't have all the fields of the format
func IncompleteRecord(validators map[string]validators.Validator, record map[string]string) bool {
	for name := range validators {
		if _, found := record[name]; !found {
			return true
		}
	}
	return false
}

// Merge returns a record with the fields and values of pRecord and sRecord that exist in the format plus id.
// If a field exists in both records then the field and value is taken from pRecord.
func Merge(validators map[string]validators.Validator, pRecord, sRecord map[string]string) (mRecord map[string]string) {
	mRecord = make(map[string]string)
	for name := range validators {
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

// Validate takes a record and returns a slice with the names of the fields that failed the validation of the format
func Validate(ctx context.Context, validators map[string]validators.Validator, record map[string]string) (failed []string) {
	failed = make([]string, 0, len(record))
	for name, value := range record {
		if name != "id" {
			if validator, found := validators[name]; found {
				if !validator.Validate(ctx, value) {
					failed = append(failed, name)
				}
			} else {
				failed = append(failed, name)
			}
		}
	}
	return failed
}
