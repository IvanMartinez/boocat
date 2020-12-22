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

// DB is the database interface
type DB interface {
	AddRecord(ctx context.Context, format string, record map[string]string) (string, error)
	UpdateRecord(ctx context.Context, format string, record map[string]string) error
	GetAllRecords(ctx context.Context, format string) ([]map[string]string, error)
	GetRecord(ctx context.Context, format, id string) (map[string]string, error)
}

// mongoDB database
type mongoDB struct {
	// MongoDB client
	client *mongo.Client
	// Map of collections. Every collection contains the records of a format
	// (author, book...)
	collections map[string]*mongo.Collection
}

// Connect connects to the database and initializes the collections
// @TODO: Maybe the initialization of collections should be separated
func Connect(ctx context.Context, dbURI *string, formats []string) *mongoDB {

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

	return &mongoDB{
		client:      cli,
		collections: collections,
	}
}

// Disconnect disconnects the database
func (db *mongoDB) Disconnect(ctx context.Context) {
	db.client.Disconnect(ctx)
}

// AddRecord adds a new record (author, book...)
func (db *mongoDB) AddRecord(ctx context.Context, format string, record map[string]string) (string, error) {
	if col, found := db.collections[format]; found {
		if _, found := record["id"]; found {
			return "", errors.New("new record cannot have id")
		}
		if res, err := col.InsertOne(ctx, record); err == nil {
			return res.InsertedID.(primitive.ObjectID).Hex(), err
		} else {
			return "", err
		}
	}
	return "", errors.New("format not found")
}

// Update updates a record (author, book...)
func (db *mongoDB) UpdateRecord(ctx context.Context, format string, record map[string]string) error {
	if col, found := db.collections[format]; found {
		if id, fields := splitID(record); id != "" {
			// Get ObjectID as used by MongoDB
			objectID, _ := primitive.ObjectIDFromHex(id)
			_, err := col.ReplaceOne(ctx, bson.M{"_id": objectID},
				fields)
			return err
		}
		return errors.New("record doesn't have id")
	}
	return errors.New("format not found")
}

// GetAll returns all records of a specific format from the database
func (db *mongoDB) GetAllRecords(ctx context.Context, format string) ([]map[string]string, error) {
	if col, found := db.collections[format]; found {
		cursor, err := col.Find(context.TODO(), bson.M{})
		if err != nil {
			return nil, err
		}
		var documents []map[string]string
		if err = cursor.All(context.TODO(), &documents); err != nil {
			return nil, err
		}
		return documentsToRecords(documents), nil
	}
	return nil, errors.New("format not found")
}

// Get returns a record from the database
func (db *mongoDB) GetRecord(ctx context.Context, format, id string) (map[string]string, error) {
	if col, found := db.collections[format]; found {
		// Get ObjectID as used by MongoDB
		objectID, _ := primitive.ObjectIDFromHex(id)
		var document map[string]string
		err := col.FindOne(context.TODO(),
			bson.M{"_id": objectID}).Decode(&document)
		if err != nil {
			return nil, err
		}
		return documentToRecord(document), nil
	}
	return nil, errors.New("format not found")
}

func splitID(record map[string]string) (id string, recordNoID map[string]string) {
	recordNoID = make(map[string]string)
	for name, value := range record {
		if name == "id" {
			id = value
		} else {
			recordNoID[name] = value
		}
	}
	return id, recordNoID
}

// Return map[string]string slice from MongoDB document slice. It just renames "_id" keys to "id".
func documentsToRecords(docs []map[string]string) []map[string]string {
	records := make([]map[string]string, 0, len(docs))
	for _, d := range docs {
		records = append(records, documentToRecord(d))
	}
	return records
}

// Return map[string]string from MongoDB document. It just renames "_id" key to "id".
func documentToRecord(doc map[string]string) map[string]string {
	if id, found := doc["_id"]; found {
		delete(doc, "_id")
		doc["id"] = id
		return doc
	}
	return doc
}
