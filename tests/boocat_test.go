package tests

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"testing"

	"github.com/ivanmartinez/boocat"
	"github.com/ivanmartinez/boocat/formats"
)

// initiaziledDB returns a MockDB with data for testing
func initializedDB() (db *MockDB) {
	db = NewDB()
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "Haruki Murakami",
		"birthdate": "1949",
		"biography": "Japanese",
	})
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "George Orwell",
		"birthdate": "1903",
		"biography": "English",
	})
	db.AddRecord(context.TODO(), "author", map[string]string{
		"name":      "miguel de cervantes saavedra",
		"birthdate": "MDXLVII",
		"biography": "Spanish",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Norwegian Wood",
		"year":     "1987",
		"author":   "author1",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Kafka On The Shore",
		"year":     "2002",
		"author":   "author1",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Animal Farm",
		"year":     "1945",
		"author":   "author2",
		"synopsys": "fable",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":   "Nineteen Eighty-Four",
		"year":   "1949",
		"author": "author2",
	})
	return db
}

// TestEditNew tests boocat.EditNew
func TestEditNew(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run EditNew for author format
	tplName, tplData := boocat.EditNew(context.TODO(), db, "author", "",
		nil)
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "Author", "/author/save", 3); err != nil {
		t.Error(err)
	}
	// Check birthdate field
	if err := checkField(form.Fields, "birthdate",
		"Year of birth", "A year", "", false); err != nil {

		t.Error(err)
	}
}

// TestSaveNew tests boocat.SaveNew when validation succeeds
func TestSaveNew(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run SaveNew with a new book
	tplName, tplData := boocat.SaveNew(context.TODO(), db, "book", "",
		map[string]string{
			"name":     "The Wind-Up Bird Chronicle",
			"year":     "1995",
			"author":   "author1",
			"synopsis": "novel",
		})
	record := tplData.(boocat.TemplateRecord)

	// Check the template
	if tplName != "view" {
		t.Errorf("expected template \"view\" but got \"%v\"", tplName)
	}
	// Check the record
	if err := checkTemplateRecord(record,
		"/book/"+db.LastID("book"),
		map[string]string{
			"Name":     "The Wind-Up Bird Chronicle",
			"Year":     "1995",
			"Synopsis": "novel",
			"Author":   "author1",
		}); err != nil {

		t.Error(err)
	}
}

func TestSaveNewValidationFail(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run SaveNew with a new author
	tplName, tplData := boocat.SaveNew(context.TODO(), db, "author", "",
		map[string]string{
			"name":      "miguel de cervantes saavedra",
			"birthdate": "",
		})
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "Author", "/author/save", 3); err != nil {
		t.Error(err)
	}
	// Check name field
	if err := checkField(form.Fields, "name",
		"Name", "A-Z,a-z", "miguel de cervantes saavedra", true); err != nil {

		t.Error(err)
	}
	// Check birthdate field
	if err := checkField(form.Fields, "birthdate",
		"Year of birth", "A year", "", true); err != nil {

		t.Error(err)
	}
}

// TestEditExisting tests boocat.EditExisting when validation succeeds
func TestEditExisting(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run EditExisting with the last book in the database
	tplName, tplData := boocat.EditExisting(context.TODO(), db, "book",
		"book4", nil)
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "Book", "/book/"+db.LastID("book")+"/save",
		4); err != nil {

		t.Error(err)
	}
	// Check name field
	if err := checkField(form.Fields, "name", "Name", "A-Z,a-z",
		"Nineteen Eighty-Four", false); err != nil {

		t.Error(err)
	}
}

// TestEditExistingValidationFail tests boocat.EditExisting when validation
// fails. This could happen because format validation could change after the record
// was created.
func TestEditExistingValidationFail(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run EditExisting with the last book in the database
	tplName, tplData := boocat.EditExisting(context.TODO(), db, "author",
		db.LastID("author"), nil)
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "Author", "/author/"+db.LastID("author")+"/save",
		3); err != nil {

		t.Error(err)
	}
	// Check name field
	if err := checkField(form.Fields, "name", "Name", "A-Z,a-z",
		"miguel de cervantes saavedra", true); err != nil {

		t.Error(err)
	}
	// Check birthdate field
	if err := checkField(form.Fields, "birthdate", "Year of birth", "A year",
		"MDXLVII", true); err != nil {

		t.Error(err)
	}
}

// TestSaveExisting tests boocat.SaveExisting
func TestSaveExisting(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run SaveExisting with a new author
	tplName, tplData := boocat.SaveExisting(context.TODO(), db, "author",
		db.LastID("author"),
		map[string]string{
			"name":      "Simone De Beauvoir",
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
			"Name":          "Simone De Beauvoir",
			"Year of birth": "1908",
			"Biography":     "",
		}); err != nil {

		t.Error(err)
	}
}

// TestSaveExistingValidationFail tests boocat.SaveExisting when validation
// fails
func TestSaveExistingValidationFail(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run SaveExisting with a new author
	tplName, tplData := boocat.SaveExisting(context.TODO(), db, "book",
		"book1",
		map[string]string{
			"name":   "the road to wigan pier",
			"year":   "MCMXXXVII",
			"author": "noauthor",
		})
	form := tplData.(boocat.TemplateForm)

	// Check the template
	if tplName != "edit" {
		t.Errorf("expected template \"edit\" but got \"%v\"", tplName)
	}
	// Check the form
	if err := checkForm(form, "Book", "/book/book1/save", 4); err != nil {
		t.Error(err)
	}
	// Check name field
	if err := checkField(form.Fields, "name",
		"Name", "A-Z,a-z", "the road to wigan pier", true); err != nil {

		t.Error(err)
	}
	// Check birthdate field
	if err := checkField(form.Fields, "year",
		"Year", "A year", "MCMXXXVII", true); err != nil {

		t.Error(err)
	}
	// Check author field
	if err := checkField(form.Fields, "author",
		"Author", "Writer", "noauthor", true); err != nil {

		t.Error(err)
	}
}

// TestView tests boocat.View
func TestView(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

	// Run View with the last author in the database
	tplName, tplData := boocat.View(context.TODO(), db, "author",
		"author2", nil)

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
			"Biography":     "English",
		}); err != nil {

		t.Error(err)
	}
}

// TestList tests boocat.List
func TestList(t *testing.T) {
	// Initialize formats and database
	db := initializedDB()
	formats.Initialize(db)

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
			"name":     "Norwegian Wood",
			"year":     "1987",
			"synopsis": "novel",
			"author":   "author1",
		}); err != nil {

		t.Error(err)
	}

	// Find another record with the expected URL
	if err := findCheckTemplateRecord(records, "/book/book2",
		map[string]string{
			"name":     "Kafka On The Shore",
			"year":     "2002",
			"synopsis": "novel",
			"author":   "author1",
		}); err != nil {

		t.Error(err)
	}
}

// checkForm checks the values and number of fields of a TemplateForm
func checkForm(form boocat.TemplateForm, name, url string,
	fields int) error {

	if form.Name != name {
		return fmt.Errorf("unexpected form label \"%v\" should be \"%v\"",
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

// checkField checks the values of a Field
func checkField(fields []formats.Field, name,
	label, description, value string, valFail bool) error {

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
			if field.ValidationFailed && !valFail {
				return errors.New(
					"value should not have failed validation")
			}
			if !field.ValidationFailed && valFail {
				return errors.New(
					"value should have failed validation")
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
		return fmt.Errorf("unexpected number of %v fields should be %v",
			len(values), len(expectedValues))
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
