package tests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ivanmartinez/boocat/server/formats"
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
	records map[string]map[string]string
}

// NewDB returns a new MockDB with sets for author and book records
func NewDB() (db *MockDB) {
	return &MockDB{
		recordSets: map[string]*RecordSet{
			"author": {
				name:         "author",
				lastIDNumber: 0,
				records:      make(map[string]map[string]string, 0),
			},
			"book": {
				name:         "book",
				lastIDNumber: 0,
				records:      make(map[string]map[string]string, 0),
			},
		},
	}
}

// AddRecord adds a new record (author, book...)
func (db *MockDB) AddRecord(_ context.Context, format string, record map[string]string) (string, error) {
	if rSet, found := db.recordSets[format]; found {
		if _, found := record["id"]; found {
			return "", errors.New("new record cannot have id")
		}
		id := rSet.nextID()
		record["id"] = id
		rSet.records[id] = record
		return id, nil
	}
	return "", errors.New("format not found")
}

// UpdateRecord updates a record (author, book...)
func (db *MockDB) UpdateRecord(_ context.Context, format string, record map[string]string) error {
	if rSet, found := db.recordSets[format]; found {
		if id, found := record["id"]; found {
			rSet.records[id] = record
			return nil
		}
		return errors.New("record doesn't have id")
	}
	return errors.New("format not found")
}

// GetRecord returns a record from the database by the id field
func (db *MockDB) GetRecord(_ context.Context, format, id string) (map[string]string, error) {
	if rSet, found := db.recordSets[format]; found {
		if record, found := rSet.records[id]; found {
			return record, nil
		}
		return nil, fmt.Errorf("unknown record %v", id)
	}
	return nil, errors.New("format not found")
}

// GetAllRecords returns all records of a specific format from the database
func (db *MockDB) GetAllRecords(_ context.Context, format string) ([]map[string]string, error) {
	if rSet, found := db.recordSets[format]; found {
		// Convert from map of records to slice of records
		slice := make([]map[string]string, len(rSet.records), len(rSet.records))
		i := 0
		for _, record := range rSet.records {
			slice[i] = record
			i++
		}
		return slice, nil
	}
	return nil, errors.New("format not found")
}

// SearchRecord returns all records of a specific format from the database that contains the search term in the value of
// their fields
func (db *MockDB) SearchRecord(_ context.Context, formatName, search string) ([]map[string]string, error) {
	if rSet, found := db.recordSets[formatName]; found {
		// Convert from map of records to slice of records
		slice := make([]map[string]string, 0, len(rSet.records))
		for _, record := range rSet.records {
			if matchesSearch(record, search) {
				slice = append(slice, record)
			}
		}
		return slice, nil
	}
	return nil, errors.New("format not found")
}

// referenceValidator returns a validator for references to records of the passed format name
func (db *MockDB) ReferenceValidator(formatName string) formats.Validate {
	return func(ctx context.Context, value interface{}) bool {
		stringValue := fmt.Sprintf("%v", value)
		_, err := db.GetRecord(ctx, formatName, stringValue)
		return err == nil
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

// matchesSearch returns if the value of any field of the record contains the search term, case-insensitive.
func matchesSearch(record map[string]string, search string) bool {
	for _, value := range record {
		if strings.Contains(strings.ToLower(value), strings.ToLower(search)) {
			return true
		}
	}
	return false
}
