package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName  = "strwiki"
	colName = "col"
)

type Post struct {
	Title string `json:”title,omitempty”`
	Body  string `json:”body,omitempty”`
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

func InsertPost(title string, body string) {
	post := Post{title, body}
	collection := cli.Database("my_database").Collection("posts")
	insertResult, err := collection.InsertOne(context.TODO(), post)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
}

/*
func GetPost(id bson.ObjectId) {
	collection := cli.Database("my_database").Collection("posts")
	filter := bson.D
	var post Post

	err := collection.FindOne(context.TODO(), filter).Decode(&post)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Found post with title ", post.Title)
}*/
