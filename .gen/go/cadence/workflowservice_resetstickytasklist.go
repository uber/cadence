// Code generated by thriftrw v1.12.0. DO NOT EDIT.
// @generated

package cadence

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// WorkflowService_ResetStickyTaskList_Args represents the arguments for the WorkflowService.ResetStickyTaskList function.
//
// The arguments for ResetStickyTaskList are sent and received over the wire as this struct.
type WorkflowService_ResetStickyTaskList_Args struct {
	ResetRequest *shared.ResetStickyTaskListRequest `json:"resetRequest,omitempty"`
}

// ToWire translates a WorkflowService_ResetStickyTaskList_Args struct into a Thrift-level intermediate
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
func (v *WorkflowService_ResetStickyTaskList_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.ResetRequest != nil {
		w, err = v.ResetRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ResetStickyTaskListRequest_Read(w wire.Value) (*shared.ResetStickyTaskListRequest, error) {
	var v shared.ResetStickyTaskListRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_ResetStickyTaskList_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_ResetStickyTaskList_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_ResetStickyTaskList_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_ResetStickyTaskList_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.ResetRequest, err = _ResetStickyTaskListRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_ResetStickyTaskList_Args
// struct.
func (v *WorkflowService_ResetStickyTaskList_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.ResetRequest != nil {
		fields[i] = fmt.Sprintf("ResetRequest: %v", v.ResetRequest)
		i++
	}

	return fmt.Sprintf("WorkflowService_ResetStickyTaskList_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_ResetStickyTaskList_Args match the
// provided WorkflowService_ResetStickyTaskList_Args.
//
// This function performs a deep comparison.
func (v *WorkflowService_ResetStickyTaskList_Args) Equals(rhs *WorkflowService_ResetStickyTaskList_Args) bool {
	if !((v.ResetRequest == nil && rhs.ResetRequest == nil) || (v.ResetRequest != nil && rhs.ResetRequest != nil && v.ResetRequest.Equals(rhs.ResetRequest))) {
		return false
	}

	return true
}

// GetResetRequest returns the value of ResetRequest if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Args) GetResetRequest() (o *shared.ResetStickyTaskListRequest) {
	if v.ResetRequest != nil {
		return v.ResetRequest
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "ResetStickyTaskList" for this struct.
func (v *WorkflowService_ResetStickyTaskList_Args) MethodName() string {
	return "ResetStickyTaskList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *WorkflowService_ResetStickyTaskList_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// WorkflowService_ResetStickyTaskList_Helper provides functions that aid in handling the
// parameters and return values of the WorkflowService.ResetStickyTaskList
// function.
var WorkflowService_ResetStickyTaskList_Helper = struct {
	// Args accepts the parameters of ResetStickyTaskList in-order and returns
	// the arguments struct for the function.
	Args func(
		resetRequest *shared.ResetStickyTaskListRequest,
	) *WorkflowService_ResetStickyTaskList_Args

	// IsException returns true if the given error can be thrown
	// by ResetStickyTaskList.
	//
	// An error can be thrown by ResetStickyTaskList only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for ResetStickyTaskList
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// ResetStickyTaskList into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by ResetStickyTaskList
	//
	//   value, err := ResetStickyTaskList(args)
	//   result, err := WorkflowService_ResetStickyTaskList_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from ResetStickyTaskList: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*shared.ResetStickyTaskListResponse, error) (*WorkflowService_ResetStickyTaskList_Result, error)

	// UnwrapResponse takes the result struct for ResetStickyTaskList
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if ResetStickyTaskList threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := WorkflowService_ResetStickyTaskList_Helper.UnwrapResponse(result)
	UnwrapResponse func(*WorkflowService_ResetStickyTaskList_Result) (*shared.ResetStickyTaskListResponse, error)
}{}

func init() {
	WorkflowService_ResetStickyTaskList_Helper.Args = func(
		resetRequest *shared.ResetStickyTaskListRequest,
	) *WorkflowService_ResetStickyTaskList_Args {
		return &WorkflowService_ResetStickyTaskList_Args{
			ResetRequest: resetRequest,
		}
	}

	WorkflowService_ResetStickyTaskList_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		case *shared.LimitExceededError:
			return true
		case *shared.ServiceBusyError:
			return true
		default:
			return false
		}
	}

	WorkflowService_ResetStickyTaskList_Helper.WrapResponse = func(success *shared.ResetStickyTaskListResponse, err error) (*WorkflowService_ResetStickyTaskList_Result, error) {
		if err == nil {
			return &WorkflowService_ResetStickyTaskList_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_ResetStickyTaskList_Result.BadRequestError")
			}
			return &WorkflowService_ResetStickyTaskList_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_ResetStickyTaskList_Result.InternalServiceError")
			}
			return &WorkflowService_ResetStickyTaskList_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_ResetStickyTaskList_Result.EntityNotExistError")
			}
			return &WorkflowService_ResetStickyTaskList_Result{EntityNotExistError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_ResetStickyTaskList_Result.LimitExceededError")
			}
			return &WorkflowService_ResetStickyTaskList_Result{LimitExceededError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_ResetStickyTaskList_Result.ServiceBusyError")
			}
			return &WorkflowService_ResetStickyTaskList_Result{ServiceBusyError: e}, nil
		}

		return nil, err
	}
	WorkflowService_ResetStickyTaskList_Helper.UnwrapResponse = func(result *WorkflowService_ResetStickyTaskList_Result) (success *shared.ResetStickyTaskListResponse, err error) {
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
		if result.LimitExceededError != nil {
			err = result.LimitExceededError
			return
		}
		if result.ServiceBusyError != nil {
			err = result.ServiceBusyError
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

// WorkflowService_ResetStickyTaskList_Result represents the result of a WorkflowService.ResetStickyTaskList function call.
//
// The result of a ResetStickyTaskList execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type WorkflowService_ResetStickyTaskList_Result struct {
	// Value returned by ResetStickyTaskList after a successful execution.
	Success              *shared.ResetStickyTaskListResponse `json:"success,omitempty"`
	BadRequestError      *shared.BadRequestError             `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError        `json:"internalServiceError,omitempty"`
	EntityNotExistError  *shared.EntityNotExistsError        `json:"entityNotExistError,omitempty"`
	LimitExceededError   *shared.LimitExceededError          `json:"limitExceededError,omitempty"`
	ServiceBusyError     *shared.ServiceBusyError            `json:"serviceBusyError,omitempty"`
}

// ToWire translates a WorkflowService_ResetStickyTaskList_Result struct into a Thrift-level intermediate
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
func (v *WorkflowService_ResetStickyTaskList_Result) ToWire() (wire.Value, error) {
	var (
		fields [6]wire.Field
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
	if v.LimitExceededError != nil {
		w, err = v.LimitExceededError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 4, Value: w}
		i++
	}
	if v.ServiceBusyError != nil {
		w, err = v.ServiceBusyError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("WorkflowService_ResetStickyTaskList_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ResetStickyTaskListResponse_Read(w wire.Value) (*shared.ResetStickyTaskListResponse, error) {
	var v shared.ResetStickyTaskListResponse
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_ResetStickyTaskList_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_ResetStickyTaskList_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_ResetStickyTaskList_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_ResetStickyTaskList_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _ResetStickyTaskListResponse_Read(field.Value)
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
				v.LimitExceededError, err = _LimitExceededError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 5:
			if field.Value.Type() == wire.TStruct {
				v.ServiceBusyError, err = _ServiceBusyError_Read(field.Value)
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
	if v.LimitExceededError != nil {
		count++
	}
	if v.ServiceBusyError != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("WorkflowService_ResetStickyTaskList_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_ResetStickyTaskList_Result
// struct.
func (v *WorkflowService_ResetStickyTaskList_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [6]string
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
	if v.LimitExceededError != nil {
		fields[i] = fmt.Sprintf("LimitExceededError: %v", v.LimitExceededError)
		i++
	}
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}

	return fmt.Sprintf("WorkflowService_ResetStickyTaskList_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_ResetStickyTaskList_Result match the
// provided WorkflowService_ResetStickyTaskList_Result.
//
// This function performs a deep comparison.
func (v *WorkflowService_ResetStickyTaskList_Result) Equals(rhs *WorkflowService_ResetStickyTaskList_Result) bool {
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
	if !((v.LimitExceededError == nil && rhs.LimitExceededError == nil) || (v.LimitExceededError != nil && rhs.LimitExceededError != nil && v.LimitExceededError.Equals(rhs.LimitExceededError))) {
		return false
	}
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}

	return true
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetSuccess() (o *shared.ResetStickyTaskListResponse) {
	if v.Success != nil {
		return v.Success
	}

	return
}

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// GetEntityNotExistError returns the value of EntityNotExistError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetEntityNotExistError() (o *shared.EntityNotExistsError) {
	if v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}

	return
}

// GetLimitExceededError returns the value of LimitExceededError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetLimitExceededError() (o *shared.LimitExceededError) {
	if v.LimitExceededError != nil {
		return v.LimitExceededError
	}

	return
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_ResetStickyTaskList_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "ResetStickyTaskList" for this struct.
func (v *WorkflowService_ResetStickyTaskList_Result) MethodName() string {
	return "ResetStickyTaskList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *WorkflowService_ResetStickyTaskList_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
