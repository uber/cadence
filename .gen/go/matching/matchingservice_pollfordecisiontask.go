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

package matching

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/multierr"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap/zapcore"
	"strings"
)

// MatchingService_PollForDecisionTask_Args represents the arguments for the MatchingService.PollForDecisionTask function.
//
// The arguments for PollForDecisionTask are sent and received over the wire as this struct.
type MatchingService_PollForDecisionTask_Args struct {
	PollRequest *PollForDecisionTaskRequest `json:"pollRequest,omitempty"`
}

// ToWire translates a MatchingService_PollForDecisionTask_Args struct into a Thrift-level intermediate
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
func (v *MatchingService_PollForDecisionTask_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.PollRequest != nil {
		w, err = v.PollRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _PollForDecisionTaskRequest_1_Read(w wire.Value) (*PollForDecisionTaskRequest, error) {
	var v PollForDecisionTaskRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a MatchingService_PollForDecisionTask_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a MatchingService_PollForDecisionTask_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v MatchingService_PollForDecisionTask_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *MatchingService_PollForDecisionTask_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.PollRequest, err = _PollForDecisionTaskRequest_1_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a MatchingService_PollForDecisionTask_Args
// struct.
func (v *MatchingService_PollForDecisionTask_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.PollRequest != nil {
		fields[i] = fmt.Sprintf("PollRequest: %v", v.PollRequest)
		i++
	}

	return fmt.Sprintf("MatchingService_PollForDecisionTask_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this MatchingService_PollForDecisionTask_Args match the
// provided MatchingService_PollForDecisionTask_Args.
//
// This function performs a deep comparison.
func (v *MatchingService_PollForDecisionTask_Args) Equals(rhs *MatchingService_PollForDecisionTask_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.PollRequest == nil && rhs.PollRequest == nil) || (v.PollRequest != nil && rhs.PollRequest != nil && v.PollRequest.Equals(rhs.PollRequest))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of MatchingService_PollForDecisionTask_Args.
func (v *MatchingService_PollForDecisionTask_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.PollRequest != nil {
		err = multierr.Append(err, enc.AddObject("pollRequest", v.PollRequest))
	}
	return err
}

// GetPollRequest returns the value of PollRequest if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Args) GetPollRequest() (o *PollForDecisionTaskRequest) {
	if v.PollRequest != nil {
		return v.PollRequest
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "PollForDecisionTask" for this struct.
func (v *MatchingService_PollForDecisionTask_Args) MethodName() string {
	return "PollForDecisionTask"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *MatchingService_PollForDecisionTask_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// MatchingService_PollForDecisionTask_Helper provides functions that aid in handling the
// parameters and return values of the MatchingService.PollForDecisionTask
// function.
var MatchingService_PollForDecisionTask_Helper = struct {
	// Args accepts the parameters of PollForDecisionTask in-order and returns
	// the arguments struct for the function.
	Args func(
		pollRequest *PollForDecisionTaskRequest,
	) *MatchingService_PollForDecisionTask_Args

	// IsException returns true if the given error can be thrown
	// by PollForDecisionTask.
	//
	// An error can be thrown by PollForDecisionTask only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for PollForDecisionTask
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// PollForDecisionTask into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by PollForDecisionTask
	//
	//   value, err := PollForDecisionTask(args)
	//   result, err := MatchingService_PollForDecisionTask_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from PollForDecisionTask: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*PollForDecisionTaskResponse, error) (*MatchingService_PollForDecisionTask_Result, error)

	// UnwrapResponse takes the result struct for PollForDecisionTask
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if PollForDecisionTask threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := MatchingService_PollForDecisionTask_Helper.UnwrapResponse(result)
	UnwrapResponse func(*MatchingService_PollForDecisionTask_Result) (*PollForDecisionTaskResponse, error)
}{}

func init() {
	MatchingService_PollForDecisionTask_Helper.Args = func(
		pollRequest *PollForDecisionTaskRequest,
	) *MatchingService_PollForDecisionTask_Args {
		return &MatchingService_PollForDecisionTask_Args{
			PollRequest: pollRequest,
		}
	}

	MatchingService_PollForDecisionTask_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.LimitExceededError:
			return true
		case *shared.ServiceBusyError:
			return true
		default:
			return false
		}
	}

	MatchingService_PollForDecisionTask_Helper.WrapResponse = func(success *PollForDecisionTaskResponse, err error) (*MatchingService_PollForDecisionTask_Result, error) {
		if err == nil {
			return &MatchingService_PollForDecisionTask_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_PollForDecisionTask_Result.BadRequestError")
			}
			return &MatchingService_PollForDecisionTask_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_PollForDecisionTask_Result.InternalServiceError")
			}
			return &MatchingService_PollForDecisionTask_Result{InternalServiceError: e}, nil
		case *shared.LimitExceededError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_PollForDecisionTask_Result.LimitExceededError")
			}
			return &MatchingService_PollForDecisionTask_Result{LimitExceededError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_PollForDecisionTask_Result.ServiceBusyError")
			}
			return &MatchingService_PollForDecisionTask_Result{ServiceBusyError: e}, nil
		}

		return nil, err
	}
	MatchingService_PollForDecisionTask_Helper.UnwrapResponse = func(result *MatchingService_PollForDecisionTask_Result) (success *PollForDecisionTaskResponse, err error) {
		if result.BadRequestError != nil {
			err = result.BadRequestError
			return
		}
		if result.InternalServiceError != nil {
			err = result.InternalServiceError
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

// MatchingService_PollForDecisionTask_Result represents the result of a MatchingService.PollForDecisionTask function call.
//
// The result of a PollForDecisionTask execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type MatchingService_PollForDecisionTask_Result struct {
	// Value returned by PollForDecisionTask after a successful execution.
	Success              *PollForDecisionTaskResponse `json:"success,omitempty"`
	BadRequestError      *shared.BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError `json:"internalServiceError,omitempty"`
	LimitExceededError   *shared.LimitExceededError   `json:"limitExceededError,omitempty"`
	ServiceBusyError     *shared.ServiceBusyError     `json:"serviceBusyError,omitempty"`
}

// ToWire translates a MatchingService_PollForDecisionTask_Result struct into a Thrift-level intermediate
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
func (v *MatchingService_PollForDecisionTask_Result) ToWire() (wire.Value, error) {
	var (
		fields [5]wire.Field
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
	if v.LimitExceededError != nil {
		w, err = v.LimitExceededError.ToWire()
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

	if i != 1 {
		return wire.Value{}, fmt.Errorf("MatchingService_PollForDecisionTask_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _PollForDecisionTaskResponse_Read(w wire.Value) (*PollForDecisionTaskResponse, error) {
	var v PollForDecisionTaskResponse
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a MatchingService_PollForDecisionTask_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a MatchingService_PollForDecisionTask_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v MatchingService_PollForDecisionTask_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *MatchingService_PollForDecisionTask_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _PollForDecisionTaskResponse_Read(field.Value)
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
				v.LimitExceededError, err = _LimitExceededError_Read(field.Value)
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
	if v.LimitExceededError != nil {
		count++
	}
	if v.ServiceBusyError != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("MatchingService_PollForDecisionTask_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a MatchingService_PollForDecisionTask_Result
// struct.
func (v *MatchingService_PollForDecisionTask_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [5]string
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
	if v.LimitExceededError != nil {
		fields[i] = fmt.Sprintf("LimitExceededError: %v", v.LimitExceededError)
		i++
	}
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}

	return fmt.Sprintf("MatchingService_PollForDecisionTask_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this MatchingService_PollForDecisionTask_Result match the
// provided MatchingService_PollForDecisionTask_Result.
//
// This function performs a deep comparison.
func (v *MatchingService_PollForDecisionTask_Result) Equals(rhs *MatchingService_PollForDecisionTask_Result) bool {
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
	if !((v.LimitExceededError == nil && rhs.LimitExceededError == nil) || (v.LimitExceededError != nil && rhs.LimitExceededError != nil && v.LimitExceededError.Equals(rhs.LimitExceededError))) {
		return false
	}
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of MatchingService_PollForDecisionTask_Result.
func (v *MatchingService_PollForDecisionTask_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
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
	if v.LimitExceededError != nil {
		err = multierr.Append(err, enc.AddObject("limitExceededError", v.LimitExceededError))
	}
	if v.ServiceBusyError != nil {
		err = multierr.Append(err, enc.AddObject("serviceBusyError", v.ServiceBusyError))
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Result) GetSuccess() (o *PollForDecisionTaskResponse) {
	if v.Success != nil {
		return v.Success
	}

	return
}

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// GetLimitExceededError returns the value of LimitExceededError if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Result) GetLimitExceededError() (o *shared.LimitExceededError) {
	if v.LimitExceededError != nil {
		return v.LimitExceededError
	}

	return
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *MatchingService_PollForDecisionTask_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "PollForDecisionTask" for this struct.
func (v *MatchingService_PollForDecisionTask_Result) MethodName() string {
	return "PollForDecisionTask"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *MatchingService_PollForDecisionTask_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
