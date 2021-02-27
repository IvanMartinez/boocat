package formats

import (
	"context"
	"strings"
)

// Format definition
type Format struct {
	// Name of the format
	Name string
	// Field names and validators
	Fields map[string]Validate
	// Name of the searchable fields
	Searchable map[string]struct{}
}

// Map of available formats
var Formats map[string]Format

// Signature of validation functions
type Validate func(ctx context.Context, value interface{}) bool

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
			if validateFunc, found := f.Fields[name]; found {
				if validateFunc != nil && !validateFunc(ctx, value) {
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
	for field := range fields {
		if _, found := f.Searchable[field]; !found {
			return false
		}
	}
	return true
}
