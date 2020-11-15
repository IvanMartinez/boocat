package database

import (
	"context"
	"log"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName      = "strki"
	recsColName = "records"
)

//@TODO: This may not belong to database
type Record struct {
	DbID        string
	FieldValues map[string]string
}

type Form struct {
	Name   string
	Fields []FormField
}

type FormField struct {
	Name        string
	Label       string
	Description string
	Validator   *regexp.Regexp
}

type DB interface {
	Add(ctx context.Context, fields map[string]string) error
	Update(ctx context.Context, record Record) error
	GetAll(ctx context.Context) ([]Record, error)
	Get(ctx context.Context, id string) (*Record, error)
	GetForm(ctx context.Context, id string) (*Form, error)
}

// Database collections
type MongoDB struct {
	client     *mongo.Client
	recordsCol *mongo.Collection
}

// Connect connects to the database and initialies the collections
func Connect(ctx context.Context, dbURI *string) *MongoDB {
	cli, err := mongo.NewClient(options.Client().ApplyURI(*dbURI))
	if err != nil {
		log.Fatal(err)
	}

	err = cli.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db := cli.Database(dbName)
	recordsCol := db.Collection(recsColName)

	return &MongoDB{
		client:     cli,
		recordsCol: recordsCol,
	}
}

// Disconnect disconnects the database
func (db *MongoDB) Disconnect(ctx context.Context) {
	db.client.Disconnect(ctx)
}

// Add adds a new record to the database with the given fields
func (db *MongoDB) Add(ctx context.Context, values map[string]string) error {
	_, err := db.recordsCol.InsertOne(ctx, values)
	return err
}

// Update updates a record in the database
func (db *MongoDB) Update(ctx context.Context, record Record) error {
	objectID, _ := primitive.ObjectIDFromHex(record.DbID)
	_, err := db.recordsCol.ReplaceOne(ctx, bson.M{"_id": objectID},
		record.FieldValues)
	return err
}

// GetAll returns all records from
func (db *MongoDB) GetAll(ctx context.Context) ([]Record, error) {
	cursor, err := db.recordsCol.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	var documents []map[string]string
	if err = cursor.All(context.TODO(), &documents); err != nil {
		return nil, err
	}

	records := documentsToRecords(documents)
	return records, nil
}

// Get returns a record from the database
func (db *MongoDB) Get(ctx context.Context, id string) (*Record, error) {
	var document map[string]string
	objectID, _ := primitive.ObjectIDFromHex(id)
	err := db.recordsCol.FindOne(context.TODO(),
		bson.M{"_id": objectID}).Decode(&document)
	if err != nil {
		return nil, err
	}

	record := documentToRecord(document)
	return &record, nil
}

func (db *MongoDB) GetForm(ctx context.Context, id string) (*Form, error) {
	//@TDOO: This is a mock-up
	nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
	nameField := FormField{
		Name:  "name",
		Label: "Name",
		Description: "Use only (a-z) characters, separate words with " +
			"whitespace, start every word with capital: John Williams",
		Validator: nameRegExp,
	}
	ageRegExp, _ := regexp.Compile("1?[0-9]{1,2}")
	ageField := FormField{
		Name:        "age",
		Label:       "Age",
		Description: "Number between 0 and 199",
		Validator:   ageRegExp,
	}
	genderRegExp, _ := regexp.Compile("M|F|N")
	genderField := FormField{
		Name:        "gender",
		Label:       "Gender",
		Description: "M for male, F for female, or N in any other case",
		Validator:   genderRegExp,
	}

	return &Form{
		Name:   "Person",
		Fields: []FormField{nameField, ageField, genderField},
	}, nil
}

func documentsToRecords(maps []map[string]string) (records []Record) {
	records = make([]Record, 0, len(maps))
	for _, m := range maps {
		records = append(records, documentToRecord(m))
	}
	return records
}

func documentToRecord(m map[string]string) (record Record) {
	record.FieldValues = make(map[string]string)
	for key, value := range m {
		if key == "_id" {
			record.DbID = value
		} else {
			record.FieldValues[key] = value
		}
	}
	return record
}
