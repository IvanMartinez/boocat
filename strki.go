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

type FieldWithValue struct {
	Name        string
	Label       string
	Description string
	Value       string
}

// @TODO: Is There a better solution than this global variable?
var HTTPURL string

func EditNew(ctx context.Context, pathID string,
	formValues map[string]string) interface{} {

	form, _ := database.GetForm(ctx, "")

	type templateData struct {
		Name    string
		Fields  []FieldWithValue
		SaveURL template.URL
	}
	tData := templateData{
		Name:    form.Name,
		Fields:  fieldsWithValue(form, nil),
		SaveURL: template.URL(HTTPURL + "/save"),
	}
	return tData
}

func EditExisting(ctx context.Context, pathID string,
	formValues map[string]string) interface{} {

	form, _ := database.GetForm(ctx, "")

	record, err := database.Get(ctx, pathID)
	if err != nil {
		log.Printf("Error getting database record: %v\n", err)
		return EditNew(ctx, pathID, formValues)
	}

	type templateData struct {
		Name    string
		Fields  []FieldWithValue
		SaveURL template.URL
	}
	tData := templateData{
		Name:    form.Name,
		Fields:  fieldsWithValue(form, record),
		SaveURL: template.URL(HTTPURL + "/save/" + record.DbID),
	}
	return tData
}

func SaveNew(ctx context.Context, pathID string,
	formValues map[string]string) interface{} {

	// @TODO: Validate values
	if err := database.Add(ctx, formValues); err != nil {
		log.Printf("Error adding record to database: %v\n", err)
	}
	return List(ctx, pathID, formValues)
}

func SaveExisting(ctx context.Context, pathID string,
	formValues map[string]string) interface{} {

	record := database.Record{
		DbID:   pathID,
		Fields: formValues,
	}
	if err := database.Update(ctx, record); err != nil {
		log.Printf("Error updating record in database: %v\n", err)
	}
	return List(ctx, pathID, formValues)
}

func List(ctx context.Context, pathID string,
	fieldValues map[string]string) interface{} {

	records, err := database.GetAll(ctx)
	if err != nil {
		log.Printf("Error getting records from database: %v\n", err)
		return nil
	}

	type templateData struct {
		EditURL template.URL
		Records []database.Record
	}
	tData := templateData{
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
			fieldsWithValue[index].Value = record.Fields[field.Name]
		}
	}
	return fieldsWithValue
}
