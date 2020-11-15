package tests

import (
	"context"
	"html/template"
	"testing"

	"github.com/ivanmartinez/strki"
)

func initializedDB() (db *MockDB) {
	db = NewDB()
	db.Add(context.TODO(), map[string]string{
		"name":   "Karen",
		"age":    "41",
		"gender": "F",
	})
	db.Add(context.TODO(), map[string]string{
		"name":   "Marcus",
		"age":    "18",
		"gender": "M",
	})
	db.Add(context.TODO(), map[string]string{
		"name":   "Alfredo",
		"age":    "60",
		"gender": "N",
	})
	return db
}

func TestEditNew(t *testing.T) {
	db := initializedDB()
	form := strki.EditNew(context.TODO(), db, "", nil).(strki.TemplateForm)

	checkForm(t, form, "Person", "/save", 3)
	nameField := findTemplateField(t, form.Fields, "name")
	checkFormField(t, nameField, "name", "Name", "A-Z,a-z", "")
	ageField := findTemplateField(t, form.Fields, "age")
	checkFormField(t, ageField, "age", "Age", "0-199", "")
	genderField := findTemplateField(t, form.Fields, "gender")
	checkFormField(t, genderField, "gender", "Gender", "M/F/N", "")
	// @TODO Check validators somehow?
}

func TestSaveNew(t *testing.T) {
	db := initializedDB()
	list := strki.SaveNew(context.TODO(), db, "",
		map[string]string{
			"name":   "Amelie",
			"age":    "34",
			"gender": "F",
		}).([]strki.TemplateRecord)

	if len(list) != 4 {
		t.Errorf("expected 4 records but got %v", len(list))
	}
	record := findTemplateRecord(t, list, "/edit/"+db.LastID())
	checkRecordValue(t, record, "name", "Amelie")
	checkRecordValue(t, record, "age", "34")
	checkRecordValue(t, record, "gender", "F")
}

func TestEditExisting(t *testing.T) {
	db := initializedDB()
	form := strki.EditExisting(context.TODO(), db, db.LastID(),
		nil).(strki.TemplateForm)

	checkForm(t, form, "Person", "/save/"+db.LastID(), 3)
	nameField := findTemplateField(t, form.Fields, "name")
	checkFormField(t, nameField, "name", "Name", "A-Z,a-z", "Alfredo")
	ageField := findTemplateField(t, form.Fields, "age")
	checkFormField(t, ageField, "age", "Age", "0-199", "60")
	genderField := findTemplateField(t, form.Fields, "gender")
	checkFormField(t, genderField, "gender", "Gender", "M/F/N", "N")
	// @TODO Check validators somehow?
}

func TestSaveExisting(t *testing.T) {
	db := initializedDB()
	records := strki.SaveExisting(context.TODO(), db, db.LastID(),
		map[string]string{
			"name":   "Amelie",
			"age":    "34",
			"gender": "F",
		}).([]strki.TemplateRecord)

	if len(records) != 3 {
		t.Errorf("expected 3 records but got %v", len(records))
	}
	record := findTemplateRecord(t, records, "/edit/"+db.LastID())
	checkRecordValue(t, record, "name", "Amelie")
	checkRecordValue(t, record, "age", "34")
	checkRecordValue(t, record, "gender", "F")
}

func TestList(t *testing.T) {
	db := initializedDB()
	records := strki.List(context.TODO(), db, "", nil).([]strki.TemplateRecord)

	if len(records) != 3 {
		t.Errorf("expected 3 records but got %v", len(records))
	}
	record1 := findTemplateRecord(t, records, "/edit/rcrd1")
	checkRecordValue(t, record1, "name", "Karen")
	record2 := findTemplateRecord(t, records, "/edit/rcrd2")
	checkRecordValue(t, record2, "age", "18")
	record3 := findTemplateRecord(t, records, "/edit/rcrd3")
	checkRecordValue(t, record3, "gender", "N")
}

func checkForm(t *testing.T, form strki.TemplateForm, name, url string,
	fields int) {

	if form.Name != name {
		t.Errorf("unexpected form name \"%v\" should be \"%v\"",
			form.Name, name)
	}
	if form.SubmitURL != template.URL(url) {
		t.Errorf("unexpected form submit URL \"%v\" should be \"%v\"",
			form.SubmitURL, url)
	}
	if len(form.Fields) != fields {
		t.Errorf("unexpected number of fields %v should be %v",
			len(form.Fields), fields)
	}
}

func findTemplateField(t *testing.T, fields []strki.TemplateField,
	name string) strki.TemplateField {

	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}

	t.Fatalf("couldn't find field \"%v\"", name)
	return strki.TemplateField{}
}

func checkFormField(t *testing.T, field strki.TemplateField, name, label,
	description, value string) {

	if field.Name != name {
		t.Errorf("unexpected field name \"%v\" should be \"%v\"", field.Name,
			name)
	}
	if field.Label != label {
		t.Errorf("unexpected field label \"%v\" should be \"%v\"", field.Label,
			label)
	}
	if field.Description != description {
		t.Errorf("unexpected field description \"%v\" should be \"%v\"",
			field.Description, description)
	}
	if field.Value != value {
		t.Errorf("unexpected field value \"%v\" should be \"%v\"", field.Value,
			value)
	}
}

func findTemplateRecord(t *testing.T, records []strki.TemplateRecord,
	URL string) strki.TemplateRecord {

	for _, record := range records {
		if record.URL == URL {
			return record
		}
	}

	t.Fatalf("couldn't find record \"%v\"", URL)
	return strki.TemplateRecord{}
}

func checkRecordValue(t *testing.T, record strki.TemplateRecord, field,
	value string) {

	if fieldValue, found := record.FieldValues[field]; !found {
		t.Errorf("field \"%v\" not found", field)
	} else if fieldValue != value {
		t.Errorf("field \"%v\" should be \"%v\" but is \"%v\"",
			field, value, fieldValue)
	}
}
