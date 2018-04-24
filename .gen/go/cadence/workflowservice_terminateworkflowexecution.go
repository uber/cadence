// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Code generated by thriftrw v1.10.0. DO NOT EDIT.
// @generated

package cadence

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// WorkflowService_TerminateWorkflowExecution_Args represents the arguments for the WorkflowService.TerminateWorkflowExecution function.
//
// The arguments for TerminateWorkflowExecution are sent and received over the wire as this struct.
type WorkflowService_TerminateWorkflowExecution_Args struct {
	TerminateRequest *shared.TerminateWorkflowExecutionRequest `json:"terminateRequest,omitempty"`
}

// ToWire translates a WorkflowService_TerminateWorkflowExecution_Args struct into a Thrift-level intermediate
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
func (v *WorkflowService_TerminateWorkflowExecution_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.TerminateRequest != nil {
		w, err = v.TerminateRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _TerminateWorkflowExecutionRequest_Read(w wire.Value) (*shared.TerminateWorkflowExecutionRequest, error) {
	var v shared.TerminateWorkflowExecutionRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_TerminateWorkflowExecution_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_TerminateWorkflowExecution_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_TerminateWorkflowExecution_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_TerminateWorkflowExecution_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.TerminateRequest, err = _TerminateWorkflowExecutionRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_TerminateWorkflowExecution_Args
// struct.
func (v *WorkflowService_TerminateWorkflowExecution_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.TerminateRequest != nil {
		fields[i] = fmt.Sprintf("TerminateRequest: %v", v.TerminateRequest)
		i++
	}

	return fmt.Sprintf("WorkflowService_TerminateWorkflowExecution_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_TerminateWorkflowExecution_Args match the
// provided WorkflowService_TerminateWorkflowExecution_Args.
//
// This function performs a deep comparison.
func (v *WorkflowService_TerminateWorkflowExecution_Args) Equals(rhs *WorkflowService_TerminateWorkflowExecution_Args) bool {
	if !((v.TerminateRequest == nil && rhs.TerminateRequest == nil) || (v.TerminateRequest != nil && rhs.TerminateRequest != nil && v.TerminateRequest.Equals(rhs.TerminateRequest))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "TerminateWorkflowExecution" for this struct.
func (v *WorkflowService_TerminateWorkflowExecution_Args) MethodName() string {
	return "TerminateWorkflowExecution"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *WorkflowService_TerminateWorkflowExecution_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// WorkflowService_TerminateWorkflowExecution_Helper provides functions that aid in handling the
// parameters and return values of the WorkflowService.TerminateWorkflowExecution
// function.
var WorkflowService_TerminateWorkflowExecution_Helper = struct {
	// Args accepts the parameters of TerminateWorkflowExecution in-order and returns
	// the arguments struct for the function.
	Args func(
		terminateRequest *shared.TerminateWorkflowExecutionRequest,
	) *WorkflowService_TerminateWorkflowExecution_Args

	// IsException returns true if the given error can be thrown
	// by TerminateWorkflowExecution.
	//
	// An error can be thrown by TerminateWorkflowExecution only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for TerminateWorkflowExecution
	// given the error returned by it. The provided error may
	// be nil if TerminateWorkflowExecution did not fail.
	//
	// This allows mapping errors returned by TerminateWorkflowExecution into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// TerminateWorkflowExecution
	//
	//   err := TerminateWorkflowExecution(args)
	//   result, err := WorkflowService_TerminateWorkflowExecution_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from TerminateWorkflowExecution: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*WorkflowService_TerminateWorkflowExecution_Result, error)

	// UnwrapResponse takes the result struct for TerminateWorkflowExecution
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if TerminateWorkflowExecution threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := WorkflowService_TerminateWorkflowExecution_Helper.UnwrapResponse(result)
	UnwrapResponse func(*WorkflowService_TerminateWorkflowExecution_Result) error
}{}

func init() {
	WorkflowService_TerminateWorkflowExecution_Helper.Args = func(
		terminateRequest *shared.TerminateWorkflowExecutionRequest,
	) *WorkflowService_TerminateWorkflowExecution_Args {
		return &WorkflowService_TerminateWorkflowExecution_Args{
			TerminateRequest: terminateRequest,
		}
	}

	WorkflowService_TerminateWorkflowExecution_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		case *shared.ServiceBusyError:
			return true
		case *shared.DomainNotActiveError:
			return true
		default:
			return false
		}
	}

	WorkflowService_TerminateWorkflowExecution_Helper.WrapResponse = func(err error) (*WorkflowService_TerminateWorkflowExecution_Result, error) {
		if err == nil {
			return &WorkflowService_TerminateWorkflowExecution_Result{}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_TerminateWorkflowExecution_Result.BadRequestError")
			}
			return &WorkflowService_TerminateWorkflowExecution_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_TerminateWorkflowExecution_Result.InternalServiceError")
			}
			return &WorkflowService_TerminateWorkflowExecution_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_TerminateWorkflowExecution_Result.EntityNotExistError")
			}
			return &WorkflowService_TerminateWorkflowExecution_Result{EntityNotExistError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_TerminateWorkflowExecution_Result.ServiceBusyError")
			}
			return &WorkflowService_TerminateWorkflowExecution_Result{ServiceBusyError: e}, nil
		case *shared.DomainNotActiveError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_TerminateWorkflowExecution_Result.DomainNotActiveError")
			}
			return &WorkflowService_TerminateWorkflowExecution_Result{DomainNotActiveError: e}, nil
		}

		return nil, err
	}
	WorkflowService_TerminateWorkflowExecution_Helper.UnwrapResponse = func(result *WorkflowService_TerminateWorkflowExecution_Result) (err error) {
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
		if result.ServiceBusyError != nil {
			err = result.ServiceBusyError
			return
		}
		if result.DomainNotActiveError != nil {
			err = result.DomainNotActiveError
			return
		}
		return
	}

}

// WorkflowService_TerminateWorkflowExecution_Result represents the result of a WorkflowService.TerminateWorkflowExecution function call.
//
// The result of a TerminateWorkflowExecution execution is sent and received over the wire as this struct.
type WorkflowService_TerminateWorkflowExecution_Result struct {
	BadRequestError      *shared.BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError `json:"internalServiceError,omitempty"`
	EntityNotExistError  *shared.EntityNotExistsError `json:"entityNotExistError,omitempty"`
	ServiceBusyError     *shared.ServiceBusyError     `json:"serviceBusyError,omitempty"`
	DomainNotActiveError *shared.DomainNotActiveError `json:"domainNotActiveError,omitempty"`
}

// ToWire translates a WorkflowService_TerminateWorkflowExecution_Result struct into a Thrift-level intermediate
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
func (v *WorkflowService_TerminateWorkflowExecution_Result) ToWire() (wire.Value, error) {
	var (
		fields [5]wire.Field
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
	if v.ServiceBusyError != nil {
		w, err = v.ServiceBusyError.ToWire()
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

	if i > 1 {
		return wire.Value{}, fmt.Errorf("WorkflowService_TerminateWorkflowExecution_Result should have at most one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a WorkflowService_TerminateWorkflowExecution_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_TerminateWorkflowExecution_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_TerminateWorkflowExecution_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_TerminateWorkflowExecution_Result) FromWire(w wire.Value) error {
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
				v.ServiceBusyError, err = _ServiceBusyError_Read(field.Value)
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
	if v.ServiceBusyError != nil {
		count++
	}
	if v.DomainNotActiveError != nil {
		count++
	}
	if count > 1 {
		return fmt.Errorf("WorkflowService_TerminateWorkflowExecution_Result should have at most one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_TerminateWorkflowExecution_Result
// struct.
func (v *WorkflowService_TerminateWorkflowExecution_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [5]string
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
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}
	if v.DomainNotActiveError != nil {
		fields[i] = fmt.Sprintf("DomainNotActiveError: %v", v.DomainNotActiveError)
		i++
	}

	return fmt.Sprintf("WorkflowService_TerminateWorkflowExecution_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_TerminateWorkflowExecution_Result match the
// provided WorkflowService_TerminateWorkflowExecution_Result.
//
// This function performs a deep comparison.
func (v *WorkflowService_TerminateWorkflowExecution_Result) Equals(rhs *WorkflowService_TerminateWorkflowExecution_Result) bool {
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
		return false
	}
	if !((v.EntityNotExistError == nil && rhs.EntityNotExistError == nil) || (v.EntityNotExistError != nil && rhs.EntityNotExistError != nil && v.EntityNotExistError.Equals(rhs.EntityNotExistError))) {
		return false
	}
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}
	if !((v.DomainNotActiveError == nil && rhs.DomainNotActiveError == nil) || (v.DomainNotActiveError != nil && rhs.DomainNotActiveError != nil && v.DomainNotActiveError.Equals(rhs.DomainNotActiveError))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "TerminateWorkflowExecution" for this struct.
func (v *WorkflowService_TerminateWorkflowExecution_Result) MethodName() string {
	return "TerminateWorkflowExecution"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *WorkflowService_TerminateWorkflowExecution_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
