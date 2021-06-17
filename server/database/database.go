package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ivanmartinez/boocat/log"
	bcerrors "github.com/ivanmartinez/boocat/server/errors"
	"github.com/ivanmartinez/boocat/server/formats"
)

const (
	dbName = "boocat"
)

// DB is the database interface
type DB interface {
	AddRecord(ctx context.Context, formatName string, record map[string]string) (string, error)
	UpdateRecord(ctx context.Context, formatName string, record map[string]string) error
	GetAllRecords(ctx context.Context, formatName string) ([]map[string]string, error)
	GetRecord(ctx context.Context, formatName, id string) (map[string]string, error)
	SearchRecord(ctx context.Context, formatName, value string) ([]map[string]string, error)
	ReferenceValidator(formatName string) formats.Validate
}

// mongoDB database
type mongoDB struct {
	// MongoDB client
	client *mongo.Client
	// Map of collections. Every collection contains the records of a format
	// (author, book...)
	collections map[string]*mongo.Collection
}

// Name and fields of a collection text index
type textIndex struct {
	name   string
	fields map[string]struct{}
}

// Initialize connects to the database and initializes the collections
// @TODO: Maybe the initialization of collections should be separated
func Initialize(ctx context.Context, dbURI *string, formats map[string]formats.Format) *mongoDB {

	// Create and connect the client
	cli, err := mongo.NewClient(options.Client().ApplyURI(*dbURI))
	if err != nil {
		log.Error.Fatal(err)
	}
	if err = cli.Connect(ctx); err != nil {
		log.Error.Fatal(err)
	}

	// Initialize the collections
	db := cli.Database(dbName)
	collections := make(map[string]*mongo.Collection, len(formats))
	for _, format := range formats {
		collection := db.Collection(format.Name)
		collections[format.Name] = collection
		indexes := collection.Indexes()
		if index, ok := findTextIndex(ctx, indexes); ok {
			if !format.SearchableAre(index.fields) {
				if _, err := indexes.DropOne(ctx, index.name); err != nil {
					log.Error.Fatal(err)
				}
				indexModel := textIndexModel(format)
				if _, err := indexes.CreateOne(ctx, indexModel); err != nil {
					log.Error.Fatal(err)
				}
				log.Info.Printf("Updated text index for %v", format.Name)
			}
		} else {
			indexModel := textIndexModel(format)
			if _, err := indexes.CreateOne(ctx, indexModel); err != nil {
				log.Error.Fatal(err)
			}
			log.Info.Printf("Created text index for %v", format.Name)
		}
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
func (db *mongoDB) AddRecord(ctx context.Context, formatName string, record map[string]string) (string, error) {
	col, found := db.collections[formatName]
	if !found {
		return "", bcerrors.ErrFormatNotFound
	}
	if _, found := record["id"]; found {
		return "", bcerrors.ErrRecordHasID
	}
	res, err := col.InsertOne(ctx, record)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), err
}

// UpdateRecord updates a record (author, book...)
func (db *mongoDB) UpdateRecord(ctx context.Context, formatName string, record map[string]string) error {
	col, found := db.collections[formatName]
	if !found {
		return bcerrors.ErrFormatNotFound
	}
	id, fields := splitID(record)
	if id == "" {
		return bcerrors.ErrRecordDoesntHaveID
	}
	// Get ObjectID as used by MongoDB
	objectID, _ := primitive.ObjectIDFromHex(id)
	result, err := col.ReplaceOne(ctx, bson.M{"_id": objectID},
		fields)
	if result.MatchedCount != 1 {
		return bcerrors.ErrRecordNotFound
	}
	return err
}

// GetRecord returns a record from the database by the id field
func (db *mongoDB) GetRecord(ctx context.Context, formatName, id string) (map[string]string, error) {
	col, found := db.collections[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	// Get ObjectID as used by MongoDB
	objectID, _ := primitive.ObjectIDFromHex(id)
	var document map[string]string
	err := col.FindOne(ctx, bson.M{"_id": objectID}).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, bcerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return documentToRecord(document), nil
}

// GetAllRecords returns all records of a specific format from the database
func (db *mongoDB) GetAllRecords(ctx context.Context, formatName string) ([]map[string]string, error) {
	col, found := db.collections[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	cursor, err := col.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	var documents []map[string]string
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	return documentsToRecords(documents), nil
}

// SearchRecord returns all records of a specific format from the database that matches the search term
func (db *mongoDB) SearchRecord(ctx context.Context, formatName, search string) ([]map[string]string, error) {
	col, found := db.collections[formatName]
	if !found {
		return nil, bcerrors.ErrFormatNotFound
	}
	cursor, err := col.Find(ctx, bson.M{"$text": bson.M{"$search": search}})
	if err != nil {
		return nil, err
	}
	var documents []map[string]string
	if err = cursor.All(context.TODO(), &documents); err != nil {
		return nil, err
	}
	return documentsToRecords(documents), nil
}

// referenceValidator returns a validator for references to records of the passed format name
func (db *mongoDB) ReferenceValidator(formatName string) formats.Validate {
	return func(ctx context.Context, value interface{}) string {
		stringValue := fmt.Sprintf("%v", value)
		if _, err := db.GetRecord(ctx, formatName, stringValue); err != nil {
			return fmt.Sprintf("record of format '%s' and ID '%s' not found", formatName, stringValue)
		}
		return ""
	}
}

// findTextIndex looks for a text index in the passed indexes and returns it if found
func findTextIndex(ctx context.Context, indexes mongo.IndexView) (textIndex, bool) {
	// Iterate through the collection's indexes
	cursor, err := indexes.List(ctx)
	if err != nil {
		log.Error.Fatal(err)
	}
	var result bson.M
	for cursor.Next(ctx) {
		err := cursor.Decode(&result)
		if err != nil {
			log.Error.Fatal(err)
		}
		if keyValue, found := result["key"]; found {
			if keyMap, ok := keyValue.(bson.M); ok {
				if keyMap["_fts"] == "text" {
					// It's a text index. Get the name.
					if name, ok := result["name"].(string); ok {
						index := textIndex{
							name:   name,
							fields: map[string]struct{}{},
						}
						// Get the fields
						if weightsValue, found := result["weights"]; found {
							if weightsMap, ok := weightsValue.(bson.M); ok {
								for field := range weightsMap {
									index.fields[field] = struct{}{}
								}
							}
						}
						// Return the index data
						return index, true
					}
				}
			}
		}
	}
	// No text index found
	return textIndex{}, false
}

// textIndexModel returns the model to create a text index that matches the searchable fields of the format
func textIndexModel(format formats.Format) mongo.IndexModel {
	keys := make(bson.D, 0, len(format.Searchable))
	for field := range format.Searchable {
		keys = append(keys, primitive.E{Key: field, Value: "text"})
	}
	return mongo.IndexModel{
		Keys: keys,
	}
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
