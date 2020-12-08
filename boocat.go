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
var (
	HTTPURL string
	DB      database.DB
)

func Get(ctx context.Context, pFormat string, params map[string]string) interface{} {
	if id, found := params["id"]; found {
		return getRecord(ctx, pFormat, id)
	}
	return list(ctx, pFormat)
}

func Update(ctx context.Context, pFormat string, parameters map[string]string) interface{} {
	return nil
}

// EditNew returns the data to generate a HTML form based on the format
func EditNew(ctx context.Context, pFormat, _pRecord string,
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
	return "/edit", tData
}

// SaveNew saves (creates) a new record (author, book, etc)
func SaveNew(ctx context.Context, pFormat, _pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	tplFields, validationFailed := format.ValidatedFieldsWithValue(
		ctx, submittedValues)

	if !validationFailed {
		dbID, err := DB.AddRecord(ctx, pFormat, submittedValues)
		if err != nil {
			log.Printf("error adding record to database: %v\n", err)
		}

		tplData := getRecord(ctx, pFormat, dbID)
		return "", tplData
	}

	tData := TemplateForm{
		Name:      format.Label,
		Fields:    tplFields,
		SubmitURL: template.URL(HTTPURL + "/" + pFormat + "/save"),
	}

	return "/edit", tData
}

// EditNew returns the data to generate a HTML form based on the format.
// It also returns the values of a record (author, book...) to pre-fill the
// form.
func EditExisting(ctx context.Context, pFormat, pRecord string,
	_submittedValues map[string]string) (string, interface{}) {

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	record, err := DB.GetRecord(ctx, pFormat, pRecord)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		tplName, tplData := EditNew(ctx, pFormat, pRecord, nil)
		return tplName, tplData
	}

	tplFields, _ := format.ValidatedFieldsWithValue(ctx, record)

	tData := TemplateForm{
		Name:   format.Label,
		Fields: tplFields,
		SubmitURL: template.URL(
			HTTPURL + "/" + pFormat + "/" + record["id"] + "/save"),
	}
	return "/edit", tData
}

// SaveExisting saves (updates) a existing record (author, book...)
func SaveExisting(ctx context.Context, pFormat, pRecord string,
	submittedValues map[string]string) (string, interface{}) {

	format, found := formats.Get(pFormat)
	if !found {
		log.Printf("couldn't find format \"%v\"", pFormat)
		return "", nil
	}

	tplFields, validationFailed := format.ValidatedFieldsWithValue(ctx, submittedValues)

	if !validationFailed {
		if err := DB.UpdateRecord(ctx, pFormat, submittedValues); err != nil {
			log.Printf("error updating record in database: %v\n", err)
		}

		tplData := getRecord(ctx, pFormat, pRecord)
		return "", tplData
	}

	tData := TemplateForm{
		Name:   format.Label,
		Fields: tplFields,
		SubmitURL: template.URL(
			HTTPURL + "/" + pFormat + "/" + pRecord + "/save"),
	}

	return "/edit", tData
}

// getRecord return a record of a format (author, book...)
func getRecord(ctx context.Context, format, id string) map[string]string {
	record, err := DB.GetRecord(ctx, format, id)
	if err != nil {
		log.Printf("error getting database record: %v\n", err)
		//_, tplData := EditNew(ctx, format, id, nil)
		return nil
	}

	return record
}

// list returns a slice of all records of a format (authors, books...)
func list(ctx context.Context, format string) []map[string]string {
	records, err := DB.GetAllRecords(ctx, format)
	if err != nil {
		log.Printf("error getting records from database: %v\n", err)
		return nil
	}
	return records
}

// templateRecords takes a slice of records (authors, books...) and returns
// a slice of TemplateRecords to generate a HTML list of records
func templateRecords(records []map[string]string, baseURL string) []TemplateRecord {
	tRecords := make([]TemplateRecord, len(records), len(records))
	for i, record := range records {
		templateRecord := TemplateRecord{
			URL:         baseURL + record["id"],
			FieldValues: record,
		}
		tRecords[i] = templateRecord
	}

	return tRecords
}
