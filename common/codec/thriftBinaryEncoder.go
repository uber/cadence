package codec

import (
	"errors"
)

var nonThriftObjectError = errors.New("not a thrift object")

type thriftBinaryEncoder struct {
	thriftEncoder ThriftEncoder
}

func NewThriftBinaryEncoder() BinaryEncoder {
	return &thriftBinaryEncoder{
		thriftEncoder: NewThriftRWEncoder(),
	}
}

func (e *thriftBinaryEncoder) Encode(object interface{}) ([]byte, error) {
	thriftObject, ok := object.(ThriftObject)
	if !ok {
		return nil, nonThriftObjectError
	}

	return e.thriftEncoder.Encode(thriftObject)
}

func (e *thriftBinaryEncoder) Decode(payload []byte, value interface{}) error {
	thriftValue, ok := value.(ThriftObject)
	if !ok {
		return nonThriftObjectError
	}

	return e.thriftEncoder.Decode(payload, thriftValue)
}
