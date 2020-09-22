package serialization

import (
	"bytes"
	"fmt"

	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
)

type thriftRWType interface {
	ToWire() (wire.Value, error)
	FromWire(w wire.Value) error
}

func validateEncoding(encoding string) error {
	switch common.EncodingType(encoding) {
	case common.EncodingTypeProto, common.EncodingTypeThriftRW:
		return nil
	default:
		return fmt.Errorf("invalid encoding type: %v", encoding)
	}
}

func encodeErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("error serializing struct to blob: %v", err)
}

func decodeErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("error deserializing blob to struct: %v", err)
}

func thriftRWEncode(t thriftRWType) (p.DataBlob, error) {
	blob := p.DataBlob{Encoding: common.EncodingTypeThriftRW}
	value, err := t.ToWire()
	if err != nil {
		return blob, encodeErr(err)
	}
	var b bytes.Buffer
	if err := protocol.Binary.Encode(value, &b); err != nil {
		return blob, encodeErr(err)
	}
	blob.Data = b.Bytes()
	return blob, nil
}

func thriftRWDecode(data []byte, result thriftRWType) error {
	value, err := protocol.Binary.Decode(bytes.NewReader(data), wire.TStruct)
	if err != nil {
		return decodeErr(err)
	}
	return decodeErr(result.FromWire(value))
}
