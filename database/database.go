package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName  = "strwiki"
	colName = "col"
)

//@TODO: This may not belong to database
type Record struct {
	DbID   string
	Fields map[string]string
}

// Context, database and prepared statements
var (
	cli *mongo.Client
	col *mongo.Collection
)

// Open opens the database and prepares the statements
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
	col = db.Collection(colName)

	return cli
}

func Add(record map[string]string) {
	//@TODO: record should be marshalled
	insertResult, err := col.InsertOne(context.TODO(), record)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
}

func GetAll() []Record {
	cursor, err := col.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var documents []map[string]string
	if err = cursor.All(context.TODO(), &documents); err != nil {
		log.Fatal(err)
	}
	records := documentsToRecords(documents)
	return records
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
