package tests

import (
	"context"
	"math/rand"
	"regexp"

	"github.com/ivanmartinez/strki/database"
)

type MockDB struct {
	records map[string]database.Record
}

func NewDB(records map[string]database.Record) (db *MockDB) {
	return &MockDB{
		records: records,
	}
}

func (db *MockDB) Add(ctx context.Context, values map[string]string) error {
	id := randString(5)
	db.records[id] = database.Record{
		DbID:        id,
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
	nameField := database.Field{
		Name:  "name",
		Label: "Name",
		Description: "Use only (a-z) characters, separate words with " +
			"whitespace, start every word with capital: John Williams",
	}
	ageField := database.Field{
		Name:        "age",
		Label:       "Age",
		Description: "Number between 0 and 199",
	}
	genderField := database.Field{
		Name:        "gender",
		Label:       "Gender",
		Description: "M for male, F for female, or N in any other case",
	}
	nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
	ageRegExp, _ := regexp.Compile("1?[0-9]{1,2}")
	genderRegExp, _ := regexp.Compile("M|F|N")
	validators := map[string]regexp.Regexp{
		"name":   *nameRegExp,
		"age":    *ageRegExp,
		"gender": *genderRegExp,
	}

	return &database.Form{
		Name:       "Person",
		Fields:     []database.Field{nameField, ageField, genderField},
		Validators: validators,
	}, nil
}

func randString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
