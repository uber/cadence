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

package cadence

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/multierr"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap/zapcore"
	"strings"
)

// WorkflowService_RegisterDomain_Args represents the arguments for the WorkflowService.RegisterDomain function.
//
// The arguments for RegisterDomain are sent and received over the wire as this struct.
type WorkflowService_RegisterDomain_Args struct {
	RegisterRequest *shared.RegisterDomainRequest `json:"registerRequest,omitempty"`
}

// ToWire translates a WorkflowService_RegisterDomain_Args struct into a Thrift-level intermediate
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
func (v *WorkflowService_RegisterDomain_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.RegisterRequest != nil {
		w, err = v.RegisterRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _RegisterDomainRequest_Read(w wire.Value) (*shared.RegisterDomainRequest, error) {
	var v shared.RegisterDomainRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_RegisterDomain_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_RegisterDomain_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_RegisterDomain_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_RegisterDomain_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.RegisterRequest, err = _RegisterDomainRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_RegisterDomain_Args
// struct.
func (v *WorkflowService_RegisterDomain_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.RegisterRequest != nil {
		fields[i] = fmt.Sprintf("RegisterRequest: %v", v.RegisterRequest)
		i++
	}

	return fmt.Sprintf("WorkflowService_RegisterDomain_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_RegisterDomain_Args match the
// provided WorkflowService_RegisterDomain_Args.
//
// This function performs a deep comparison.
func (v *WorkflowService_RegisterDomain_Args) Equals(rhs *WorkflowService_RegisterDomain_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.RegisterRequest == nil && rhs.RegisterRequest == nil) || (v.RegisterRequest != nil && rhs.RegisterRequest != nil && v.RegisterRequest.Equals(rhs.RegisterRequest))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of WorkflowService_RegisterDomain_Args.
func (v *WorkflowService_RegisterDomain_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.RegisterRequest != nil {
		err = multierr.Append(err, enc.AddObject("registerRequest", v.RegisterRequest))
	}
	return err
}

// GetRegisterRequest returns the value of RegisterRequest if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RegisterDomain_Args) GetRegisterRequest() (o *shared.RegisterDomainRequest) {
	if v.RegisterRequest != nil {
		return v.RegisterRequest
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "RegisterDomain" for this struct.
func (v *WorkflowService_RegisterDomain_Args) MethodName() string {
	return "RegisterDomain"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *WorkflowService_RegisterDomain_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// WorkflowService_RegisterDomain_Helper provides functions that aid in handling the
// parameters and return values of the WorkflowService.RegisterDomain
// function.
var WorkflowService_RegisterDomain_Helper = struct {
	// Args accepts the parameters of RegisterDomain in-order and returns
	// the arguments struct for the function.
	Args func(
		registerRequest *shared.RegisterDomainRequest,
	) *WorkflowService_RegisterDomain_Args

	// IsException returns true if the given error can be thrown
	// by RegisterDomain.
	//
	// An error can be thrown by RegisterDomain only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for RegisterDomain
	// given the error returned by it. The provided error may
	// be nil if RegisterDomain did not fail.
	//
	// This allows mapping errors returned by RegisterDomain into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// RegisterDomain
	//
	//   err := RegisterDomain(args)
	//   result, err := WorkflowService_RegisterDomain_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from RegisterDomain: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*WorkflowService_RegisterDomain_Result, error)

	// UnwrapResponse takes the result struct for RegisterDomain
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if RegisterDomain threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := WorkflowService_RegisterDomain_Helper.UnwrapResponse(result)
	UnwrapResponse func(*WorkflowService_RegisterDomain_Result) error
}{}

func init() {
	WorkflowService_RegisterDomain_Helper.Args = func(
		registerRequest *shared.RegisterDomainRequest,
	) *WorkflowService_RegisterDomain_Args {
		return &WorkflowService_RegisterDomain_Args{
			RegisterRequest: registerRequest,
		}
	}

	WorkflowService_RegisterDomain_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.DomainAlreadyExistsError:
			return true
		case *shared.ServiceBusyError:
			return true
		default:
			return false
		}
	}

	WorkflowService_RegisterDomain_Helper.WrapResponse = func(err error) (*WorkflowService_RegisterDomain_Result, error) {
		if err == nil {
			return &WorkflowService_RegisterDomain_Result{}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RegisterDomain_Result.BadRequestError")
			}
			return &WorkflowService_RegisterDomain_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RegisterDomain_Result.InternalServiceError")
			}
			return &WorkflowService_RegisterDomain_Result{InternalServiceError: e}, nil
		case *shared.DomainAlreadyExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RegisterDomain_Result.DomainExistsError")
			}
			return &WorkflowService_RegisterDomain_Result{DomainExistsError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for WorkflowService_RegisterDomain_Result.ServiceBusyError")
			}
			return &WorkflowService_RegisterDomain_Result{ServiceBusyError: e}, nil
		}

		return nil, err
	}
	WorkflowService_RegisterDomain_Helper.UnwrapResponse = func(result *WorkflowService_RegisterDomain_Result) (err error) {
		if result.BadRequestError != nil {
			err = result.BadRequestError
			return
		}
		if result.InternalServiceError != nil {
			err = result.InternalServiceError
			return
		}
		if result.DomainExistsError != nil {
			err = result.DomainExistsError
			return
		}
		if result.ServiceBusyError != nil {
			err = result.ServiceBusyError
			return
		}
		return
	}

}

// WorkflowService_RegisterDomain_Result represents the result of a WorkflowService.RegisterDomain function call.
//
// The result of a RegisterDomain execution is sent and received over the wire as this struct.
type WorkflowService_RegisterDomain_Result struct {
	BadRequestError      *shared.BadRequestError          `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError     `json:"internalServiceError,omitempty"`
	DomainExistsError    *shared.DomainAlreadyExistsError `json:"domainExistsError,omitempty"`
	ServiceBusyError     *shared.ServiceBusyError         `json:"serviceBusyError,omitempty"`
}

// ToWire translates a WorkflowService_RegisterDomain_Result struct into a Thrift-level intermediate
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
func (v *WorkflowService_RegisterDomain_Result) ToWire() (wire.Value, error) {
	var (
		fields [4]wire.Field
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
	if v.DomainExistsError != nil {
		w, err = v.DomainExistsError.ToWire()
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

	if i > 1 {
		return wire.Value{}, fmt.Errorf("WorkflowService_RegisterDomain_Result should have at most one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DomainAlreadyExistsError_Read(w wire.Value) (*shared.DomainAlreadyExistsError, error) {
	var v shared.DomainAlreadyExistsError
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a WorkflowService_RegisterDomain_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a WorkflowService_RegisterDomain_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v WorkflowService_RegisterDomain_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *WorkflowService_RegisterDomain_Result) FromWire(w wire.Value) error {
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
				v.DomainExistsError, err = _DomainAlreadyExistsError_Read(field.Value)
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
	if v.BadRequestError != nil {
		count++
	}
	if v.InternalServiceError != nil {
		count++
	}
	if v.DomainExistsError != nil {
		count++
	}
	if v.ServiceBusyError != nil {
		count++
	}
	if count > 1 {
		return fmt.Errorf("WorkflowService_RegisterDomain_Result should have at most one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a WorkflowService_RegisterDomain_Result
// struct.
func (v *WorkflowService_RegisterDomain_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [4]string
	i := 0
	if v.BadRequestError != nil {
		fields[i] = fmt.Sprintf("BadRequestError: %v", v.BadRequestError)
		i++
	}
	if v.InternalServiceError != nil {
		fields[i] = fmt.Sprintf("InternalServiceError: %v", v.InternalServiceError)
		i++
	}
	if v.DomainExistsError != nil {
		fields[i] = fmt.Sprintf("DomainExistsError: %v", v.DomainExistsError)
		i++
	}
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}

	return fmt.Sprintf("WorkflowService_RegisterDomain_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this WorkflowService_RegisterDomain_Result match the
// provided WorkflowService_RegisterDomain_Result.
//
// This function performs a deep comparison.
func (v *WorkflowService_RegisterDomain_Result) Equals(rhs *WorkflowService_RegisterDomain_Result) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.BadRequestError == nil && rhs.BadRequestError == nil) || (v.BadRequestError != nil && rhs.BadRequestError != nil && v.BadRequestError.Equals(rhs.BadRequestError))) {
		return false
	}
	if !((v.InternalServiceError == nil && rhs.InternalServiceError == nil) || (v.InternalServiceError != nil && rhs.InternalServiceError != nil && v.InternalServiceError.Equals(rhs.InternalServiceError))) {
		return false
	}
	if !((v.DomainExistsError == nil && rhs.DomainExistsError == nil) || (v.DomainExistsError != nil && rhs.DomainExistsError != nil && v.DomainExistsError.Equals(rhs.DomainExistsError))) {
		return false
	}
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of WorkflowService_RegisterDomain_Result.
func (v *WorkflowService_RegisterDomain_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.BadRequestError != nil {
		err = multierr.Append(err, enc.AddObject("badRequestError", v.BadRequestError))
	}
	if v.InternalServiceError != nil {
		err = multierr.Append(err, enc.AddObject("internalServiceError", v.InternalServiceError))
	}
	if v.DomainExistsError != nil {
		err = multierr.Append(err, enc.AddObject("domainExistsError", v.DomainExistsError))
	}
	if v.ServiceBusyError != nil {
		err = multierr.Append(err, enc.AddObject("serviceBusyError", v.ServiceBusyError))
	}
	return err
}

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RegisterDomain_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RegisterDomain_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// GetDomainExistsError returns the value of DomainExistsError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RegisterDomain_Result) GetDomainExistsError() (o *shared.DomainAlreadyExistsError) {
	if v.DomainExistsError != nil {
		return v.DomainExistsError
	}

	return
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *WorkflowService_RegisterDomain_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "RegisterDomain" for this struct.
func (v *WorkflowService_RegisterDomain_Result) MethodName() string {
	return "RegisterDomain"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *WorkflowService_RegisterDomain_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
