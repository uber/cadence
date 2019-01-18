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

// Code generated by thriftrw v1.15.0. DO NOT EDIT.
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

// HistoryService_SyncActivity_Args represents the arguments for the HistoryService.SyncActivity function.
//
// The arguments for SyncActivity are sent and received over the wire as this struct.
type HistoryService_SyncActivity_Args struct {
	SyncActivityRequest *SyncActivityRequest `json:"syncActivityRequest,omitempty"`
}

// ToWire translates a HistoryService_SyncActivity_Args struct into a Thrift-level intermediate
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
func (v *HistoryService_SyncActivity_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.SyncActivityRequest != nil {
		w, err = v.SyncActivityRequest.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _SyncActivityRequest_Read(w wire.Value) (*SyncActivityRequest, error) {
	var v SyncActivityRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a HistoryService_SyncActivity_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_SyncActivity_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_SyncActivity_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_SyncActivity_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.SyncActivityRequest, err = _SyncActivityRequest_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a HistoryService_SyncActivity_Args
// struct.
func (v *HistoryService_SyncActivity_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.SyncActivityRequest != nil {
		fields[i] = fmt.Sprintf("SyncActivityRequest: %v", v.SyncActivityRequest)
		i++
	}

	return fmt.Sprintf("HistoryService_SyncActivity_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_SyncActivity_Args match the
// provided HistoryService_SyncActivity_Args.
//
// This function performs a deep comparison.
func (v *HistoryService_SyncActivity_Args) Equals(rhs *HistoryService_SyncActivity_Args) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.SyncActivityRequest == nil && rhs.SyncActivityRequest == nil) || (v.SyncActivityRequest != nil && rhs.SyncActivityRequest != nil && v.SyncActivityRequest.Equals(rhs.SyncActivityRequest))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of HistoryService_SyncActivity_Args.
func (v *HistoryService_SyncActivity_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.SyncActivityRequest != nil {
		err = multierr.Append(err, enc.AddObject("syncActivityRequest", v.SyncActivityRequest))
	}
	return err
}

// GetSyncActivityRequest returns the value of SyncActivityRequest if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Args) GetSyncActivityRequest() (o *SyncActivityRequest) {
	if v != nil && v.SyncActivityRequest != nil {
		return v.SyncActivityRequest
	}

	return
}

// IsSetSyncActivityRequest returns true if SyncActivityRequest is not nil.
func (v *HistoryService_SyncActivity_Args) IsSetSyncActivityRequest() bool {
	return v != nil && v.SyncActivityRequest != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "SyncActivity" for this struct.
func (v *HistoryService_SyncActivity_Args) MethodName() string {
	return "SyncActivity"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *HistoryService_SyncActivity_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// HistoryService_SyncActivity_Helper provides functions that aid in handling the
// parameters and return values of the HistoryService.SyncActivity
// function.
var HistoryService_SyncActivity_Helper = struct {
	// Args accepts the parameters of SyncActivity in-order and returns
	// the arguments struct for the function.
	Args func(
		syncActivityRequest *SyncActivityRequest,
	) *HistoryService_SyncActivity_Args

	// IsException returns true if the given error can be thrown
	// by SyncActivity.
	//
	// An error can be thrown by SyncActivity only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for SyncActivity
	// given the error returned by it. The provided error may
	// be nil if SyncActivity did not fail.
	//
	// This allows mapping errors returned by SyncActivity into a
	// serializable result struct. WrapResponse returns a
	// non-nil error if the provided error cannot be thrown by
	// SyncActivity
	//
	//   err := SyncActivity(args)
	//   result, err := HistoryService_SyncActivity_Helper.WrapResponse(err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from SyncActivity: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(error) (*HistoryService_SyncActivity_Result, error)

	// UnwrapResponse takes the result struct for SyncActivity
	// and returns the erorr returned by it (if any).
	//
	// The error is non-nil only if SyncActivity threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   err := HistoryService_SyncActivity_Helper.UnwrapResponse(result)
	UnwrapResponse func(*HistoryService_SyncActivity_Result) error
}{}

func init() {
	HistoryService_SyncActivity_Helper.Args = func(
		syncActivityRequest *SyncActivityRequest,
	) *HistoryService_SyncActivity_Args {
		return &HistoryService_SyncActivity_Args{
			SyncActivityRequest: syncActivityRequest,
		}
	}

	HistoryService_SyncActivity_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		case *ShardOwnershipLostError:
			return true
		case *shared.ServiceBusyError:
			return true
		case *shared.RetryTaskError:
			return true
		default:
			return false
		}
	}

	HistoryService_SyncActivity_Helper.WrapResponse = func(err error) (*HistoryService_SyncActivity_Result, error) {
		if err == nil {
			return &HistoryService_SyncActivity_Result{}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.BadRequestError")
			}
			return &HistoryService_SyncActivity_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.InternalServiceError")
			}
			return &HistoryService_SyncActivity_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.EntityNotExistError")
			}
			return &HistoryService_SyncActivity_Result{EntityNotExistError: e}, nil
		case *ShardOwnershipLostError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.ShardOwnershipLostError")
			}
			return &HistoryService_SyncActivity_Result{ShardOwnershipLostError: e}, nil
		case *shared.ServiceBusyError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.ServiceBusyError")
			}
			return &HistoryService_SyncActivity_Result{ServiceBusyError: e}, nil
		case *shared.RetryTaskError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for HistoryService_SyncActivity_Result.RetryTaskError")
			}
			return &HistoryService_SyncActivity_Result{RetryTaskError: e}, nil
		}

		return nil, err
	}
	HistoryService_SyncActivity_Helper.UnwrapResponse = func(result *HistoryService_SyncActivity_Result) (err error) {
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
		if result.ServiceBusyError != nil {
			err = result.ServiceBusyError
			return
		}
		if result.RetryTaskError != nil {
			err = result.RetryTaskError
			return
		}
		return
	}

}

// HistoryService_SyncActivity_Result represents the result of a HistoryService.SyncActivity function call.
//
// The result of a SyncActivity execution is sent and received over the wire as this struct.
type HistoryService_SyncActivity_Result struct {
	BadRequestError         *shared.BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError    *shared.InternalServiceError `json:"internalServiceError,omitempty"`
	EntityNotExistError     *shared.EntityNotExistsError `json:"entityNotExistError,omitempty"`
	ShardOwnershipLostError *ShardOwnershipLostError     `json:"shardOwnershipLostError,omitempty"`
	ServiceBusyError        *shared.ServiceBusyError     `json:"serviceBusyError,omitempty"`
	RetryTaskError          *shared.RetryTaskError       `json:"retryTaskError,omitempty"`
}

// ToWire translates a HistoryService_SyncActivity_Result struct into a Thrift-level intermediate
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
func (v *HistoryService_SyncActivity_Result) ToWire() (wire.Value, error) {
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
	if v.ShardOwnershipLostError != nil {
		w, err = v.ShardOwnershipLostError.ToWire()
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
	if v.RetryTaskError != nil {
		w, err = v.RetryTaskError.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 6, Value: w}
		i++
	}

	if i > 1 {
		return wire.Value{}, fmt.Errorf("HistoryService_SyncActivity_Result should have at most one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a HistoryService_SyncActivity_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryService_SyncActivity_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryService_SyncActivity_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryService_SyncActivity_Result) FromWire(w wire.Value) error {
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
				v.ShardOwnershipLostError, err = _ShardOwnershipLostError_Read(field.Value)
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
		case 6:
			if field.Value.Type() == wire.TStruct {
				v.RetryTaskError, err = _RetryTaskError_Read(field.Value)
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
	if v.ShardOwnershipLostError != nil {
		count++
	}
	if v.ServiceBusyError != nil {
		count++
	}
	if v.RetryTaskError != nil {
		count++
	}
	if count > 1 {
		return fmt.Errorf("HistoryService_SyncActivity_Result should have at most one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a HistoryService_SyncActivity_Result
// struct.
func (v *HistoryService_SyncActivity_Result) String() string {
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
	if v.ShardOwnershipLostError != nil {
		fields[i] = fmt.Sprintf("ShardOwnershipLostError: %v", v.ShardOwnershipLostError)
		i++
	}
	if v.ServiceBusyError != nil {
		fields[i] = fmt.Sprintf("ServiceBusyError: %v", v.ServiceBusyError)
		i++
	}
	if v.RetryTaskError != nil {
		fields[i] = fmt.Sprintf("RetryTaskError: %v", v.RetryTaskError)
		i++
	}

	return fmt.Sprintf("HistoryService_SyncActivity_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryService_SyncActivity_Result match the
// provided HistoryService_SyncActivity_Result.
//
// This function performs a deep comparison.
func (v *HistoryService_SyncActivity_Result) Equals(rhs *HistoryService_SyncActivity_Result) bool {
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
	if !((v.EntityNotExistError == nil && rhs.EntityNotExistError == nil) || (v.EntityNotExistError != nil && rhs.EntityNotExistError != nil && v.EntityNotExistError.Equals(rhs.EntityNotExistError))) {
		return false
	}
	if !((v.ShardOwnershipLostError == nil && rhs.ShardOwnershipLostError == nil) || (v.ShardOwnershipLostError != nil && rhs.ShardOwnershipLostError != nil && v.ShardOwnershipLostError.Equals(rhs.ShardOwnershipLostError))) {
		return false
	}
	if !((v.ServiceBusyError == nil && rhs.ServiceBusyError == nil) || (v.ServiceBusyError != nil && rhs.ServiceBusyError != nil && v.ServiceBusyError.Equals(rhs.ServiceBusyError))) {
		return false
	}
	if !((v.RetryTaskError == nil && rhs.RetryTaskError == nil) || (v.RetryTaskError != nil && rhs.RetryTaskError != nil && v.RetryTaskError.Equals(rhs.RetryTaskError))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of HistoryService_SyncActivity_Result.
func (v *HistoryService_SyncActivity_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.BadRequestError != nil {
		err = multierr.Append(err, enc.AddObject("badRequestError", v.BadRequestError))
	}
	if v.InternalServiceError != nil {
		err = multierr.Append(err, enc.AddObject("internalServiceError", v.InternalServiceError))
	}
	if v.EntityNotExistError != nil {
		err = multierr.Append(err, enc.AddObject("entityNotExistError", v.EntityNotExistError))
	}
	if v.ShardOwnershipLostError != nil {
		err = multierr.Append(err, enc.AddObject("shardOwnershipLostError", v.ShardOwnershipLostError))
	}
	if v.ServiceBusyError != nil {
		err = multierr.Append(err, enc.AddObject("serviceBusyError", v.ServiceBusyError))
	}
	if v.RetryTaskError != nil {
		err = multierr.Append(err, enc.AddObject("retryTaskError", v.RetryTaskError))
	}
	return err
}

// GetBadRequestError returns the value of BadRequestError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetBadRequestError() (o *shared.BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}

	return
}

// IsSetBadRequestError returns true if BadRequestError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetBadRequestError() bool {
	return v != nil && v.BadRequestError != nil
}

// GetInternalServiceError returns the value of InternalServiceError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetInternalServiceError() (o *shared.InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}

	return
}

// IsSetInternalServiceError returns true if InternalServiceError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetInternalServiceError() bool {
	return v != nil && v.InternalServiceError != nil
}

// GetEntityNotExistError returns the value of EntityNotExistError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetEntityNotExistError() (o *shared.EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}

	return
}

// IsSetEntityNotExistError returns true if EntityNotExistError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetEntityNotExistError() bool {
	return v != nil && v.EntityNotExistError != nil
}

// GetShardOwnershipLostError returns the value of ShardOwnershipLostError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}

	return
}

// IsSetShardOwnershipLostError returns true if ShardOwnershipLostError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetShardOwnershipLostError() bool {
	return v != nil && v.ShardOwnershipLostError != nil
}

// GetServiceBusyError returns the value of ServiceBusyError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetServiceBusyError() (o *shared.ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}

	return
}

// IsSetServiceBusyError returns true if ServiceBusyError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetServiceBusyError() bool {
	return v != nil && v.ServiceBusyError != nil
}

// GetRetryTaskError returns the value of RetryTaskError if it is set or its
// zero value if it is unset.
func (v *HistoryService_SyncActivity_Result) GetRetryTaskError() (o *shared.RetryTaskError) {
	if v != nil && v.RetryTaskError != nil {
		return v.RetryTaskError
	}

	return
}

// IsSetRetryTaskError returns true if RetryTaskError is not nil.
func (v *HistoryService_SyncActivity_Result) IsSetRetryTaskError() bool {
	return v != nil && v.RetryTaskError != nil
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "SyncActivity" for this struct.
func (v *HistoryService_SyncActivity_Result) MethodName() string {
	return "SyncActivity"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *HistoryService_SyncActivity_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
