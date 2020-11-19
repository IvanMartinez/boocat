package tests

import (
	"context"
	"fmt"
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
	if err := checkForm(authorForm, "author", "/author/save", 2); err != nil {
		t.Error(err)
	}
	// Check birthdate field
	if err := checkTemplateField(authorForm.Fields, "birthdate",
		"Year of birth", "A year", ""); err != nil {

		t.Error(err)
	}

	// Run EditNew for book format
	tplName, tplData = boocat.EditNew(context.TODO(), db, "book", "",
		nil)
	bookForm := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(bookForm, "book", "/book/save", 2); err != nil {

		t.Error(err)
	}
	// Check year field
	if err := checkTemplateField(bookForm.Fields, "year", "Year", "A year",
		""); err != nil {

		t.Error(err)
	}

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
	record := tplData.(boocat.TemplateRecord)

	// Check the template
	if tplName != "view" {
		t.Errorf("expected template \"view\" but got \"%v\"", tplName)
	}
	// Check the record
	if err := checkTemplateRecord(record,
		"/author/"+db.LastID("author"),
		map[string]string{
			"name":      "Miguel de Cervantes Saavedra",
			"birthdate": "1547",
		}); err != nil {

		t.Error(err)
	}
}

// TestEditExisting tests boocat.EditExisting
func TestEditExisting(t *testing.T) {
	// Initialize the database
	db := initializedDB()

	// Run EditExisting with the last book in the database
	tplName, tplData := boocat.EditExisting(context.TODO(), db, "book",
		db.LastID("book"), nil)
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "book", "/book/"+db.LastID("book")+"/save",
		2); err != nil {

		t.Error(err)
	}
	// Check name field
	if err := checkTemplateField(form.Fields, "name", "Name", "A-Z,a-z",
		"Nineteen Eighty-Four"); err != nil {

		t.Error(err)
	}

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
	record := tplData.(boocat.TemplateRecord)

	// Check the template
	if tplName != "view" {
		t.Errorf("expected template \"view\" but got \"%v\"", tplName)
	}
	// Find the record with the expected URL
	if err := checkTemplateRecord(record,
		"/author/"+db.LastID("author"),
		map[string]string{
			"name":      "Simone de Beauvoir",
			"birthdate": "1908",
		}); err != nil {

		t.Error(err)
	}
}

// TestView tests boocat.View
func TestView(t *testing.T) {
	// Initialize the database
	db := initializedDB()

	// Run View with the last author in the database
	tplName, tplData := boocat.View(context.TODO(), db, "author",
		db.LastID("author"), nil)
	record := tplData.(boocat.TemplateRecord)

	// Check the template
	if tplName != "view" {
		t.Errorf("expected template \"view\" but got \"%v\"", tplName)
	}
	// Check the record
	if err := checkTemplateRecord(record, "/author/author2",
		map[string]string{
			"Name":          "George Orwell",
			"Year of birth": "1903",
		}); err != nil {

		t.Error(err)
	}
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
	if err := findCheckTemplateRecord(records, "/book/book1",
		map[string]string{
			"name": "Norwegian Wood",
			"year": "1987",
		}); err != nil {

		t.Error(err)
	}

	// Find another record with the expected URL
	if err := findCheckTemplateRecord(records, "/book/book2",
		map[string]string{
			"name": "Kafka On The Shore",
			"year": "2002",
		}); err != nil {

		t.Error(err)
	}
}

// checkForm checks the values and number of fields of a TemplateForm
func checkForm(form boocat.TemplateForm, name, url string,
	fields int) error {

	if form.Name != name {
		return fmt.Errorf("unexpected form name \"%v\" should be \"%v\"",
			form.Name, name)
	}
	if form.SubmitURL != template.URL(url) {
		return fmt.Errorf("unexpected form submit URL \"%v\" should be \"%v\"",
			form.SubmitURL, url)
	}
	if len(form.Fields) != fields {
		return fmt.Errorf("unexpected number of fields %v should be %v",
			len(form.Fields), fields)
	}

	return nil
}

// checkTemplateField checks the values of a TemplateField
func checkTemplateField(fields []boocat.TemplateField, name,
	label, description, value string) error {

	for _, field := range fields {
		if field.Name == name {
			if field.Label != label {
				return fmt.Errorf(
					"unexpected field label \"%v\" should be \"%v\"",
					field.Label, label)
			}
			if field.Description != description {
				return fmt.Errorf(
					"unexpected field description \"%v\" should be \"%v\"",
					field.Description, description)
			}
			if field.Value != value {
				return fmt.Errorf(
					"unexpected field value \"%v\" should be \"%v\"",
					field.Value, value)
			}

			return nil
		}
	}

	return fmt.Errorf("couldn't find field \"%v\"", name)
}

// findCheckTemplateRecord looks for a record with the given URL in a slice,
// and then checks the values of its fields
func findCheckTemplateRecord(records []boocat.TemplateRecord, URL string,
	expectedValues map[string]string) error {

	if record := findTemplateRecord(records, URL); record != nil {
		return checkFieldValues(record.FieldValues, expectedValues)
	}

	return fmt.Errorf("couldn't find record with URL \"%v\"", URL)
}

// findTemplateRecord looks for a record with the given URL in a slice
func findTemplateRecord(records []boocat.TemplateRecord,
	URL string) *boocat.TemplateRecord {

	for _, record := range records {
		if record.URL == URL {
			return &record
		}
	}

	return nil
}

// checkTemplateRecord checks the URL and fields of a TemplateRecord
func checkTemplateRecord(record boocat.TemplateRecord, url string,
	expectedValues map[string]string) error {

	if record.URL != url {
		return fmt.Errorf("unexpected URL \"%v\" should be \"%v\"",
			record.URL, url)
	}

	return checkFieldValues(record.FieldValues, expectedValues)
}

// checkFieldValues compares two maps of field-value pairs
func checkFieldValues(values, expectedValues map[string]string) error {
	if len(values) != len(expectedValues) {
		return fmt.Errorf("expected %v fields but got %v", len(expectedValues),
			len(values))
	}

	for name, value := range values {
		if expectedValue, found := expectedValues[name]; found {
			if value != expectedValue {
				return fmt.Errorf(
					"unexpected field \"%v\" value \"%v\" should be \"%v\"",
					name, value, expectedValue)
			}
		} else {
			return fmt.Errorf("unexpected field \"%v\"", name)
		}
	}

	return nil
}
