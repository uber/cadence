package persistence

import (
	"encoding/json"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"errors"
)

type (
	// HistorySerializer is used to serialize/deserialize history
	HistorySerializer interface {
		Serialize(version int, history []*workflow.HistoryEvent) (*SerializedHistory, error)
		Deserialize(version int, data []byte) (*History, error)
	}

	// HistorySerializerFactory is a factory that vends
	// HistorySerializers based on encoding type.
	HistorySerializerFactory interface {
		// Get returns a history serializer corresponding
		// to a given encoding type
		Get(encodingType common.EncodingType) (HistorySerializer, error)
	}

	jsonHistorySerializer struct {}

	serializerFactoryImpl struct{
		jsonSerializer HistorySerializer
	}
)

var ErrUnknownEncodingType = errors.New("unknown encoding type")
var ErrHistoryVersionIncompatible = errors.New("history version incompatible with runtime")

const (
	DefaultHistoryVersion = 1
	MaxSupportedHistoryVersion = DefaultHistoryVersion
	DefaultEncodingType   = common.EncodingTypeJSON
)

// NewJSONHistorySerializer returns a JSON HistorySerializer
func NewJSONHistorySerializer() HistorySerializer {
	return &jsonHistorySerializer{}
}

func (j *jsonHistorySerializer) Serialize(version int, history []*workflow.HistoryEvent) (*SerializedHistory, error) {

	if version > MaxSupportedHistoryVersion {
		return nil, ErrHistoryVersionIncompatible
	}

	data, err := json.Marshal(history)
	if err != nil {
		return nil, err
	}
	return NewSerializedHistory(data, common.EncodingTypeJSON, version), nil
}

func (j *jsonHistorySerializer) Deserialize(version int, data []byte) (*History, error) {
	if version > MaxSupportedHistoryVersion {
		return nil, ErrHistoryVersionIncompatible
	}
	var history []*workflow.HistoryEvent
	err := json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	return &History{Version: version, Events: history}, nil
}

// NewHistorySerializerFactory creates and returns an instance
// of HistorySerializerFactory
func NewHistorySerializerFactory() HistorySerializerFactory {
	return &serializerFactoryImpl {
		jsonSerializer: NewJSONHistorySerializer(),
	}
}

func (f *serializerFactoryImpl) Get(encodingType common.EncodingType) (HistorySerializer, error) {
	switch encodingType {
	case common.EncodingTypeJSON:
		return f.jsonSerializer, nil
	default:
		return nil, ErrUnknownEncodingType
	}
}