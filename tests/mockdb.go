package tests

import (
	"context"
	"regexp"
	"strconv"

	"github.com/ivanmartinez/strki/database"
)

type MockDB struct {
	lastIDNumber int
	records      map[string]database.Record
}

func NewDB() (db *MockDB) {
	records := make(map[string]database.Record, 0)
	return &MockDB{
		lastIDNumber: 0,
		records:      records,
	}
}

func (db *MockDB) LastID() string {
	return "rcrd" + strconv.Itoa(db.lastIDNumber)
}

func (db *MockDB) Add(ctx context.Context, values map[string]string) error {
	ID := db.nextID()
	db.records[ID] = database.Record{
		DbID:        ID,
		FieldValues: values,
	}

	return nil
}

func (db *MockDB) Update(ctx context.Context, record database.Record) error {
	db.records[record.DbID] = record

	return nil
}

func (db *MockDB) GetAll(ctx context.Context) ([]database.Record, error) {
	slice := make([]database.Record, 0)
	for _, record := range db.records {
		slice = append(slice, record)
	}

	return slice, nil
}

func (db *MockDB) Get(ctx context.Context, id string) (*database.Record, error) {
	if record, found := db.records[id]; found {
		return &record, nil
	}

	return nil, nil
}

func (db *MockDB) GetForm(ctx context.Context, id string) (*database.Form, error) {
	//@TDOO: This is a mock-up
	nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
	nameField := database.FormField{
		Name:        "name",
		Label:       "Name",
		Description: "A-Z,a-z",
		Validator:   nameRegExp,
	}
	ageRegExp, _ := regexp.Compile("1?[0-9]{1,2}")
	ageField := database.FormField{
		Name:        "age",
		Label:       "Age",
		Description: "0-199",
		Validator:   ageRegExp,
	}
	genderRegExp, _ := regexp.Compile("M|F|N")
	genderField := database.FormField{
		Name:        "gender",
		Label:       "Gender",
		Description: "M/F/N",
		Validator:   genderRegExp,
	}

	return &database.Form{
		Name:   "Person",
		Fields: []database.FormField{nameField, ageField, genderField},
	}, nil
}

func (db *MockDB) nextID() string {
	db.lastIDNumber++
	return db.LastID()
}
