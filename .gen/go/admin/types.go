// Code generated by thriftrw v1.11.0. DO NOT EDIT.
// @generated

package admin

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type AccessDeniedError struct {
	Message string `json:"message,required"`
}

// ToWire translates a AccessDeniedError struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *AccessDeniedError) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Message), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a AccessDeniedError struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a AccessDeniedError struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v AccessDeniedError
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *AccessDeniedError) FromWire(w wire.Value) error {
	var err error

	messageIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Message, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				messageIsSet = true
			}
		}
	}

	if !messageIsSet {
		return errors.New("field Message of AccessDeniedError is required")
	}

	return nil
}

// String returns a readable string representation of a AccessDeniedError
// struct.
func (v *AccessDeniedError) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	fields[i] = fmt.Sprintf("Message: %v", v.Message)
	i++

	return fmt.Sprintf("AccessDeniedError{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this AccessDeniedError match the
// provided AccessDeniedError.
//
// This function performs a deep comparison.
func (v *AccessDeniedError) Equals(rhs *AccessDeniedError) bool {
	if !(v.Message == rhs.Message) {
		return false
	}

	return true
}

func (v *AccessDeniedError) Error() string {
	return v.String()
}

type InquiryWorkflowExecutionResponse struct {
	ShardId      *string `json:"shardId,omitempty"`
	HistoryAddr  *string `json:"historyAddr,omitempty"`
	MatchingAddr *string `json:"matchingAddr,omitempty"`
}

// ToWire translates a InquiryWorkflowExecutionResponse struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//   x, err := v.ToWire()
//   if err != nil {
//     return err
//   }
//
//   if err := binaryProtocol.Encode(x, writer); err != nil {
//     return err
//   }
func (v *InquiryWorkflowExecutionResponse) ToWire() (wire.Value, error) {
	var (
		fields [3]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.ShardId != nil {
		w, err = wire.NewValueString(*(v.ShardId)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.HistoryAddr != nil {
		w, err = wire.NewValueString(*(v.HistoryAddr)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}
	if v.MatchingAddr != nil {
		w, err = wire.NewValueString(*(v.MatchingAddr)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 30, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a InquiryWorkflowExecutionResponse struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a InquiryWorkflowExecutionResponse struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v InquiryWorkflowExecutionResponse
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *InquiryWorkflowExecutionResponse) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.ShardId = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.HistoryAddr = &x
				if err != nil {
					return err
				}

			}
		case 30:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.MatchingAddr = &x
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a InquiryWorkflowExecutionResponse
// struct.
func (v *InquiryWorkflowExecutionResponse) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [3]string
	i := 0
	if v.ShardId != nil {
		fields[i] = fmt.Sprintf("ShardId: %v", *(v.ShardId))
		i++
	}
	if v.HistoryAddr != nil {
		fields[i] = fmt.Sprintf("HistoryAddr: %v", *(v.HistoryAddr))
		i++
	}
	if v.MatchingAddr != nil {
		fields[i] = fmt.Sprintf("MatchingAddr: %v", *(v.MatchingAddr))
		i++
	}

	return fmt.Sprintf("InquiryWorkflowExecutionResponse{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this InquiryWorkflowExecutionResponse match the
// provided InquiryWorkflowExecutionResponse.
//
// This function performs a deep comparison.
func (v *InquiryWorkflowExecutionResponse) Equals(rhs *InquiryWorkflowExecutionResponse) bool {
	if !_String_EqualsPtr(v.ShardId, rhs.ShardId) {
		return false
	}
	if !_String_EqualsPtr(v.HistoryAddr, rhs.HistoryAddr) {
		return false
	}
	if !_String_EqualsPtr(v.MatchingAddr, rhs.MatchingAddr) {
		return false
	}

	return true
}

// GetShardId returns the value of ShardId if it is set or its
// zero value if it is unset.
func (v *InquiryWorkflowExecutionResponse) GetShardId() (o string) {
	if v.ShardId != nil {
		return *v.ShardId
	}

	return
}

// GetHistoryAddr returns the value of HistoryAddr if it is set or its
// zero value if it is unset.
func (v *InquiryWorkflowExecutionResponse) GetHistoryAddr() (o string) {
	if v.HistoryAddr != nil {
		return *v.HistoryAddr
	}

	return
}

// GetMatchingAddr returns the value of MatchingAddr if it is set or its
// zero value if it is unset.
func (v *InquiryWorkflowExecutionResponse) GetMatchingAddr() (o string) {
	if v.MatchingAddr != nil {
		return *v.MatchingAddr
	}

	return
}
