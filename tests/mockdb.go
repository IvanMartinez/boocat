package tests

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/ivanmartinez/boocat/database"
)

// MockDB is a database mock for testing
type MockDB struct {
	recordSets map[string]*RecordSet
}

// RecordSet is a set of records of a specific format (author, book...)
type RecordSet struct {
	// Name of the set
	name string
	// Last number used in the generation of record IDs
	lastIDNumber int
	// Map of records
	records map[string]database.Record
}

// NewDB returns a new MockDB with sets for author and book records
func NewDB() (db *MockDB) {
	return &MockDB{
		recordSets: map[string]*RecordSet{
			"author": &RecordSet{
				name:         "author",
				lastIDNumber: 0,
				records:      make(map[string]database.Record, 0),
			},
			"book": &RecordSet{
				name:         "book",
				lastIDNumber: 0,
				records:      make(map[string]database.Record, 0),
			},
		},
	}
}

// AddRecord adds a new record (author, book...) with the given field values to
// the database
func (db *MockDB) AddRecord(ctx context.Context, format string,
	values map[string]string) error {

	rSet, found := db.recordSets[format]
	if !found {
		return fmt.Errorf("unknown format %v", format)
	}

	ID := rSet.nextID()
	rSet.records[ID] = database.Record{
		DbID:        ID,
		FieldValues: values,
	}

	return nil
}

// Update updates a database record (author, book...) with the given
// field values
func (db *MockDB) UpdateRecord(ctx context.Context, format string,
	record database.Record) error {

	rSet, found := db.recordSets[format]
	if !found {
		return fmt.Errorf("unknown format %v", format)
	}

	rSet.records[record.DbID] = record

	return nil
}

// GetAll returns all records of a specific format from the database
func (db *MockDB) GetAllRecords(ctx context.Context,
	format string) ([]database.Record, error) {

	rSet, found := db.recordSets[format]
	if !found {
		return nil, fmt.Errorf("unknown format %v", format)
	}

	slice := make([]database.Record, len(rSet.records), len(rSet.records))
	i := 0
	for _, record := range rSet.records {
		slice[i] = record
		i++
	}

	return slice, nil
}

// Get returns a record from the database
func (db *MockDB) GetRecord(ctx context.Context, format,
	id string) (*database.Record, error) {

	rSet, found := db.recordSets[format]
	if !found {
		return nil, fmt.Errorf("unknown format %v", format)
	}

	if record, found := rSet.records[id]; found {
		return &record, nil
	}

	return nil, nil
}

// GetFormat returns a format
func (db *MockDB) GetFormat(ctx context.Context,
	id string) (*database.Format, error) {
	//@TDOO: This is a mock-up
	if id == "author" {
		nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
		nameField := database.FormatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameRegExp,
		}
		birthdateRegExp, _ := regexp.Compile("[1|2][0-9]{3}")
		birthdateField := database.FormatField{
			Name:        "birthdate",
			Label:       "Year of birth",
			Description: "A year",
			Validator:   birthdateRegExp,
		}

		return &database.Format{
			Name:   "author",
			Fields: []database.FormatField{nameField, birthdateField},
		}, nil

	} else if id == "book" {
		nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
		nameField := database.FormatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameRegExp,
		}
		yearRegExp, _ := regexp.Compile("[1|2][0-9]{3}")
		yearField := database.FormatField{
			Name:        "year",
			Label:       "Year",
			Description: "A year",
			Validator:   yearRegExp,
		}

		return &database.Format{
			Name:   "book",
			Fields: []database.FormatField{nameField, yearField},
		}, nil

	} else {
		return &database.Format{}, nil
	}
}

// LastID takes a format and returns the database ID of the last record of
// that format inserted in the database. This is used in testing to be able to
// retrieve the record and check its values.
func (db *MockDB) LastID(format string) string {
	rSet, found := db.recordSets[format]
	if !found {
		return ""
	}

	return rSet.name + strconv.Itoa(rSet.lastIDNumber)
}

// lastID returns the database ID of the last record inserted in the set
func (rSet *RecordSet) lastID() string {
	return rSet.name + strconv.Itoa(rSet.lastIDNumber)
}

// nextID generates a new ID to insert a record in a set
func (rSet *RecordSet) nextID() string {
	rSet.lastIDNumber++
	return rSet.lastID()
}
