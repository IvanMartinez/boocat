package boocat

import (
	"context"
	"html/template"
	"log"

	"github.com/ivanmartinez/boocat/database"
)

// TemplateForm contains the data to generate a form with a HTML template
type TemplateForm struct {
	Name      string
	Fields    []TemplateField
	SubmitURL template.URL
}

// TemplateField contains the data to generate a field of a form with
// a HTML template
type TemplateField struct {
	Name        string
	Label       string
	Description string
	Value       string
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
	_submittedValues map[string]string) interface{} {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		return nil
	}

	tData := TemplateForm{
		Name:      format.Name,
		Fields:    fieldsWithValue(format, nil),
		SubmitURL: template.URL(HTTPURL + "/save/" + pFormat),
	}
	return tData
}

// SaveNew saves (creates) a new record (author, book, etc)
func SaveNew(ctx context.Context, db database.DB, pFormat, _pRecord string,
	submittedValues map[string]string) interface{} {

	// @TODO: Validate values
	if err := db.AddRecord(ctx, pFormat, submittedValues); err != nil {
		log.Printf("error adding record to database: %v\n", err)
	}
	return List(ctx, db, pFormat, "", submittedValues)
}

// EditNew returns the data to generate a HTML form based on the format.
// It also returns the values of a record (author, book...) to pre-fill the
// form.
func EditExisting(ctx context.Context, db database.DB, pFormat, pRecord string,
	_submittedValues map[string]string) interface{} {

	format, err := db.GetFormat(ctx, pFormat)
	if err != nil {
		return nil
	}

	record, err := db.GetRecord(ctx, pFormat, pRecord)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		return EditNew(ctx, db, pFormat, pRecord, nil)
	}

	tData := TemplateForm{
		Name:   format.Name,
		Fields: fieldsWithValue(format, record),
		SubmitURL: template.URL(
			HTTPURL + "/save/" + pFormat + "/" + record.DbID),
	}
	return tData
}

// SaveNew saves (updates) a existing record (author, book...)
func SaveExisting(ctx context.Context, db database.DB, pFormat, pRecord string,
	submittedValues map[string]string) interface{} {

	record := database.Record{
		DbID:        pRecord,
		FieldValues: submittedValues,
	}
	if err := db.UpdateRecord(ctx, pFormat, record); err != nil {
		log.Printf("error updating record in database: %v\n", err)
	}
	return List(ctx, db, pFormat, "", submittedValues)
}

// List returns the data to generate a HTML list of records (authors, books...)
func List(ctx context.Context, db database.DB, pFormat, _pRecord string,
	_submittedValues map[string]string) interface{} {

	records, err := db.GetAllRecords(ctx, pFormat)
	if err != nil {
		log.Printf("error getting records from database: %v\n", err)
		return nil
	}

	tData := templateRecords(records, HTTPURL+"/edit/"+pFormat+"/")
	return tData
}

// fieldsWithValue takes a format and a returns a slice of TemplateField to
// generate a HTML form. If a record (author, book...) is passed then the
// fields will have the values of the record.
func fieldsWithValue(format *database.Format,
	record *database.Record) []TemplateField {

	fieldsWithValue := make([]TemplateField, len(format.Fields),
		len(format.Fields))
	for index, field := range format.Fields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		if record != nil {
			fieldsWithValue[index].Value = record.FieldValues[field.Name]
		}
	}
	return fieldsWithValue
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
