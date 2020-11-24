package database

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName = "boocat"
)

// Record represents the data of an entity, currently author or book
// @TODO: Remove this type?
type Record struct {
	DbID        string            // ID of the record in the database
	FieldValues map[string]string // Fields and values of the record
}

// DB is the database interface
type DB interface {
	AddRecord(ctx context.Context, format string,
		fields map[string]string) (string, error)
	UpdateRecord(ctx context.Context, format string, record Record) error
	GetAllRecords(ctx context.Context, format string) ([]Record, error)
	GetRecord(ctx context.Context, format, id string) (*Record, error)
}

// MongoDB database
type MongoDB struct {
	// MongoDB client
	client *mongo.Client
	// Map of collections. Every collection contains the records of a format
	// (author, book...)
	collections map[string]*mongo.Collection
}

// Connect connects to the database and initializes the collections
// @TODO: Maybe the initialization of collections should be separated
func Connect(ctx context.Context, dbURI *string,
	formats []string) *MongoDB {

	// Create and connect the client
	cli, err := mongo.NewClient(options.Client().ApplyURI(*dbURI))
	if err != nil {
		log.Fatal(err)
	}
	err = cli.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the collections
	db := cli.Database(dbName)
	collections := make(map[string]*mongo.Collection, len(formats))
	for _, format := range formats {
		collections[format] = db.Collection(format)
	}

	return &MongoDB{
		client:      cli,
		collections: collections,
	}
}

// Disconnect disconnects the database
func (db *MongoDB) Disconnect(ctx context.Context) {
	db.client.Disconnect(ctx)
}

// AddRecord adds a new record (author, book...) with the given field values to
// the database
func (db *MongoDB) AddRecord(ctx context.Context, format string,
	values map[string]string) (string, error) {

	// If there is a collection for the format
	if col, found := db.collections[format]; found {
		// Insert the record
		if res, err := col.InsertOne(ctx, values); err == nil {
			return res.InsertedID.(primitive.ObjectID).Hex(), err
		} else {
			return "", err
		}
	}

	return "", errors.New("format not found")
}

// Update updates a database record (author, book...) with the given
// field values
func (db *MongoDB) UpdateRecord(ctx context.Context, format string,
	record Record) error {

	// If there is a collection for the format
	if col, found := db.collections[format]; found {
		// Get ObjectID as used by MongoDB
		objectID, _ := primitive.ObjectIDFromHex(record.DbID)
		// Replace the record
		_, err := col.ReplaceOne(ctx, bson.M{"_id": objectID},
			record.FieldValues)
		return err
	}

	return errors.New("format not found")
}

// GetAll returns all records of a specific format from the database
func (db *MongoDB) GetAllRecords(ctx context.Context,
	format string) ([]Record, error) {

	// If there is a collection for the format
	if col, found := db.collections[format]; found {
		// Get a cursor to read all the records
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			return nil, err
		}

		// Read all the records
		var documents []map[string]string
		if err = cursor.All(context.TODO(), &documents); err != nil {
			return nil, err
		}

		// Convert to slice of Record
		records := documentsToRecords(documents)
		return records, nil
	}

	return nil, errors.New("format not found")
}

// Get returns a record from the database
func (db *MongoDB) GetRecord(ctx context.Context, format,
	id string) (*Record, error) {

	// If there is a collection for the format
	if col, found := db.collections[format]; found {
		// Get ObjectID as used by MongoDB
		objectID, _ := primitive.ObjectIDFromHex(id)
		// Read the record
		var document map[string]string
		err := col.FindOne(context.TODO(),
			bson.M{"_id": objectID}).Decode(&document)
		if err != nil {
			return nil, err
		}

		// Convert to Record
		record := documentToRecord(document)
		return &record, nil
	}

	return nil, errors.New("format not found")
}

// documentsToRecords converts a slice of MongoDB documents into a slice of
// Record
func documentsToRecords(maps []map[string]string) (records []Record) {
	records = make([]Record, 0, len(maps))
	for _, m := range maps {
		records = append(records, documentToRecord(m))
	}
	return records
}

// documentToRecord converts a MongoDB document into a Record (author, book...)
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
