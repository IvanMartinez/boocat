package database

import (
	"context"
	"log"

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
	DbID   string
	Fields map[string]string
}

// Database collections
var (
	recordsCol *mongo.Collection
)

// Connect connects to the database and initialies the collections
func Connect(ctx context.Context, dbURI *string) *mongo.Client {
	cli, err := mongo.NewClient(options.Client().ApplyURI(*dbURI))
	if err != nil {
		log.Fatal(err)
	}

	err = cli.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db := cli.Database(dbName)
	recordsCol = db.Collection(recsColName)

	return cli
}

// Add adds a new record to the database with the given fields
func Add(ctx context.Context, fields map[string]string) error {
	_, err := recordsCol.InsertOne(ctx, fields)
	return err
}

// Update updates a record in the database
func Update(ctx context.Context, record Record) error {
	objectID, _ := primitive.ObjectIDFromHex(record.DbID)
	_, err := recordsCol.ReplaceOne(ctx, bson.M{"_id": objectID},
		record.Fields)
	return err
}

// GetAll returns all records from
func GetAll(ctx context.Context) ([]Record, error) {
	cursor, err := recordsCol.Find(context.TODO(), bson.M{})
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
func Get(ctx context.Context, id string) (*Record, error) {
	var document map[string]string
	objectID, _ := primitive.ObjectIDFromHex(id)
	err := recordsCol.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&document)
	if err != nil {
		return nil, err
	}

	record := documentToRecord(document)
	return &record, nil
}

func documentsToRecords(maps []map[string]string) (records []Record) {
	records = make([]Record, 0, len(maps))
	for _, m := range maps {
		records = append(records, documentToRecord(m))
	}
	return records
}

func documentToRecord(m map[string]string) (record Record) {
	record.Fields = make(map[string]string)
	for key, value := range m {
		if key == "_id" {
			record.DbID = value
		} else {
			record.Fields[key] = value
		}
	}
	return record
}
