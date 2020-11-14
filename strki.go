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
	Fields    []FieldWithValue
	SubmitURL template.URL
}

type FieldWithValue struct {
	Name        string
	Label       string
	Description string
	Value       string
}

type TemplateList struct {
	EditURL template.URL
	Records []database.Record
}

// @TODO: Is There a better solution than this global variable?
var HTTPURL string

func EditNew(ctx context.Context, db database.DB, pathID string,
	formValues map[string]string) interface{} {

	form, _ := db.GetForm(ctx, "")

	tData := TemplateForm{
		Name:      form.Name,
		Fields:    fieldsWithValue(form, nil),
		SubmitURL: template.URL(HTTPURL + "/save"),
	}
	return tData
}

func EditExisting(ctx context.Context, db database.DB, pathID string,
	formValues map[string]string) interface{} {

	form, _ := db.GetForm(ctx, "")

	record, err := db.Get(ctx, pathID)
	if err != nil {
		log.Printf("Error getting database record: %v\n", err)
		return EditNew(ctx, db, pathID, formValues)
	}

	tData := TemplateForm{
		Name:      form.Name,
		Fields:    fieldsWithValue(form, record),
		SubmitURL: template.URL(HTTPURL + "/save/" + record.DbID),
	}
	return tData
}

func SaveNew(ctx context.Context, db database.DB, pathID string,
	formValues map[string]string) interface{} {

	// @TODO: Validate values
	if err := db.Add(ctx, formValues); err != nil {
		log.Printf("Error adding record to database: %v\n", err)
	}
	return List(ctx, db, pathID, formValues)
}

func SaveExisting(ctx context.Context, db database.DB, pathID string,
	formValues map[string]string) interface{} {

	record := database.Record{
		DbID:        pathID,
		FieldValues: formValues,
	}
	if err := db.Update(ctx, record); err != nil {
		log.Printf("Error updating record in database: %v\n", err)
	}
	return List(ctx, db, pathID, formValues)
}

func List(ctx context.Context, db database.DB, pathID string,
	fieldValues map[string]string) interface{} {

	records, err := db.GetAll(ctx)
	if err != nil {
		log.Printf("Error getting records from database: %v\n", err)
		return nil
	}

	tData := TemplateList{
		EditURL: template.URL(HTTPURL + "/edit"),
		Records: records,
	}
	return tData
}

func fieldsWithValue(form *database.Form,
	record *database.Record) []FieldWithValue {

	fieldsWithValue := make([]FieldWithValue, len(form.Fields),
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
