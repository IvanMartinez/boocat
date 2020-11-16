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

// @TODO: Remove this type?
type Record struct {
	DbID        string
	FieldValues map[string]string
}

type Format struct {
	Name   string
	Fields []FormatField
}

type FormatField struct {
	Name        string
	Label       string
	Description string
	Validator   *regexp.Regexp
}

type DB interface {
	AddRecord(ctx context.Context, format string,
		fields map[string]string) error
	UpdateRecord(ctx context.Context, format string, record Record) error
	GetAllRecords(ctx context.Context, format string) ([]Record, error)
	GetRecord(ctx context.Context, format, id string) (*Record, error)
	GetFormat(ctx context.Context, id string) (*Format, error)
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

// AddRecord adds a new record to the database with the given fields
// @TODO: Use collection accordingly to format
func (db *MongoDB) AddRecord(ctx context.Context, format string,
	values map[string]string) error {

	_, err := db.recordsCol.InsertOne(ctx, values)
	return err
}

// Update updates a record in the database
// @TODO: Use collection accordingly to format
func (db *MongoDB) UpdateRecord(ctx context.Context, format string,
	record Record) error {

	objectID, _ := primitive.ObjectIDFromHex(record.DbID)
	_, err := db.recordsCol.ReplaceOne(ctx, bson.M{"_id": objectID},
		record.FieldValues)
	return err
}

// GetAll returns all records from
// @TODO: Use collection accordingly to format
func (db *MongoDB) GetAllRecords(ctx context.Context,
	format string) ([]Record, error) {

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
// @TODO: Use collection accordingly to format
func (db *MongoDB) GetRecord(ctx context.Context, format,
	id string) (*Record, error) {

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

//@TDOO: This is a mock-up
func (db *MongoDB) GetFormat(ctx context.Context, id string) (*Format, error) {
	if id == "author" {
		nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
		nameField := FormatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameRegExp,
		}
		birthdateRegExp, _ := regexp.Compile("[1|2][0-9]{3}")
		birthdateField := FormatField{
			Name:        "birthdate",
			Label:       "Year of birth",
			Description: "A year",
			Validator:   birthdateRegExp,
		}

		return &Format{
			Name:   "author",
			Fields: []FormatField{nameField, birthdateField},
		}, nil

	} else if id == "book" {
		nameRegExp, _ := regexp.Compile("([A-Z][a-z]* )*([A-Z][a-z]*)")
		nameField := FormatField{
			Name:        "name",
			Label:       "Name",
			Description: "A-Z,a-z",
			Validator:   nameRegExp,
		}
		yearRegExp, _ := regexp.Compile("[1|2][0-9]{3}")
		yearField := FormatField{
			Name:        "year",
			Label:       "Year",
			Description: "A year",
			Validator:   yearRegExp,
		}

		return &Format{
			Name:   "book",
			Fields: []FormatField{nameField, yearField},
		}, nil

	} else {
		return &Format{}, nil
	}
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
