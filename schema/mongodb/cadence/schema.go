package cadence

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/uber/cadence/common"
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

var ConfigStoreIndexes = []mongo.IndexModel{
	{
		Keys: bsonx.Doc{
			{
				Key: "rowType", Value: bsonx.Int32(-1),
			},
			{
				Key: "version", Value: bsonx.Int32(-1),
			},
		},
		Options: &options.IndexOptions{
			Unique: common.BoolPtr(true),
		},
	},
}