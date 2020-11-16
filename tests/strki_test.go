package tests

import (
	"context"
	"html/template"
	"testing"

	"github.com/ivanmartinez/strki"
)

func initializedDB() (db *MockDB) {
	db = NewDB()
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "Haruki Murakami",
		"birthdate": "1949",
	})
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "George Orwell",
		"birthdate": "1903",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name": "Norwegian Wood",
		"year": "1987",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name": "Kafka On The Shore",
		"year": "2002",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name": "Animal Farm",
		"year": "1945",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name": "Nineteen Eighty-Four",
		"year": "1949",
	})
	return db
}

func TestEditNew(t *testing.T) {
	db := initializedDB()

	authorFormat := strki.EditNew(context.TODO(), db, "author", "",
		nil).(strki.TemplateForm)
	checkForm(t, authorFormat, "author", "/save/author", 2)
	ageField := findTemplateField(t, authorFormat.Fields, "birthdate")
	checkFormatField(t, ageField, "birthdate", "Year of birth", "A year", "")

	bookFormat := strki.EditNew(context.TODO(), db, "book", "",
		nil).(strki.TemplateForm)
	checkForm(t, bookFormat, "book", "/save/book", 2)
	nameField := findTemplateField(t, bookFormat.Fields, "year")
	checkFormatField(t, nameField, "year", "Year", "A year", "")

	// @TODO Check validators somehow?
}

func TestSaveNew(t *testing.T) {
	db := initializedDB()
	list := strki.SaveNew(context.TODO(), db, "author", "",
		map[string]string{
			"name":      "Miguel de Cervantes Saavedra",
			"birthdate": "1547",
		}).([]strki.TemplateRecord)

	if len(list) != 3 {
		t.Errorf("expected 3 records but got %v", len(list))
	}
	record := findTemplateRecord(t, list, "/edit/author/"+db.LastID("author"))
	checkRecordValue(t, record, "name", "Miguel de Cervantes Saavedra")
	checkRecordValue(t, record, "birthdate", "1547")
}

func TestEditExisting(t *testing.T) {
	db := initializedDB()
	form := strki.EditExisting(context.TODO(), db, "book", db.LastID("book"),
		nil).(strki.TemplateForm)

	checkForm(t, form, "book", "/save/book/"+db.LastID("book"), 2)
	nameField := findTemplateField(t, form.Fields, "name")
	checkFormatField(t, nameField, "name", "Name", "A-Z,a-z",
		"Nineteen Eighty-Four")
	// @TODO Check validators somehow?
}

func TestSaveExisting(t *testing.T) {
	db := initializedDB()
	records := strki.SaveExisting(context.TODO(), db, "author",
		db.LastID("author"),
		map[string]string{
			"name":      "Simone de Beauvoir",
			"birthdate": "1908",
		}).([]strki.TemplateRecord)

	if len(records) != 2 {
		t.Errorf("expected 2 records but got %v", len(records))
	}
	record := findTemplateRecord(t, records,
		"/edit/author/"+db.LastID("author"))
	checkRecordValue(t, record, "name", "Simone de Beauvoir")
	checkRecordValue(t, record, "birthdate", "1908")
}

func TestList(t *testing.T) {
	db := initializedDB()
	records := strki.List(context.TODO(), db, "book", "",
		nil).([]strki.TemplateRecord)

	if len(records) != 4 {
		t.Errorf("expected 4 records but got %v", len(records))
	}
	record1 := findTemplateRecord(t, records, "/edit/book/book1")
	checkRecordValue(t, record1, "name", "Norwegian Wood")
	record2 := findTemplateRecord(t, records, "/edit/book/book2")
	checkRecordValue(t, record2, "year", "2002")
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

func checkFormatField(t *testing.T, field strki.TemplateField, name, label,
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
