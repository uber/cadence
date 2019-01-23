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

// Code generated by thriftrw v1.13.1. DO NOT EDIT.
// @generated

package history

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/multierr"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap/zapcore"
	"strings"
)

// HistoryService_SignalWithStartWorkflowExecution_Args represents the arguments for the HistoryService.SignalWithStartWorkflowExecution function.
//
// The arguments for SignalWithStartWorkflowExecution are sent and received over the wire as this struct.
type HistoryService_SignalWithStartWorkflowExecution_Args struct {
	SignalWithStartRequest *SignalWithStartWorkflowExecutionRequest `json:"signalWithStartRequest,omitempty"`
}

// ToWire translates a HistoryService_SignalWithStartWorkflowExecution_Args struct into a Thrift-level intermediate
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
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.SignalWithStartRequest != nil {
		w, err = v.SignalWithStartRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _SignalWithStartWorkflowExecutionRequest_1_Read(w wire.Value) (*SignalWithStartWorkflowExecutionRequest, error) {
	var v SignalWithStartWorkflowExecutionRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_SignalWithStartWorkflowExecution_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_SignalWithStartWorkflowExecution_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_SignalWithStartWorkflowExecution_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.SignalWithStartRequest, err = _SignalWithStartWorkflowExecutionRequest_1_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a HistoryService_SignalWithStartWorkflowExecution_Args
// struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.SignalWithStartRequest != nil {
		fields[i] = fmt.Sprintf("SignalWithStartRequest: %v", v.SignalWithStartRequest)
		i++
	}

	return fmt.Sprintf("HistoryService_SignalWithStartWorkflowExecution_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_SignalWithStartWorkflowExecution_Args match the
// provided HistoryService_SignalWithStartWorkflowExecution_Args.
//
// This function performs a deep comparison.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) Equals(rhs *HistoryService_SignalWithStartWorkflowExecution_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.SignalWithStartRequest == nil && rhs.SignalWithStartRequest == nil) || (v.SignalWithStartRequest != nil && rhs.SignalWithStartRequest != nil && v.SignalWithStartRequest.Equals(rhs.SignalWithStartRequest))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of HistoryService_SignalWithStartWorkflowExecution_Args.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.SignalWithStartRequest != nil {
		err = multierr.Append(err, enc.AddObject("signalWithStartRequest", v.SignalWithStartRequest))
	}
	return err
}

// GetSignalWithStartRequest returns the value of SignalWithStartRequest if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) GetSignalWithStartRequest() (o *SignalWithStartWorkflowExecutionRequest) {
	if v.SignalWithStartRequest != nil {
		return v.SignalWithStartRequest
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "SignalWithStartWorkflowExecution" for this struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) MethodName() string {
	return "SignalWithStartWorkflowExecution"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// HistoryService_SignalWithStartWorkflowExecution_Helper provides functions that aid in handling the
// parameters and return values of the HistoryService.SignalWithStartWorkflowExecution
// function.
var HistoryService_SignalWithStartWorkflowExecution_Helper = struct {
	// Args accepts the parameters of SignalWithStartWorkflowExecution in-order and returns
	// the arguments struct for the function.
	Args func(
		signalWithStartRequest *SignalWithStartWorkflowExecutionRequest,
	) *HistoryService_SignalWithStartWorkflowExecution_Args

	// IsException returns true if the given error can be thrown
	// by SignalWithStartWorkflowExecution.
	//
	// An error can be thrown by SignalWithStartWorkflowExecution only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for SignalWithStartWorkflowExecution
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// SignalWithStartWorkflowExecution into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by SignalWithStartWorkflowExecution
	//
	//   value, err := SignalWithStartWorkflowExecution(args)
	//   result, err := HistoryService_SignalWithStartWorkflowExecution_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from SignalWithStartWorkflowExecution: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*shared.StartWorkflowExecutionResponse, error) (*HistoryService_SignalWithStartWorkflowExecution_Result, error)

	// UnwrapResponse takes the result struct for SignalWithStartWorkflowExecution
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if SignalWithStartWorkflowExecution threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := HistoryService_SignalWithStartWorkflowExecution_Helper.UnwrapResponse(result)
	UnwrapResponse func(*HistoryService_SignalWithStartWorkflowExecution_Result) (*shared.StartWorkflowExecutionResponse, error)
}{}

func init() {
	HistoryService_SignalWithStartWorkflowExecution_Helper.Args = func(
		signalWithStartRequest *SignalWithStartWorkflowExecutionRequest,
	) *HistoryService_SignalWithStartWorkflowExecution_Args {
		return &HistoryService_SignalWithStartWorkflowExecution_Args{
			SignalWithStartRequest: signalWithStartRequest,
		}
	}

	HistoryService_SignalWithStartWorkflowExecution_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *ShardOwnershipLostError:
			return true
		case *shared.DomainNotActiveError:
			return true
		case *shared.LimitExceededError:
			return true
		case *shared.ServiceBusyError:
			return true
		case *shared.WorkflowExecutionAlreadyStartedError:
			return true
		default:
			return false
		}
	}

	HistoryService_SignalWithStartWorkflowExecution_Helper.WrapResponse = func(success *shared.StartWorkflowExecutionResponse, err error) (*HistoryService_SignalWithStartWorkflowExecution_Result, error) {
		if err == nil {
			return &HistoryService_SignalWithStartWorkflowExecution_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.BadRequestError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.InternalServiceError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{InternalServiceError: e}, nil
		case *ShardOwnershipLostError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.ShardOwnershipLostError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{ShardOwnershipLostError: e}, nil
		case *shared.DomainNotActiveError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.DomainNotActiveError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{DomainNotActiveError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.LimitExceededError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{LimitExceededError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.ServiceBusyError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{ServiceBusyError: e}, nil
		case *shared.WorkflowExecutionAlreadyStartedError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SignalWithStartWorkflowExecution_Result.WorkflowAlreadyStartedError")
			}
			return &HistoryService_SignalWithStartWorkflowExecution_Result{WorkflowAlreadyStartedError: e}, nil
		}

		return nil, err
	}
	HistoryService_SignalWithStartWorkflowExecution_Helper.UnwrapResponse = func(result *HistoryService_SignalWithStartWorkflowExecution_Result) (success *shared.StartWorkflowExecutionResponse, err error) {
		if result.BadRequestError != nil {
			err = result.BadRequestError
			return
		}
		if result.InternalServiceError != nil {
			err = result.InternalServiceError
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
		if result.ServiceBusyError != nil {
			err = result.ServiceBusyError
			return
		}
		if result.WorkflowAlreadyStartedError != nil {
			err = result.WorkflowAlreadyStartedError
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

// HistoryService_SignalWithStartWorkflowExecution_Result represents the result of a HistoryService.SignalWithStartWorkflowExecution function call.
//
// The result of a SignalWithStartWorkflowExecution execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type HistoryService_SignalWithStartWorkflowExecution_Result struct {
	// Value returned by SignalWithStartWorkflowExecution after a successful execution.
	Success                     *shared.StartWorkflowExecutionResponse       `json:"success,omitempty"`
	BadRequestError             *shared.BadRequestError                      `json:"badRequestError,omitempty"`
	InternalServiceError        *shared.InternalServiceError                 `json:"internalServiceError,omitempty"`
	ShardOwnershipLostError     *ShardOwnershipLostError                     `json:"shardOwnershipLostError,omitempty"`
	DomainNotActiveError        *shared.DomainNotActiveError                 `json:"domainNotActiveError,omitempty"`
	LimitExceededError          *shared.LimitExceededError                   `json:"limitExceededError,omitempty"`
	ServiceBusyError            *shared.ServiceBusyError                     `json:"serviceBusyError,omitempty"`
	WorkflowAlreadyStartedError *shared.WorkflowExecutionAlreadyStartedError `json:"workflowAlreadyStartedError,omitempty"`
}

// ToWire translates a HistoryService_SignalWithStartWorkflowExecution_Result struct into a Thrift-level intermediate
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
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) ToWire() (wire.Value, error) {
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
	if v.ShardOwnershipLostError != nil {
		w, err = v.ShardOwnershipLostError.ToWire()
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
	if v.WorkflowAlreadyStartedError != nil {
		w, err = v.WorkflowAlreadyStartedError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 7, Value: w}
		i++
	}

	if i != 1 {
		return wire.Value{}, fmt.Errorf("HistoryService_SignalWithStartWorkflowExecution_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _StartWorkflowExecutionResponse_Read(w wire.Value) (*shared.StartWorkflowExecutionResponse, error) {
	var v shared.StartWorkflowExecutionResponse
	err := v.FromWire(w)
	return &v, err
}

func _WorkflowExecutionAlreadyStartedError_Read(w wire.Value) (*shared.WorkflowExecutionAlreadyStartedError, error) {
	var v shared.WorkflowExecutionAlreadyStartedError
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_SignalWithStartWorkflowExecution_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_SignalWithStartWorkflowExecution_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_SignalWithStartWorkflowExecution_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _StartWorkflowExecutionResponse_Read(field.Value)
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
				v.ShardOwnershipLostError, err = _ShardOwnershipLostError_Read(field.Value)
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
		case 7:
			if field.Value.Type() == wire.TStruct {
				v.WorkflowAlreadyStartedError, err = _WorkflowExecutionAlreadyStartedError_Read(field.Value)
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
	if v.ShardOwnershipLostError != nil {
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
	if v.WorkflowAlreadyStartedError != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("HistoryService_SignalWithStartWorkflowExecution_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a HistoryService_SignalWithStartWorkflowExecution_Result
// struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) String() string {
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
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}
	if v.WorkflowAlreadyStartedError != nil {
		fields[i] = fmt.Sprintf("WorkflowAlreadyStartedError: %v", v.WorkflowAlreadyStartedError)
		i++
	}

	return fmt.Sprintf("HistoryService_SignalWithStartWorkflowExecution_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_SignalWithStartWorkflowExecution_Result match the
// provided HistoryService_SignalWithStartWorkflowExecution_Result.
//
// This function performs a deep comparison.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) Equals(rhs *HistoryService_SignalWithStartWorkflowExecution_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
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
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}
	if !((v.WorkflowAlreadyStartedError == nil && rhs.WorkflowAlreadyStartedError == nil) || (v.WorkflowAlreadyStartedError != nil && rhs.WorkflowAlreadyStartedError != nil && v.WorkflowAlreadyStartedError.Equals(rhs.WorkflowAlreadyStartedError))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of HistoryService_SignalWithStartWorkflowExecution_Result.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Success != nil {
		err = multierr.Append(err, enc.AddObject("success", v.Success))
	}
	if v.BadRequestError != nil {
		err = multierr.Append(err, enc.AddObject("badRequestError", v.BadRequestError))
	}
	if v.InternalServiceError != nil {
		err = multierr.Append(err, enc.AddObject("internalServiceError", v.InternalServiceError))
	}
	if v.ShardOwnershipLostError != nil {
		err = multierr.Append(err, enc.AddObject("shardOwnershipLostError", v.ShardOwnershipLostError))
	}
	if v.DomainNotActiveError != nil {
		err = multierr.Append(err, enc.AddObject("domainNotActiveError", v.DomainNotActiveError))
	}
	if v.LimitExceededError != nil {
		err = multierr.Append(err, enc.AddObject("limitExceededError", v.LimitExceededError))
	}
	if v.ServiceBusyError != nil {
		err = multierr.Append(err, enc.AddObject("serviceBusyError", v.ServiceBusyError))
	}
	if v.WorkflowAlreadyStartedError != nil {
		err = multierr.Append(err, enc.AddObject("workflowAlreadyStartedError", v.WorkflowAlreadyStartedError))
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetSuccess() (o *shared.StartWorkflowExecutionResponse) {
	if v.Success != nil {
		return v.Success
	}

	return
}

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// GetShardOwnershipLostError returns the value of ShardOwnershipLostError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}

	return
}

// GetDomainNotActiveError returns the value of DomainNotActiveError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetDomainNotActiveError() (o *shared.DomainNotActiveError) {
	if v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}

	return
}

// GetLimitExceededError returns the value of LimitExceededError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetLimitExceededError() (o *shared.LimitExceededError) {
	if v.LimitExceededError != nil {
		return v.LimitExceededError
	}

	return
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// GetWorkflowAlreadyStartedError returns the value of WorkflowAlreadyStartedError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetWorkflowAlreadyStartedError() (o *shared.WorkflowExecutionAlreadyStartedError) {
	if v.WorkflowAlreadyStartedError != nil {
		return v.WorkflowAlreadyStartedError
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "SignalWithStartWorkflowExecution" for this struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) MethodName() string {
	return "SignalWithStartWorkflowExecution"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
