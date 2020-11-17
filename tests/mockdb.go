package tests

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/ivanmartinez/strki/database"
)

type MockDB struct {
	collections map[string]*Collection
}

type Collection struct {
	name         string
	lastIDNumber int
	records      map[string]database.Record
}

func NewDB() (db *MockDB) {
	return &MockDB{
		collections: map[string]*Collection{
			"author": &Collection{
				name:         "author",
				lastIDNumber: 0,
				records:      make(map[string]database.Record, 0),
			},
			"book": &Collection{
				name:         "book",
				lastIDNumber: 0,
				records:      make(map[string]database.Record, 0),
			},
		},
	}
}

func (db *MockDB) AddRecord(ctx context.Context, format string,
	values map[string]string) error {

	col, found := db.collections[format]
	if !found {
		return fmt.Errorf("unknown format %v", format)
	}

	ID := col.nextID()
	col.records[ID] = database.Record{
		DbID:        ID,
		FieldValues: values,
	}

	return nil
}

func (db *MockDB) UpdateRecord(ctx context.Context, format string,
	record database.Record) error {

	col, found := db.collections[format]
	if !found {
		return fmt.Errorf("unknown format %v", format)
	}

	col.records[record.DbID] = record

	return nil
}

func (db *MockDB) GetAllRecords(ctx context.Context,
	format string) ([]database.Record, error) {

	col, found := db.collections[format]
	if !found {
		return nil, fmt.Errorf("unknown format %v", format)
	}

	slice := make([]database.Record, len(col.records), len(col.records))
	i := 0
	for _, record := range col.records {
		slice[i] = record
		i++
	}

	return slice, nil
}

func (db *MockDB) GetRecord(ctx context.Context, format,
	id string) (*database.Record, error) {

	col, found := db.collections[format]
	if !found {
		return nil, fmt.Errorf("unknown format %v", format)
	}

	if record, found := col.records[id]; found {
		return &record, nil
	}

	return nil, nil
}

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

func (db *MockDB) LastID(format string) string {
	col, found := db.collections[format]
	if !found {
		return ""
	}

	return col.name + strconv.Itoa(col.lastIDNumber)
}

func (col *Collection) lastID() string {
	return col.name + strconv.Itoa(col.lastIDNumber)
}

func (col *Collection) nextID() string {
	col.lastIDNumber++
	return col.lastID()
}
