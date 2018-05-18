// Code generated by thriftrw v1.11.0. DO NOT EDIT.
// @generated

package history

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// HistoryService_RecordActivityTaskHeartbeat_Args represents the arguments for the HistoryService.RecordActivityTaskHeartbeat function.
//
// The arguments for RecordActivityTaskHeartbeat are sent and received over the wire as this struct.
type HistoryService_RecordActivityTaskHeartbeat_Args struct {
	HeartbeatRequest *RecordActivityTaskHeartbeatRequest `json:"heartbeatRequest,omitempty"`
}

// ToWire translates a HistoryService_RecordActivityTaskHeartbeat_Args struct into a Thrift-level intermediate
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
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.HeartbeatRequest != nil {
		w, err = v.HeartbeatRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RecordActivityTaskHeartbeatRequest_1_Read(w wire.Value) (*RecordActivityTaskHeartbeatRequest, error) {
	var v RecordActivityTaskHeartbeatRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_RecordActivityTaskHeartbeat_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_RecordActivityTaskHeartbeat_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_RecordActivityTaskHeartbeat_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.HeartbeatRequest, err = _RecordActivityTaskHeartbeatRequest_1_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a HistoryService_RecordActivityTaskHeartbeat_Args
// struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.HeartbeatRequest != nil {
		fields[i] = fmt.Sprintf("HeartbeatRequest: %v", v.HeartbeatRequest)
		i++
	}

	return fmt.Sprintf("HistoryService_RecordActivityTaskHeartbeat_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_RecordActivityTaskHeartbeat_Args match the
// provided HistoryService_RecordActivityTaskHeartbeat_Args.
//
// This function performs a deep comparison.
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) Equals(rhs *HistoryService_RecordActivityTaskHeartbeat_Args) bool {
	if !((v.HeartbeatRequest == nil && rhs.HeartbeatRequest == nil) || (v.HeartbeatRequest != nil && rhs.HeartbeatRequest != nil && v.HeartbeatRequest.Equals(rhs.HeartbeatRequest))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "RecordActivityTaskHeartbeat" for this struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) MethodName() string {
	return "RecordActivityTaskHeartbeat"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// HistoryService_RecordActivityTaskHeartbeat_Helper provides functions that aid in handling the
// parameters and return values of the HistoryService.RecordActivityTaskHeartbeat
// function.
var HistoryService_RecordActivityTaskHeartbeat_Helper = struct {
	// Args accepts the parameters of RecordActivityTaskHeartbeat in-order and returns
	// the arguments struct for the function.
	Args func(
		heartbeatRequest *RecordActivityTaskHeartbeatRequest,
	) *HistoryService_RecordActivityTaskHeartbeat_Args

	// IsException returns true if the given error can be thrown
	// by RecordActivityTaskHeartbeat.
	//
	// An error can be thrown by RecordActivityTaskHeartbeat only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for RecordActivityTaskHeartbeat
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// RecordActivityTaskHeartbeat into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by RecordActivityTaskHeartbeat
	//
	//   value, err := RecordActivityTaskHeartbeat(args)
	//   result, err := HistoryService_RecordActivityTaskHeartbeat_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from RecordActivityTaskHeartbeat: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*shared.RecordActivityTaskHeartbeatResponse, error) (*HistoryService_RecordActivityTaskHeartbeat_Result, error)

	// UnwrapResponse takes the result struct for RecordActivityTaskHeartbeat
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if RecordActivityTaskHeartbeat threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := HistoryService_RecordActivityTaskHeartbeat_Helper.UnwrapResponse(result)
	UnwrapResponse func(*HistoryService_RecordActivityTaskHeartbeat_Result) (*shared.RecordActivityTaskHeartbeatResponse, error)
}{}

func init() {
	HistoryService_RecordActivityTaskHeartbeat_Helper.Args = func(
		heartbeatRequest *RecordActivityTaskHeartbeatRequest,
	) *HistoryService_RecordActivityTaskHeartbeat_Args {
		return &HistoryService_RecordActivityTaskHeartbeat_Args{
			HeartbeatRequest: heartbeatRequest,
		}
	}

	HistoryService_RecordActivityTaskHeartbeat_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		case *ShardOwnershipLostError:
			return true
		case *shared.DomainNotActiveError:
			return true
		case *shared.LimitExceededError:
			return true
		default:
			return false
		}
	}

	HistoryService_RecordActivityTaskHeartbeat_Helper.WrapResponse = func(success *shared.RecordActivityTaskHeartbeatResponse, err error) (*HistoryService_RecordActivityTaskHeartbeat_Result, error) {
		if err == nil {
			return &HistoryService_RecordActivityTaskHeartbeat_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.BadRequestError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.InternalServiceError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.EntityNotExistError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{EntityNotExistError: e}, nil
		case *ShardOwnershipLostError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.ShardOwnershipLostError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{ShardOwnershipLostError: e}, nil
		case *shared.DomainNotActiveError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.DomainNotActiveError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{DomainNotActiveError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskHeartbeat_Result.LimitExceededError")
			}
			return &HistoryService_RecordActivityTaskHeartbeat_Result{LimitExceededError: e}, nil
		}

		return nil, err
	}
	HistoryService_RecordActivityTaskHeartbeat_Helper.UnwrapResponse = func(result *HistoryService_RecordActivityTaskHeartbeat_Result) (success *shared.RecordActivityTaskHeartbeatResponse, err error) {
		if result.BadRequestError != nil {
			err = result.BadRequestError
			return
		}
		if result.InternalServiceError != nil {
			err = result.InternalServiceError
			return
		}
		if result.EntityNotExistError != nil {
			err = result.EntityNotExistError
			return
		}
		if result.ShardOwnershipLostError != nil {
			err = result.ShardOwnershipLostError
			return
		}
		if result.DomainNotActiveError != nil {
			err = result.DomainNotActiveError
			return
		}
		if result.LimitExceededError != nil {
			err = result.LimitExceededError
			return
		}

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// HistoryService_RecordActivityTaskHeartbeat_Result represents the result of a HistoryService.RecordActivityTaskHeartbeat function call.
//
// The result of a RecordActivityTaskHeartbeat execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type HistoryService_RecordActivityTaskHeartbeat_Result struct {
	// Value returned by RecordActivityTaskHeartbeat after a successful execution.
	Success                 *shared.RecordActivityTaskHeartbeatResponse `json:"success,omitempty"`
	BadRequestError         *shared.BadRequestError                     `json:"badRequestError,omitempty"`
	InternalServiceError    *shared.InternalServiceError                `json:"internalServiceError,omitempty"`
	EntityNotExistError     *shared.EntityNotExistsError                `json:"entityNotExistError,omitempty"`
	ShardOwnershipLostError *ShardOwnershipLostError                    `json:"shardOwnershipLostError,omitempty"`
	DomainNotActiveError    *shared.DomainNotActiveError                `json:"domainNotActiveError,omitempty"`
	LimitExceededError      *shared.LimitExceededError                  `json:"limitExceededError,omitempty"`
}

// ToWire translates a HistoryService_RecordActivityTaskHeartbeat_Result struct into a Thrift-level intermediate
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
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) ToWire() (wire.Value, error) {
	var (
		fields [7]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Success != nil {
		w, err = v.Success.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 0, Value: w}
		i++
	}
	if v.BadRequestError != nil {
		w, err = v.BadRequestError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	if v.InternalServiceError != nil {
		w, err = v.InternalServiceError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	if v.EntityNotExistError != nil {
		w, err = v.EntityNotExistError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 3, Value: w}
		i++
	}
	if v.ShardOwnershipLostError != nil {
		w, err = v.ShardOwnershipLostError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 4, Value: w}
		i++
	}
	if v.DomainNotActiveError != nil {
		w, err = v.DomainNotActiveError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}
	if v.LimitExceededError != nil {
		w, err = v.LimitExceededError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 6, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("HistoryService_RecordActivityTaskHeartbeat_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RecordActivityTaskHeartbeatResponse_Read(w wire.Value) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	var v shared.RecordActivityTaskHeartbeatResponse
	err := v.FromWire(w)
	return &v, err
}

func _DomainNotActiveError_Read(w wire.Value) (*shared.DomainNotActiveError, error) {
	var v shared.DomainNotActiveError
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_RecordActivityTaskHeartbeat_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_RecordActivityTaskHeartbeat_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_RecordActivityTaskHeartbeat_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _RecordActivityTaskHeartbeatResponse_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.BadRequestError, err = _BadRequestError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 2:
			if field.Value.Type() == wire.TStruct {
				v.InternalServiceError, err = _InternalServiceError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 3:
			if field.Value.Type() == wire.TStruct {
				v.EntityNotExistError, err = _EntityNotExistsError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 4:
			if field.Value.Type() == wire.TStruct {
				v.ShardOwnershipLostError, err = _ShardOwnershipLostError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 5:
			if field.Value.Type() == wire.TStruct {
				v.DomainNotActiveError, err = _DomainNotActiveError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 6:
			if field.Value.Type() == wire.TStruct {
				v.LimitExceededError, err = _LimitExceededError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	count := 0
	if v.Success != nil {
		count++
	}
	if v.BadRequestError != nil {
		count++
	}
	if v.InternalServiceError != nil {
		count++
	}
	if v.EntityNotExistError != nil {
		count++
	}
	if v.ShardOwnershipLostError != nil {
		count++
	}
	if v.DomainNotActiveError != nil {
		count++
	}
	if v.LimitExceededError != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("HistoryService_RecordActivityTaskHeartbeat_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a HistoryService_RecordActivityTaskHeartbeat_Result
// struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [7]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}
	if v.BadRequestError != nil {
		fields[i] = fmt.Sprintf("BadRequestError: %v", v.BadRequestError)
		i++
	}
	if v.InternalServiceError != nil {
		fields[i] = fmt.Sprintf("InternalServiceError: %v", v.InternalServiceError)
		i++
	}
	if v.EntityNotExistError != nil {
		fields[i] = fmt.Sprintf("EntityNotExistError: %v", v.EntityNotExistError)
		i++
	}
	if v.ShardOwnershipLostError != nil {
		fields[i] = fmt.Sprintf("ShardOwnershipLostError: %v", v.ShardOwnershipLostError)
		i++
	}
	if v.DomainNotActiveError != nil {
		fields[i] = fmt.Sprintf("DomainNotActiveError: %v", v.DomainNotActiveError)
		i++
	}
	if v.LimitExceededError != nil {
		fields[i] = fmt.Sprintf("LimitExceededError: %v", v.LimitExceededError)
		i++
	}

	return fmt.Sprintf("HistoryService_RecordActivityTaskHeartbeat_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_RecordActivityTaskHeartbeat_Result match the
// provided HistoryService_RecordActivityTaskHeartbeat_Result.
//
// This function performs a deep comparison.
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) Equals(rhs *HistoryService_RecordActivityTaskHeartbeat_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
		return false
	}
	if !((v.EntityNotExistError == nil && rhs.EntityNotExistError == nil) || (v.EntityNotExistError != nil && rhs.EntityNotExistError != nil && v.EntityNotExistError.Equals(rhs.EntityNotExistError))) {
		return false
	}
	if !((v.ShardOwnershipLostError == nil && rhs.ShardOwnershipLostError == nil) || (v.ShardOwnershipLostError != nil && rhs.ShardOwnershipLostError != nil && v.ShardOwnershipLostError.Equals(rhs.ShardOwnershipLostError))) {
		return false
	}
	if !((v.DomainNotActiveError == nil && rhs.DomainNotActiveError == nil) || (v.DomainNotActiveError != nil && rhs.DomainNotActiveError != nil && v.DomainNotActiveError.Equals(rhs.DomainNotActiveError))) {
		return false
	}
	if !((v.LimitExceededError == nil && rhs.LimitExceededError == nil) || (v.LimitExceededError != nil && rhs.LimitExceededError != nil && v.LimitExceededError.Equals(rhs.LimitExceededError))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "RecordActivityTaskHeartbeat" for this struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) MethodName() string {
	return "RecordActivityTaskHeartbeat"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
