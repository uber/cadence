package serialization

import "github.com/uber/cadence/common"

type (
	parser struct{
		encoder Encoder
		decoders map[string]Decoder
	}
)

func NewParser(encodingType common.EncodingType, supportedDecodingTypes ...common.EncodingType) (Parser, error) {
	p := &parser{}
	if err := validateEncoding(encodingType); err != nil {
		return nil, err
	}
	switch encodingType {
	case common.EncodingTypeThriftRW:
		p.encoder = NewThriftEncoder()
	case common.EncodingTypeProto:
		p.encoder = NewProtoEncoder()
	default:
		panic("unsupported encoding format")
	}
	for _, dt := range supportedDecodingTypes {
		if err := validateEncoding(dt); err != nil {
			return nil, err
		}
		switch dt {
		case common.EncodingTypeThriftRW:
			p.decoders[string(dt)] =
		}
	}
}