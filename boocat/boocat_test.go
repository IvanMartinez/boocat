package boocat

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	bcerrors "github.com/ivanmartinez/boocat/boocat/errors"
	"github.com/ivanmartinez/boocat/log"
)

// MockDB is a database mock for testing
type MockDB struct {
	records map[string][]map[string]string
}

// NewDB returns a new MockDB with sets for author and book records
func NewDB() (db *MockDB) {
	return &MockDB{
		records: map[string][]map[string]string{
			"author": {},
			"book":   {},
		},
	}
}

// AddRecord adds a new record of the format
func (db *MockDB) AddRecord(_ context.Context, format string, record map[string]string) (string, error) {
	if _, found := record["id"]; found {
		return "", bcerrors.ErrRecordHasID
	}
	if _, found := db.records[format]; !found {
		return "", bcerrors.ErrFormatNotFound
	}
	id := strconv.Itoa(len(db.records[format]))
	record["id"] = id
	db.records[format] = append(db.records[format], record)
	return id, nil
}

// UpdateRecord updates a record of the format
func (db *MockDB) UpdateRecord(_ context.Context, formatName string, record map[string]string) error {
	if _, found := record["id"]; !found {
		return bcerrors.ErrRecordDoesntHaveID
	}
	slice, found := db.records[formatName]
	if !found {
		return bcerrors.ErrFormatNotFound
	}
	i, err := strconv.Atoi(record["id"])
	if err != nil {
		return fmt.Errorf("couldn't convert id %q to integer: %v", record["id"], err)
	}
	if i < 0 || i >= len(slice) {
		return bcerrors.ErrRecordNotFound
	}
	slice[i] = record
	return nil
}

// GetRecord returns the record of the format with the id
func (db *MockDB) GetRecord(_ context.Context, formatName, id string) (map[string]string, error) {
	slice, found := db.records[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert id %q to integer: %v", id, err)
	}
	if i < 0 || i >= len(slice) {
		return nil, bcerrors.ErrRecordNotFound
	}
	record := slice[i]
	return record, nil
}

// GetAllRecords returns all records of a specific format from the database
func (db *MockDB) GetAllRecords(_ context.Context, formatName string) ([]map[string]string, error) {
	slice, found := db.records[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	return slice, nil
}

// SearchRecord returns all records the format that have the value in their searchable fields, which in this case are
// all fields
func (db *MockDB) SearchRecord(_ context.Context, formatName, value string) ([]map[string]string, error) {
	slice, found := db.records[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	result := make([]map[string]string, 0, len(slice))
	for _, record := range slice {
		if matchesSearch(record, value) {
			result = append(result, record)
		}
	}
	return result, nil
}

// ReferenceValidator returns a validator of references to records of the format
func (db *MockDB) ReferenceValidator(formatName string) Validate {
	return func(ctx context.Context, value interface{}) string {
		stringValue := fmt.Sprintf("%v", value)
		_, err := db.GetRecord(ctx, formatName, stringValue)
		if err != nil {
			return err.Error()
		}
		return ""
	}
}

// matchesSearch returns if the value of any field of the record contains the search term, case-insensitive.
func matchesSearch(record map[string]string, search string) bool {
	for _, value := range record {
		if strings.Contains(strings.ToLower(value), strings.ToLower(search)) {
			return true
		}
	}
	return false
}

// TestGetRecord tests successfully getting a record with GetRecord
func TestGetRecord(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	result, err := bc.GetRecord(context.Background(), "author", db.records["author"][0]["id"])
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, db.records["author"][0]) {
		t.Errorf("unexpected record: %v", result)
	}
}

// TestListRecords tests successfully getting records with ListRecords
func TestListRecords(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	result, err := bc.ListRecords(context.Background(), "book")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, db.records["book"]) {
		t.Errorf("unexpected records: %v", result)
	}
}

// TestSearchRecords tests successfully searching records with SearchRecords
func TestSearchRecords(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	result, err := bc.SearchRecords(context.Background(), "author", "orwell")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 && !reflect.DeepEqual(result[1], db.records["author"][1]) {
		t.Errorf("unexpected record: %v", result)
	}
}

// TestAddRecord tests successfully adding a record with AddRecord
func TestAddRecord(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	result, err := bc.AddRecord(
		context.Background(),
		"book",
		map[string]string{
			"name":     "The Wind-Up Bird Chronicle",
			"year":     "1995",
			"author":   "0",
			"synopsis": "novel",
		})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Check the response to the request
	i, err := strconv.Atoi(result)
	if err != nil {
		t.Errorf("couldn't convert result %q to integer: %v", result, err)
	}
	// Check that the record is stored as expected
	if !reflect.DeepEqual(db.records["book"][i], map[string]string{
		"id":       "4",
		"name":     "The Wind-Up Bird Chronicle",
		"year":     "1995",
		"author":   "0",
		"synopsis": "novel",
	}) {
		t.Errorf("unexpected record: %v", db.records["book"][4])
	}
}

// TestAddRecordValidationFail tests validation fails when attempting to add a record with AddRecord
func TestAddRecordValidationFail(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	booksCount := len(db.records["book"])
	result, err := bc.AddRecord(
		context.Background(),
		"book",
		map[string]string{
			"name":     "the wind-up bird chronicle",
			"year":     "MCMXCV",
			"author":   "3",
			"synopsis": "novel",
		})
	var validationErrors bcerrors.ValidationFailedError
	if errors.As(err, &validationErrors) {
		if !reflect.DeepEqual(validationErrors.Failed, map[string]string{
			"name":   "doesn't match regular expression",
			"year":   "not a valid year number",
			"author": "record not found",
		}) {
			t.Errorf("unexpected validation errors: %v", validationErrors.Failed)
		}
	} else {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("unexpected result: %v", result)
	}
	if len(db.records["book"]) != booksCount {
		t.Errorf("number of books has changed to %v", len(db.records["book"]))
	}
}

// TestUpdateRecord tests successfully updating a record with UpdateRecord
func TestUpdateRecord(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	err := bc.UpdateRecord(
		context.Background(),
		"author",
		map[string]string{
			"id":        "2",
			"name":      "Miguel De Cervantes Saavedra",
			"birthdate": "1547",
			"biography": "Spanish",
		})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Check that the record is stored as expected
	if !reflect.DeepEqual(db.records["author"][2], map[string]string{
		"id":        "2",
		"name":      "Miguel De Cervantes Saavedra",
		"birthdate": "1547",
		"biography": "Spanish",
	}) {
		t.Errorf("unexpected record: %v", db.records["author"][2])
	}
}

// TestUpdateRecordValidationFail tests validation fails when attempting to update a record with UpdateRecord
func TestUpdateRecordValidationFail(t *testing.T) {
	db := initializedDatabase()
	bc := initializedBoocat(db)
	storedRecord := db.records["author"][1]
	err := bc.UpdateRecord(
		context.Background(),
		"author",
		map[string]string{
			"id":        "1",
			"name":      "george orwell",
			"birthdate": "MCMIII",
		})
	var validationErrors bcerrors.ValidationFailedError
	if errors.As(err, &validationErrors) {
		if !reflect.DeepEqual(validationErrors.Failed, map[string]string{
			"name":      "doesn't match regular expression",
			"birthdate": "not a valid year number",
		}) {
			t.Errorf("unexpected validation errors: %v", validationErrors.Failed)
		}
	} else {
		t.Errorf("unexpected error: %v", err)
	}
	// Check that the record is stored as expected
	if !reflect.DeepEqual(db.records["author"][1], storedRecord) {
		t.Errorf("updated record: %v", db.records["author"][1])
	}
}

// initializedDatabase returns a MockDB with data for testing
func initializedDatabase() (db *MockDB) {
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
		"biography": "spanish",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Norwegian Wood",
		"year":     "1987",
		"author":   "0",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Kafka On The Shore",
		"year":     "2002",
		"author":   "0",
		"synopsis": "novel",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Animal Farm",
		"year":     "1945",
		"author":   "1",
		"synopsys": "fable",
	})
	db.AddRecord(context.TODO(), "book", map[string]string{
		"name":     "Nineteen Eighty-Four",
		"year":     "1949",
		"author":   "1",
		"synopsys": "dystopia",
	})
	return db
}

// initializedBoocat returns a boocat API and logic initialized with the database
func initializedBoocat(db database) *Boocat {
	var bc Boocat
	bc.SetFormat("author", Format{
		Name: "author",
		Fields: map[string]Validate{
			"name":      regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$"),
			"birthdate": validateYear,
			"biography": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "biography": {}},
	})
	bc.SetFormat("book", Format{
		Name: "book",
		Fields: map[string]Validate{
			"name":     regExpValidator("^([A-Z][a-z]*)([ |-][A-Z][a-z]*)*$"),
			"year":     validateYear,
			"author":   db.ReferenceValidator("author"),
			"synopsis": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "synopsis": {}},
	})
	bc.SetDatabase(db)
	return &bc
}

// reqExpValidator returns a validator that uses the regular expression passed as argument
func regExpValidator(regExpString string) Validate {
	regExp, err := regexp.Compile(regExpString)
	if err != nil {
		log.Error.Fatal(err)
	}
	return func(_ context.Context, value interface{}) string {
		stringValue := fmt.Sprintf("%v", value)
		if !regExp.MatchString(stringValue) {
			return "doesn't match regular expression"
		}
		return ""
	}
}

// validateYear is a validator that validates a year
func validateYear(_ context.Context, value interface{}) string {
	stringValue := fmt.Sprintf("%v", value)
	year, err := strconv.Atoi(stringValue)
	if err != nil {
		return "not a valid year number"
	}
	if year < 0 {
		return "Invalid year"
	}
	return ""
}
