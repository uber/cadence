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

// WorkflowService_RespondDecisionTaskFailed_Args represents the arguments for the WorkflowService.RespondDecisionTaskFailed function.
//
// The arguments for RespondDecisionTaskFailed are sent and received over the wire as this struct.
type WorkflowService_RespondDecisionTaskFailed_Args struct {
	FailedRequest *shared.RespondDecisionTaskFailedRequest `json:"failedRequest,omitempty"`
}

// ToWire translates a WorkflowService_RespondDecisionTaskFailed_Args struct into a Thrift-level intermediate
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
func (v *WorkflowService_RespondDecisionTaskFailed_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.FailedRequest != nil {
		w, err = v.FailedRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RespondDecisionTaskFailedRequest_Read(w wire.Value) (*shared.RespondDecisionTaskFailedRequest, error) {
	var v shared.RespondDecisionTaskFailedRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_RespondDecisionTaskFailed_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_RespondDecisionTaskFailed_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_RespondDecisionTaskFailed_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_RespondDecisionTaskFailed_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.FailedRequest, err = _RespondDecisionTaskFailedRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_RespondDecisionTaskFailed_Args
// struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.FailedRequest != nil {
		fields[i] = fmt.Sprintf("FailedRequest: %v", v.FailedRequest)
		i++
	}

	return fmt.Sprintf("WorkflowService_RespondDecisionTaskFailed_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_RespondDecisionTaskFailed_Args match the
// provided WorkflowService_RespondDecisionTaskFailed_Args.
//
// This function performs a deep comparison.
func (v *WorkflowService_RespondDecisionTaskFailed_Args) Equals(rhs *WorkflowService_RespondDecisionTaskFailed_Args) bool {
	if !((v.FailedRequest == nil && rhs.FailedRequest == nil) || (v.FailedRequest != nil && rhs.FailedRequest != nil && v.FailedRequest.Equals(rhs.FailedRequest))) {
		return false
	}

	return true
}

// GetFailedRequest returns the value of FailedRequest if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Args) GetFailedRequest() (o *shared.RespondDecisionTaskFailedRequest) {
	if v.FailedRequest != nil {
		return v.FailedRequest
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "RespondDecisionTaskFailed" for this struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Args) MethodName() string {
	return "RespondDecisionTaskFailed"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// WorkflowService_RespondDecisionTaskFailed_Helper provides functions that aid in handling the
// parameters and return values of the WorkflowService.RespondDecisionTaskFailed
// function.
var WorkflowService_RespondDecisionTaskFailed_Helper = struct {
	// Args accepts the parameters of RespondDecisionTaskFailed in-order and returns
	// the arguments struct for the function.
	Args func(
		failedRequest *shared.RespondDecisionTaskFailedRequest,
	) *WorkflowService_RespondDecisionTaskFailed_Args

	// IsException returns true if the given error can be thrown
	// by RespondDecisionTaskFailed.
	//
	// An error can be thrown by RespondDecisionTaskFailed only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for RespondDecisionTaskFailed
	// given the error returned by it. The provided error may
	// be nil if RespondDecisionTaskFailed did not fail.
	//
	// This allows mapping errors returned by RespondDecisionTaskFailed into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// RespondDecisionTaskFailed
	//
	//   err := RespondDecisionTaskFailed(args)
	//   result, err := WorkflowService_RespondDecisionTaskFailed_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from RespondDecisionTaskFailed: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*WorkflowService_RespondDecisionTaskFailed_Result, error)

	// UnwrapResponse takes the result struct for RespondDecisionTaskFailed
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if RespondDecisionTaskFailed threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := WorkflowService_RespondDecisionTaskFailed_Helper.UnwrapResponse(result)
	UnwrapResponse func(*WorkflowService_RespondDecisionTaskFailed_Result) error
}{}

func init() {
	WorkflowService_RespondDecisionTaskFailed_Helper.Args = func(
		failedRequest *shared.RespondDecisionTaskFailedRequest,
	) *WorkflowService_RespondDecisionTaskFailed_Args {
		return &WorkflowService_RespondDecisionTaskFailed_Args{
			FailedRequest: failedRequest,
		}
	}

	WorkflowService_RespondDecisionTaskFailed_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		case *shared.DomainNotActiveError:
			return true
		case *shared.LimitExceededError:
			return true
		case *shared.ServiceBusyError:
			return true
		default:
			return false
		}
	}

	WorkflowService_RespondDecisionTaskFailed_Helper.WrapResponse = func(err error) (*WorkflowService_RespondDecisionTaskFailed_Result, error) {
		if err == nil {
			return &WorkflowService_RespondDecisionTaskFailed_Result{}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.BadRequestError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.InternalServiceError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.EntityNotExistError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{EntityNotExistError: e}, nil
		case *shared.DomainNotActiveError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.DomainNotActiveError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{DomainNotActiveError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.LimitExceededError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{LimitExceededError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RespondDecisionTaskFailed_Result.ServiceBusyError")
			}
			return &WorkflowService_RespondDecisionTaskFailed_Result{ServiceBusyError: e}, nil
		}

		return nil, err
	}
	WorkflowService_RespondDecisionTaskFailed_Helper.UnwrapResponse = func(result *WorkflowService_RespondDecisionTaskFailed_Result) (err error) {
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
		if result.DomainNotActiveError != nil {
			err = result.DomainNotActiveError
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
		return
	}

}

// WorkflowService_RespondDecisionTaskFailed_Result represents the result of a WorkflowService.RespondDecisionTaskFailed function call.
//
// The result of a RespondDecisionTaskFailed execution is sent and received over the wire as this struct.
type WorkflowService_RespondDecisionTaskFailed_Result struct {
	BadRequestError      *shared.BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError `json:"internalServiceError,omitempty"`
	EntityNotExistError  *shared.EntityNotExistsError `json:"entityNotExistError,omitempty"`
	DomainNotActiveError *shared.DomainNotActiveError `json:"domainNotActiveError,omitempty"`
	LimitExceededError   *shared.LimitExceededError   `json:"limitExceededError,omitempty"`
	ServiceBusyError     *shared.ServiceBusyError     `json:"serviceBusyError,omitempty"`
}

// ToWire translates a WorkflowService_RespondDecisionTaskFailed_Result struct into a Thrift-level intermediate
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
func (v *WorkflowService_RespondDecisionTaskFailed_Result) ToWire() (wire.Value, error) {
	var (
		fields [6]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

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
	if v.DomainNotActiveError != nil {
		w, err = v.DomainNotActiveError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 4, Value: w}
		i++
	}
	if v.LimitExceededError != nil {
		w, err = v.LimitExceededError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}
	if v.ServiceBusyError != nil {
		w, err = v.ServiceBusyError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 6, Value: w}
		i++
	}

	if i > 1 {
		return wire.Value{}, fmt.Errorf("WorkflowService_RespondDecisionTaskFailed_Result should have at most one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a WorkflowService_RespondDecisionTaskFailed_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_RespondDecisionTaskFailed_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_RespondDecisionTaskFailed_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_RespondDecisionTaskFailed_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
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
				v.DomainNotActiveError, err = _DomainNotActiveError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 5:
			if field.Value.Type() == wire.TStruct {
				v.LimitExceededError, err = _LimitExceededError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 6:
			if field.Value.Type() == wire.TStruct {
				v.ServiceBusyError, err = _ServiceBusyError_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	count := 0
	if v.BadRequestError != nil {
		count++
	}
	if v.InternalServiceError != nil {
		count++
	}
	if v.EntityNotExistError != nil {
		count++
	}
	if v.DomainNotActiveError != nil {
		count++
	}
	if v.LimitExceededError != nil {
		count++
	}
	if v.ServiceBusyError != nil {
		count++
	}
	if count > 1 {
		return fmt.Errorf("WorkflowService_RespondDecisionTaskFailed_Result should have at most one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_RespondDecisionTaskFailed_Result
// struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [6]string
	i := 0
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
	if v.DomainNotActiveError != nil {
		fields[i] = fmt.Sprintf("DomainNotActiveError: %v", v.DomainNotActiveError)
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

	return fmt.Sprintf("WorkflowService_RespondDecisionTaskFailed_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_RespondDecisionTaskFailed_Result match the
// provided WorkflowService_RespondDecisionTaskFailed_Result.
//
// This function performs a deep comparison.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) Equals(rhs *WorkflowService_RespondDecisionTaskFailed_Result) bool {
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
		return false
	}
	if !((v.EntityNotExistError == nil && rhs.EntityNotExistError == nil) || (v.EntityNotExistError != nil && rhs.EntityNotExistError != nil && v.EntityNotExistError.Equals(rhs.EntityNotExistError))) {
		return false
	}
	if !((v.DomainNotActiveError == nil && rhs.DomainNotActiveError == nil) || (v.DomainNotActiveError != nil && rhs.DomainNotActiveError != nil && v.DomainNotActiveError.Equals(rhs.DomainNotActiveError))) {
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

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// GetEntityNotExistError returns the value of EntityNotExistError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetEntityNotExistError() (o *shared.EntityNotExistsError) {
	if v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}

	return
}

// GetDomainNotActiveError returns the value of DomainNotActiveError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetDomainNotActiveError() (o *shared.DomainNotActiveError) {
	if v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}

	return
}

// GetLimitExceededError returns the value of LimitExceededError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetLimitExceededError() (o *shared.LimitExceededError) {
	if v.LimitExceededError != nil {
		return v.LimitExceededError
	}

	return
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "RespondDecisionTaskFailed" for this struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) MethodName() string {
	return "RespondDecisionTaskFailed"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *WorkflowService_RespondDecisionTaskFailed_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
