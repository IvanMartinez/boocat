package strki

import (
	"context"
	"html/template"
	"log"

	"github.com/ivanmartinez/strki/database"
)

type RequestParams struct {
	PathID     string
	FormValues map[string]string
}

type TemplateForm struct {
	Name      string
	Fields    []TemplateField
	SubmitURL template.URL
}

type TemplateField struct {
	Name        string
	Label       string
	Description string
	Value       string
}

type TemplateRecord struct {
	URL         string
	FieldValues map[string]string
}

// @TODO: Is There a better solution than this global variable?
var HTTPURL string

// @TODO: Use format
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

// @TODO: Use format
func SaveNew(ctx context.Context, db database.DB, pFormat, _pRecord string,
	submittedValues map[string]string) interface{} {

	// @TODO: Validate values
	if err := db.AddRecord(ctx, pFormat, submittedValues); err != nil {
		log.Printf("error adding record to database: %v\n", err)
	}
	return List(ctx, db, pFormat, "", submittedValues)
}

// @TODO: Use format
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

// @TODO: Use format
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

// @TODO: Use format
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

func fieldsWithValue(form *database.Format,
	record *database.Record) []TemplateField {

	fieldsWithValue := make([]TemplateField, len(form.Fields),
		len(form.Fields))
	for index, field := range form.Fields {
		fieldsWithValue[index].Name = field.Name
		fieldsWithValue[index].Label = field.Label
		fieldsWithValue[index].Description = field.Description
		if record != nil {
			fieldsWithValue[index].Value = record.FieldValues[field.Name]
		}
	}
	return fieldsWithValue
}

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
