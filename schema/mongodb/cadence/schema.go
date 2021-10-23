package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NOTE: MongoDB is schemaless, so there is no schema operation to update the fields. Here uses Go struct to define the document fields directly.

// ConfigStoreEntry is the schema of configStore
type ConfigStoreEntry struct {
	ID int `json:"_id,omitempty"`
	RowType int `json:"rowType"`
	Version int64 `json:"version"`
	Data []byte `json:"data"`
	DataEncoding string `json:"dataEncoding"`
	Timestamp int64`json:"Timestamp"`
}

func main() {
	// Declare host and port options to pass to the Connect() method
	clientOptions := options.Client().ApplyURI("mongodb://root:cadence@localhost:27017/")
	fmt.Println("clientOptions TYPE:", reflect.TypeOf(clientOptions), "n")

	// Connect to the MongoDB and return Client instance
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("mongo.Connect() ERROR: %v", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	db := client.Database("test4")

	// create collection
	byteValues, err := ioutil.ReadFile("./collection.json")
	var command interface{}
	err = bson.UnmarshalExtJSON(byteValues, false, &command)
	if err != nil {
		// Print any IO errors with the .json file
		fmt.Println("ioutil.ReadFile ERROR:", err)
		os.Exit(0)
	}
	result := db.RunCommand(context.Background(), command)
	fmt.Println("collection created", command, result.Err())

	// create index
	byteValues, err = ioutil.ReadFile("./index.json")
	err = bson.UnmarshalExtJSON(byteValues, false, &command)
	if err != nil {
		// Print any IO errors with the .json file
		fmt.Println("ioutil.ReadFile ERROR:", err)
		os.Exit(0)
	}
	result = db.RunCommand(context.Background(), command)
	fmt.Println("index created", command, result.Err())

	// insert data
	col := db.Collection("test4")
	fmt.Println("Collection type:", reflect.TypeOf(col), "n")

	// Load values from JSON file to model
	byteValues, err = ioutil.ReadFile("./docs.json")
	if err != nil {
		// Print any IO errors with the .json file
		fmt.Println("ioutil.ReadFile ERROR:", err)
	} else {
		// Print the values of the JSON docs, and insert them if no error
		fmt.Println("byteValues:", string(byteValues))

		// Declare an empty slice for the MongoFields docs
		var docs []ConfigStoreEntry

		// Unmarshal the encoded JSON byte string into the slice
		err = json.Unmarshal(byteValues, &docs)

		// Print MongoDB docs object type
		fmt.Println("nMongoFields Docs:", reflect.TypeOf(docs))

		// Iterate the slice of MongoDB struct docs
		for i := range docs {

			// Put the document element in a new variable
			doc := docs[i]
			fmt.Println("ndoc _id:", doc.ID)
			fmt.Println("doc Field Str:", doc.ID)

			// Call the InsertOne() method and pass the context and doc objects
			result, insertErr := col.InsertOne(ctx, doc)

			// Check for any insertion errors
			if insertErr != nil {
				fmt.Println("InsertOne ERROR:", insertErr)
			} else {
				fmt.Println("InsertOne() API result:", result)
			}
		}
	}
}