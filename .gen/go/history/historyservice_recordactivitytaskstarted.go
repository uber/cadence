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

// HistoryService_RecordActivityTaskStarted_Args represents the arguments for the HistoryService.RecordActivityTaskStarted function.
//
// The arguments for RecordActivityTaskStarted are sent and received over the wire as this struct.
type HistoryService_RecordActivityTaskStarted_Args struct {
	AddRequest *RecordActivityTaskStartedRequest `json:"addRequest,omitempty"`
}

// ToWire translates a HistoryService_RecordActivityTaskStarted_Args struct into a Thrift-level intermediate
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
func (v *HistoryService_RecordActivityTaskStarted_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.AddRequest != nil {
		w, err = v.AddRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RecordActivityTaskStartedRequest_Read(w wire.Value) (*RecordActivityTaskStartedRequest, error) {
	var v RecordActivityTaskStartedRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_RecordActivityTaskStarted_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_RecordActivityTaskStarted_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_RecordActivityTaskStarted_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_RecordActivityTaskStarted_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.AddRequest, err = _RecordActivityTaskStartedRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a HistoryService_RecordActivityTaskStarted_Args
// struct.
func (v *HistoryService_RecordActivityTaskStarted_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.AddRequest != nil {
		fields[i] = fmt.Sprintf("AddRequest: %v", v.AddRequest)
		i++
	}

	return fmt.Sprintf("HistoryService_RecordActivityTaskStarted_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_RecordActivityTaskStarted_Args match the
// provided HistoryService_RecordActivityTaskStarted_Args.
//
// This function performs a deep comparison.
func (v *HistoryService_RecordActivityTaskStarted_Args) Equals(rhs *HistoryService_RecordActivityTaskStarted_Args) bool {
	if !((v.AddRequest == nil && rhs.AddRequest == nil) || (v.AddRequest != nil && rhs.AddRequest != nil && v.AddRequest.Equals(rhs.AddRequest))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "RecordActivityTaskStarted" for this struct.
func (v *HistoryService_RecordActivityTaskStarted_Args) MethodName() string {
	return "RecordActivityTaskStarted"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *HistoryService_RecordActivityTaskStarted_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// HistoryService_RecordActivityTaskStarted_Helper provides functions that aid in handling the
// parameters and return values of the HistoryService.RecordActivityTaskStarted
// function.
var HistoryService_RecordActivityTaskStarted_Helper = struct {
	// Args accepts the parameters of RecordActivityTaskStarted in-order and returns
	// the arguments struct for the function.
	Args func(
		addRequest *RecordActivityTaskStartedRequest,
	) *HistoryService_RecordActivityTaskStarted_Args

	// IsException returns true if the given error can be thrown
	// by RecordActivityTaskStarted.
	//
	// An error can be thrown by RecordActivityTaskStarted only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for RecordActivityTaskStarted
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// RecordActivityTaskStarted into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by RecordActivityTaskStarted
	//
	//   value, err := RecordActivityTaskStarted(args)
	//   result, err := HistoryService_RecordActivityTaskStarted_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from RecordActivityTaskStarted: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*RecordActivityTaskStartedResponse, error) (*HistoryService_RecordActivityTaskStarted_Result, error)

	// UnwrapResponse takes the result struct for RecordActivityTaskStarted
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if RecordActivityTaskStarted threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := HistoryService_RecordActivityTaskStarted_Helper.UnwrapResponse(result)
	UnwrapResponse func(*HistoryService_RecordActivityTaskStarted_Result) (*RecordActivityTaskStartedResponse, error)
}{}

func init() {
	HistoryService_RecordActivityTaskStarted_Helper.Args = func(
		addRequest *RecordActivityTaskStartedRequest,
	) *HistoryService_RecordActivityTaskStarted_Args {
		return &HistoryService_RecordActivityTaskStarted_Args{
			AddRequest: addRequest,
		}
	}

	HistoryService_RecordActivityTaskStarted_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *EventAlreadyStartedError:
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

	HistoryService_RecordActivityTaskStarted_Helper.WrapResponse = func(success *RecordActivityTaskStartedResponse, err error) (*HistoryService_RecordActivityTaskStarted_Result, error) {
		if err == nil {
			return &HistoryService_RecordActivityTaskStarted_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.BadRequestError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.InternalServiceError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{InternalServiceError: e}, nil
		case *EventAlreadyStartedError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.EventAlreadyStartedError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{EventAlreadyStartedError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.EntityNotExistError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{EntityNotExistError: e}, nil
		case *ShardOwnershipLostError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.ShardOwnershipLostError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{ShardOwnershipLostError: e}, nil
		case *shared.DomainNotActiveError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.DomainNotActiveError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{DomainNotActiveError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_RecordActivityTaskStarted_Result.LimitExceededError")
			}
			return &HistoryService_RecordActivityTaskStarted_Result{LimitExceededError: e}, nil
		}

		return nil, err
	}
	HistoryService_RecordActivityTaskStarted_Helper.UnwrapResponse = func(result *HistoryService_RecordActivityTaskStarted_Result) (success *RecordActivityTaskStartedResponse, err error) {
		if result.BadRequestError != nil {
			err = result.BadRequestError
			return
		}
		if result.InternalServiceError != nil {
			err = result.InternalServiceError
			return
		}
		if result.EventAlreadyStartedError != nil {
			err = result.EventAlreadyStartedError
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

// HistoryService_RecordActivityTaskStarted_Result represents the result of a HistoryService.RecordActivityTaskStarted function call.
//
// The result of a RecordActivityTaskStarted execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type HistoryService_RecordActivityTaskStarted_Result struct {
	// Value returned by RecordActivityTaskStarted after a successful execution.
	Success                  *RecordActivityTaskStartedResponse `json:"success,omitempty"`
	BadRequestError          *shared.BadRequestError            `json:"badRequestError,omitempty"`
	InternalServiceError     *shared.InternalServiceError       `json:"internalServiceError,omitempty"`
	EventAlreadyStartedError *EventAlreadyStartedError          `json:"eventAlreadyStartedError,omitempty"`
	EntityNotExistError      *shared.EntityNotExistsError       `json:"entityNotExistError,omitempty"`
	ShardOwnershipLostError  *ShardOwnershipLostError           `json:"shardOwnershipLostError,omitempty"`
	DomainNotActiveError     *shared.DomainNotActiveError       `json:"domainNotActiveError,omitempty"`
	LimitExceededError       *shared.LimitExceededError         `json:"limitExceededError,omitempty"`
}

// ToWire translates a HistoryService_RecordActivityTaskStarted_Result struct into a Thrift-level intermediate
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
func (v *HistoryService_RecordActivityTaskStarted_Result) ToWire() (wire.Value, error) {
	var (
		fields [8]wire.Field
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
	if v.EventAlreadyStartedError != nil {
		w, err = v.EventAlreadyStartedError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 3, Value: w}
		i++
	}
	if v.EntityNotExistError != nil {
		w, err = v.EntityNotExistError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 4, Value: w}
		i++
	}
	if v.ShardOwnershipLostError != nil {
		w, err = v.ShardOwnershipLostError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}
	if v.DomainNotActiveError != nil {
		w, err = v.DomainNotActiveError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 6, Value: w}
		i++
	}
	if v.LimitExceededError != nil {
		w, err = v.LimitExceededError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 7, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("HistoryService_RecordActivityTaskStarted_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RecordActivityTaskStartedResponse_Read(w wire.Value) (*RecordActivityTaskStartedResponse, error) {
	var v RecordActivityTaskStartedResponse
	err := v.FromWire(w)
	return &v, err
}

func _EventAlreadyStartedError_Read(w wire.Value) (*EventAlreadyStartedError, error) {
	var v EventAlreadyStartedError
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_RecordActivityTaskStarted_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_RecordActivityTaskStarted_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_RecordActivityTaskStarted_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_RecordActivityTaskStarted_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _RecordActivityTaskStartedResponse_Read(field.Value)
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
				v.EventAlreadyStartedError, err = _EventAlreadyStartedError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 4:
			if field.Value.Type() == wire.TStruct {
				v.EntityNotExistError, err = _EntityNotExistsError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 5:
			if field.Value.Type() == wire.TStruct {
				v.ShardOwnershipLostError, err = _ShardOwnershipLostError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 6:
			if field.Value.Type() == wire.TStruct {
				v.DomainNotActiveError, err = _DomainNotActiveError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 7:
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
	if v.EventAlreadyStartedError != nil {
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
		return fmt.Errorf("HistoryService_RecordActivityTaskStarted_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a HistoryService_RecordActivityTaskStarted_Result
// struct.
func (v *HistoryService_RecordActivityTaskStarted_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [8]string
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
	if v.EventAlreadyStartedError != nil {
		fields[i] = fmt.Sprintf("EventAlreadyStartedError: %v", v.EventAlreadyStartedError)
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

	return fmt.Sprintf("HistoryService_RecordActivityTaskStarted_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_RecordActivityTaskStarted_Result match the
// provided HistoryService_RecordActivityTaskStarted_Result.
//
// This function performs a deep comparison.
func (v *HistoryService_RecordActivityTaskStarted_Result) Equals(rhs *HistoryService_RecordActivityTaskStarted_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
		return false
	}
	if !((v.EventAlreadyStartedError == nil && rhs.EventAlreadyStartedError == nil) || (v.EventAlreadyStartedError != nil && rhs.EventAlreadyStartedError != nil && v.EventAlreadyStartedError.Equals(rhs.EventAlreadyStartedError))) {
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
// This will always be "RecordActivityTaskStarted" for this struct.
func (v *HistoryService_RecordActivityTaskStarted_Result) MethodName() string {
	return "RecordActivityTaskStarted"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *HistoryService_RecordActivityTaskStarted_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
