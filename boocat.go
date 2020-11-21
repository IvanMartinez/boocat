package boocat

import (
	"context"
	"html/template"
	"log"

	"github.com/ivanmartinez/boocat/database"
)

// TemplateForm contains the data to generate a form with a HTML template
type TemplateForm struct {
	Name      string          // ID
	Fields    []TemplateField // Fields
	SubmitURL template.URL    // Submit URL
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

// TemplateRecord contains the data to show a record (author, book...) with
// a HTML template
type TemplateRecord struct {
	URL         string
	FieldValues map[string]string
}

// HTTPURL is the base URL for links and URLs in the HTML templates
// @TODO: Is There a better solution than this global variable?
var HTTPURL string

// EditNew returns the data to generate a HTML form based on the format
func EditNew(ctx context.Context, db database.DB, pFormat, _pRecord string,
	_submittedValues map[string]string) (string, interface{}) {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		log.Printf("couldn't get format \"%v\": %v\n", pFormat, err)
		return "", nil
	}

	tData := TemplateForm{
		Name:      format.Name,
		Fields:    templateFields(format),
		SubmitURL: template.URL(HTTPURL + "/" + pFormat + "/save"),
	}
	return "edit", tData
}

// SaveNew saves (creates) a new record (author, book, etc)
func SaveNew(ctx context.Context, db database.DB, pFormat, _pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		log.Printf("couldn't get format \"%v\": %v\n", pFormat, err)
		return "", nil
	}

	tplFields, validationFailed := validatedFieldsWithValue(format,
		submittedValues)

	if !validationFailed {
		dbID, err := db.AddRecord(ctx, pFormat, submittedValues)
		if err != nil {
			log.Printf("error adding record to database: %v\n", err)
		}

		tplName, tplData := View(ctx, db, pFormat, dbID, nil)
		return tplName, tplData
	} else {
		tData := TemplateForm{
			Name:      format.Name,
			Fields:    tplFields,
			SubmitURL: template.URL(HTTPURL + "/" + pFormat + "/save"),
		}

		return "edit", tData
	}
}

// EditNew returns the data to generate a HTML form based on the format.
// It also returns the values of a record (author, book...) to pre-fill the
// form.
func EditExisting(ctx context.Context, db database.DB, pFormat, pRecord string,
	_submittedValues map[string]string) (string, interface{}) {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		log.Printf("couldn't get format \"%v\": %v\n", pFormat, err)
		return "", nil
	}

	record, err := db.GetRecord(ctx, pFormat, pRecord)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		tplName, tplData := EditNew(ctx, db, pFormat, pRecord, nil)
		return tplName, tplData
	}

	tplFields, _ := validatedFieldsWithValue(format, record.FieldValues)

	tData := TemplateForm{
		Name:   format.Name,
		Fields: tplFields,
		SubmitURL: template.URL(
			HTTPURL + "/" + pFormat + "/" + record.DbID + "/save"),
	}
	return "edit", tData
}

// SaveExisting saves (updates) a existing record (author, book...)
func SaveExisting(ctx context.Context, db database.DB, pFormat, pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		log.Printf("couldn't get format \"%v\": %v\n", pFormat, err)
		return "", nil
	}

	tplFields, validationFailed := validatedFieldsWithValue(format,
		submittedValues)

	if !validationFailed {
		record := database.Record{
			DbID:        pRecord,
			FieldValues: submittedValues,
		}
		if err := db.UpdateRecord(ctx, pFormat, record); err != nil {
			log.Printf("error updating record in database: %v\n", err)
		}

		tplName, tplData := View(ctx, db, pFormat, pRecord, nil)
		return tplName, tplData
	} else {
		tData := TemplateForm{
			Name:   format.Name,
			Fields: tplFields,
			SubmitURL: template.URL(
				HTTPURL + "/" + pFormat + "/" + pRecord + "/save"),
		}

		return "edit", tData
	}
}

// View returns the data to show a record
func View(ctx context.Context, db database.DB, pFormat, pRecord string,
	_submittedValues map[string]string) (string, interface{}) {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		log.Printf("couldn't get format \"%v\": %v\n", pFormat, err)
		return "", nil
	}

	record, err := db.GetRecord(ctx, pFormat, pRecord)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		tplName, tplData := EditNew(ctx, db, pFormat, pRecord, nil)
		return tplName, tplData
	}

	tData := TemplateRecord{
		URL:         HTTPURL + "/" + pFormat + "/" + pRecord,
		FieldValues: labelValues(format, record.FieldValues),
	}
	return "view", tData
}

// List returns the data to generate a HTML list of records (authors, books...)
func List(ctx context.Context, db database.DB, pFormat, _pRecord string,
	_submittedValues map[string]string) (string, interface{}) {

	records, err := db.GetAllRecords(ctx, pFormat)
	if err != nil {
		log.Printf("error getting records from database: %v\n", err)
		return "", nil
	}

	tData := templateRecords(records, HTTPURL+"/"+pFormat+"/")
	return "list", tData
}

// templateFields takes a format and returns a slice of TemplateField
func templateFields(format *database.Format) []TemplateField {

	fieldsWithValue := make([]TemplateField, len(format.Fields),
		len(format.Fields))
	for index, field := range format.Fields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		fieldsWithValue[index].Value = ""
	}
	return fieldsWithValue
}

// validatedFieldsWithValue takes a format and a slice of field values and
// returns a slice of TemplateField with the values and results of validation
func validatedFieldsWithValue(format *database.Format,
	fieldValues map[string]string) (tplFields []TemplateField,
	valFailed bool) {

	tplFields = make([]TemplateField, len(format.Fields),
		len(format.Fields))
	valFailed = false

	for index, field := range format.Fields {
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
func labelValues(format *database.Format,
	fieldValues map[string]string) map[string]string {

	labelValues := make(map[string]string)
	for _, field := range format.Fields {
		labelValues[field.Label] = fieldValues[field.Name]
	}
	return labelValues
}

// templateRecords takes a slice of records (authors, books...) and returns
// a slice of TemplateRecords to generate a HTML list of records
func templateRecords(records []database.Record,
	baseURL string) []TemplateRecord {

	tRecords := make([]TemplateRecord, len(records), len(records))
	for i, record := range records {
		templateRecord := TemplateRecord{
			URL:         baseURL + record.DbID,
			FieldValues: record.FieldValues,
		}
		tRecords[i] = templateRecord
	}

	return tRecords
}
