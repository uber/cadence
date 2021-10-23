package cadence

// below are the names of all mongoDB collections
const (
	ClusterConfigCollectionName = "cluster_config"
)

// NOTE1: MongoDB collection is schemaless -- there is no schema file for collection. We use Go lang structs to define the collection fields.

// NOTE2: MongoDB doesn't allow using camel case or underscore in the field names

// ClusterConfigCollectionEntry is the schema of configStore
// IMPORTANT: making change to this struct is changing the MongoDB collection schema. Please make sure it's backward compatible.
type ClusterConfigCollectionEntry struct {
	ID                   int    `json:"_id,omitempty"`
	RowType              int    `json:"rowtype"`
	Version              int64  `json:"version"`
	Data                 []byte `json:"data"`
	DataEncoding         string `json:"dataencoding"`
	UnixTimestampSeconds int64  `json:"unixTimestampSeconds"`
}
