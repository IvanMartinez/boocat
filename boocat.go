package boocat

import (
	"context"
	"html/template"
	"log"

	"github.com/ivanmartinez/boocat/database"
	"github.com/ivanmartinez/boocat/formats"
)

// TemplateForm contains the data to generate a form with a HTML template
type TemplateForm struct {
	Name      string          // ID
	Fields    []formats.Field // Fields
	SubmitURL template.URL    // Submit URL
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

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	tData := TemplateForm{
		Name:      format.Label,
		Fields:    format.Fields,
		SubmitURL: template.URL(HTTPURL + "/" + pFormat + "/save"),
	}
	return "edit", tData
}

// SaveNew saves (creates) a new record (author, book, etc)
func SaveNew(ctx context.Context, db database.DB, pFormat, _pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	tplFields, validationFailed := format.ValidatedFieldsWithValue(
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
			Name:      format.Label,
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

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	record, err := db.GetRecord(ctx, pFormat, pRecord)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		tplName, tplData := EditNew(ctx, db, pFormat, pRecord, nil)
		return tplName, tplData
	}

	tplFields, _ := format.ValidatedFieldsWithValue(record.FieldValues)

	tData := TemplateForm{
		Name:   format.Label,
		Fields: tplFields,
		SubmitURL: template.URL(
			HTTPURL + "/" + pFormat + "/" + record.DbID + "/save"),
	}
	return "edit", tData
}

// SaveExisting saves (updates) a existing record (author, book...)
func SaveExisting(ctx context.Context, db database.DB, pFormat, pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	tplFields, validationFailed := format.ValidatedFieldsWithValue(
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
			Name:   format.Label,
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

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
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
		FieldValues: format.LabelValues(record.FieldValues),
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
