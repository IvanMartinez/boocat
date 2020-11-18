package tests

import (
	"context"
	"html/template"
	"testing"

	"github.com/ivanmartinez/boocat"
)

// initiaziledDB returns a MockDB with data for testing
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

// TestEditNew tests boocat.EditNew
func TestEditNew(t *testing.T) {
	// Initialize database
	db := initializedDB()

	// Run EditNew for author format
	tplName, tplData := boocat.EditNew(context.TODO(), db, "author", "",
		nil)
	authorForm := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	checkForm(t, authorForm, "author", "/save/author", 2)
	// Get birthdate field from the form
	birthField := findTemplateField(t, authorForm.Fields, "birthdate")
	// Check birthdate field
	checkTemplateField(t, birthField, "birthdate", "Year of birth", "A year",
		"")

	// Run EditNew for book format
	tplName, tplData = boocat.EditNew(context.TODO(), db, "book", "",
		nil)
	bookForm := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	checkForm(t, bookForm, "book", "/save/book", 2)
	// Get name field from the form
	nameField := findTemplateField(t, bookForm.Fields, "year")
	// Check name field
	checkTemplateField(t, nameField, "year", "Year", "A year", "")

	// @TODO Check returned validators?
}

// TestSaveNew tests boocat.SaveNew
func TestSaveNew(t *testing.T) {
	// Intialize database
	db := initializedDB()

	// Run SaveNew with a new author
	tplName, tplData := boocat.SaveNew(context.TODO(), db, "author", "",
		map[string]string{
			"name":      "Miguel de Cervantes Saavedra",
			"birthdate": "1547",
		})
	records := tplData.([]boocat.TemplateRecord)

	// Check the template
	if tplName != "list" {
		t.Errorf("expected template \"list\" but got \"%v\"", tplName)
	}
	// Check the number of records
	if len(records) != 3 {
		t.Errorf("expected 3 records but got %v", len(records))
	}
	// Find the record with the expected URL
	record := findTemplateRecord(t, records, "/edit/author/"+db.LastID("author"))
	// Check the record values
	checkRecordValue(t, record, "name", "Miguel de Cervantes Saavedra")
	checkRecordValue(t, record, "birthdate", "1547")
}

// Test EditExisting tests boocat.EditExisting
func TestEditExisting(t *testing.T) {
	// Initialize the database
	db := initializedDB()

	// Run EditExisting with the last book in the database
	tplName, tplData := boocat.EditExisting(context.TODO(), db, "book", db.LastID("book"),
		nil)
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	checkForm(t, form, "book", "/save/book/"+db.LastID("book"), 2)
	// Get name field from the form
	nameField := findTemplateField(t, form.Fields, "name")
	// Check name field
	checkTemplateField(t, nameField, "name", "Name", "A-Z,a-z",
		"Nineteen Eighty-Four")

	// @TODO Check returned validators somehow?
}

// TestSaveExisting tests boocat.SaveExisting
func TestSaveExisting(t *testing.T) {
	// Initialize database
	db := initializedDB()

	// Run SaveExisting with a new author
	tplName, tplData := boocat.SaveExisting(context.TODO(), db, "author",
		db.LastID("author"),
		map[string]string{
			"name":      "Simone de Beauvoir",
			"birthdate": "1908",
		})
	records := tplData.([]boocat.TemplateRecord)

	// Check the template
	if tplName != "list" {
		t.Errorf("expected template \"list\" but got \"%v\"", tplName)
	}
	// Check the number of records
	if len(records) != 2 {
		t.Errorf("expected 2 records but got %v", len(records))
	}
	// Find the record with the expected URL
	record := findTemplateRecord(t, records,
		"/edit/author/"+db.LastID("author"))
	// Check the record values
	checkRecordValue(t, record, "name", "Simone de Beauvoir")
	checkRecordValue(t, record, "birthdate", "1908")
}

// TestList tests boocat.List
func TestList(t *testing.T) {
	// Initialize database
	db := initializedDB()

	// Run List for book format
	tplName, tplData := boocat.List(context.TODO(), db, "book", "",
		nil)
	records := tplData.([]boocat.TemplateRecord)

	// Check the template
	if tplName != "list" {
		t.Errorf("expected template \"list\" but got \"%v\"", tplName)
	}
	// Check the number of records in the returned list data
	if len(records) != 4 {
		t.Errorf("expected 4 records but got %v", len(records))
	}
	// Find one record with the expected URL
	record1 := findTemplateRecord(t, records, "/edit/book/book1")
	// Check the name field value
	checkRecordValue(t, record1, "name", "Norwegian Wood")
	// Find another record with the expected URL
	record2 := findTemplateRecord(t, records, "/edit/book/book2")
	// Check the year field value
	checkRecordValue(t, record2, "year", "2002")
}

// checkForm checks the values and number of fields of a TemplateForm
func checkForm(t *testing.T, form boocat.TemplateForm, name, url string,
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

// findTemplateField takes a field name and returns a TemplateField from a
// slice with the same value. If the field couldn't be found then it fails and
// stops the test.
func findTemplateField(t *testing.T, fields []boocat.TemplateField,
	name string) boocat.TemplateField {

	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}

	t.Fatalf("couldn't find field \"%v\"", name)
	return boocat.TemplateField{}
}

// checkTemplateField checks the values of a TemplateField
func checkTemplateField(t *testing.T, field boocat.TemplateField, name, label,
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

// findTemplateRecord takes a record URL and returns a TemplateRecord from a
// slice with the same value. If the record couldn't be found then it fails and
// stops the test.
func findTemplateRecord(t *testing.T, records []boocat.TemplateRecord,
	URL string) boocat.TemplateRecord {

	for _, record := range records {
		if record.URL == URL {
			return record
		}
	}

	t.Fatalf("couldn't find record \"%v\"", URL)
	return boocat.TemplateRecord{}
}

// checkRecordValue checks the value of a field of a TemplateRecord
func checkRecordValue(t *testing.T, record boocat.TemplateRecord, field,
	value string) {

	if fieldValue, found := record.FieldValues[field]; !found {
		t.Errorf("field \"%v\" not found", field)
	} else if fieldValue != value {
		t.Errorf("field \"%v\" should be \"%v\" but is \"%v\"",
			field, value, fieldValue)
	}
}
