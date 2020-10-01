package types

type (
	// Closeable is an interface for any entity that supports a close operation to release resources
	Closeable interface {
		Close()
	}
)

// Data encoding types
const (
	EncodingTypeJSON     EncodingType = "json"
	EncodingTypeThriftRW EncodingType = "thriftrw"
	EncodingTypeGob      EncodingType = "gob"
	EncodingTypeUnknown  EncodingType = "unknow"
	EncodingTypeEmpty    EncodingType = ""
	EncodingTypeProto    EncodingType = "proto3"
)

type (
	// EncodingType is an enum that represents various data encoding types
	EncodingType string
)
