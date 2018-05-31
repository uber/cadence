// Code generated by thriftrw v1.11.0. DO NOT EDIT.
// @generated

package matching

import (
	"errors"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// MatchingService_DescribeTaskList_Args represents the arguments for the MatchingService.DescribeTaskList function.
//
// The arguments for DescribeTaskList are sent and received over the wire as this struct.
type MatchingService_DescribeTaskList_Args struct {
	Request *DescribeTaskListRequest `json:"request,omitempty"`
}

// ToWire translates a MatchingService_DescribeTaskList_Args struct into a Thrift-level intermediate
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
func (v *MatchingService_DescribeTaskList_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Request != nil {
		w, err = v.Request.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DescribeTaskListRequest_1_Read(w wire.Value) (*DescribeTaskListRequest, error) {
	var v DescribeTaskListRequest
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a MatchingService_DescribeTaskList_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a MatchingService_DescribeTaskList_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v MatchingService_DescribeTaskList_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *MatchingService_DescribeTaskList_Args) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TStruct {
				v.Request, err = _DescribeTaskListRequest_1_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a MatchingService_DescribeTaskList_Args
// struct.
func (v *MatchingService_DescribeTaskList_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Request != nil {
		fields[i] = fmt.Sprintf("Request: %v", v.Request)
		i++
	}

	return fmt.Sprintf("MatchingService_DescribeTaskList_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this MatchingService_DescribeTaskList_Args match the
// provided MatchingService_DescribeTaskList_Args.
//
// This function performs a deep comparison.
func (v *MatchingService_DescribeTaskList_Args) Equals(rhs *MatchingService_DescribeTaskList_Args) bool {
	if !((v.Request == nil && rhs.Request == nil) || (v.Request != nil && rhs.Request != nil && v.Request.Equals(rhs.Request))) {
		return false
	}

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "DescribeTaskList" for this struct.
func (v *MatchingService_DescribeTaskList_Args) MethodName() string {
	return "DescribeTaskList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *MatchingService_DescribeTaskList_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// MatchingService_DescribeTaskList_Helper provides functions that aid in handling the
// parameters and return values of the MatchingService.DescribeTaskList
// function.
var MatchingService_DescribeTaskList_Helper = struct {
	// Args accepts the parameters of DescribeTaskList in-order and returns
	// the arguments struct for the function.
	Args func(
		request *DescribeTaskListRequest,
	) *MatchingService_DescribeTaskList_Args

	// IsException returns true if the given error can be thrown
	// by DescribeTaskList.
	//
	// An error can be thrown by DescribeTaskList only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for DescribeTaskList
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// DescribeTaskList into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by DescribeTaskList
	//
	//   value, err := DescribeTaskList(args)
	//   result, err := MatchingService_DescribeTaskList_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from DescribeTaskList: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*shared.DescribeTaskListResponse, error) (*MatchingService_DescribeTaskList_Result, error)

	// UnwrapResponse takes the result struct for DescribeTaskList
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if DescribeTaskList threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := MatchingService_DescribeTaskList_Helper.UnwrapResponse(result)
	UnwrapResponse func(*MatchingService_DescribeTaskList_Result) (*shared.DescribeTaskListResponse, error)
}{}

func init() {
	MatchingService_DescribeTaskList_Helper.Args = func(
		request *DescribeTaskListRequest,
	) *MatchingService_DescribeTaskList_Args {
		return &MatchingService_DescribeTaskList_Args{
			Request: request,
		}
	}

	MatchingService_DescribeTaskList_Helper.IsException = func(err error) bool {
		switch err.(type) {
		case *shared.BadRequestError:
			return true
		case *shared.InternalServiceError:
			return true
		case *shared.EntityNotExistsError:
			return true
		default:
			return false
		}
	}

	MatchingService_DescribeTaskList_Helper.WrapResponse = func(success *shared.DescribeTaskListResponse, err error) (*MatchingService_DescribeTaskList_Result, error) {
		if err == nil {
			return &MatchingService_DescribeTaskList_Result{Success: success}, nil
		}

		switch e := err.(type) {
		case *shared.BadRequestError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_DescribeTaskList_Result.BadRequestError")
			}
			return &MatchingService_DescribeTaskList_Result{BadRequestError: e}, nil
		case *shared.InternalServiceError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_DescribeTaskList_Result.InternalServiceError")
			}
			return &MatchingService_DescribeTaskList_Result{InternalServiceError: e}, nil
		case *shared.EntityNotExistsError:
			if e == nil {
				return nil, errors.New("WrapResponse received non-nil error type with nil value for MatchingService_DescribeTaskList_Result.EntityNotExistError")
			}
			return &MatchingService_DescribeTaskList_Result{EntityNotExistError: e}, nil
		}

		return nil, err
	}
	MatchingService_DescribeTaskList_Helper.UnwrapResponse = func(result *MatchingService_DescribeTaskList_Result) (success *shared.DescribeTaskListResponse, err error) {
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

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// MatchingService_DescribeTaskList_Result represents the result of a MatchingService.DescribeTaskList function call.
//
// The result of a DescribeTaskList execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type MatchingService_DescribeTaskList_Result struct {
	// Value returned by DescribeTaskList after a successful execution.
	Success              *shared.DescribeTaskListResponse `json:"success,omitempty"`
	BadRequestError      *shared.BadRequestError          `json:"badRequestError,omitempty"`
	InternalServiceError *shared.InternalServiceError     `json:"internalServiceError,omitempty"`
	EntityNotExistError  *shared.EntityNotExistsError     `json:"entityNotExistError,omitempty"`
}

// ToWire translates a MatchingService_DescribeTaskList_Result struct into a Thrift-level intermediate
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
func (v *MatchingService_DescribeTaskList_Result) ToWire() (wire.Value, error) {
	var (
		fields [4]wire.Field
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

	if i != 1 {
		return wire.Value{}, fmt.Errorf("MatchingService_DescribeTaskList_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DescribeTaskListResponse_Read(w wire.Value) (*shared.DescribeTaskListResponse, error) {
	var v shared.DescribeTaskListResponse
	err := v.FromWire(w)
	return &v, err
}

func _EntityNotExistsError_Read(w wire.Value) (*shared.EntityNotExistsError, error) {
	var v shared.EntityNotExistsError
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a MatchingService_DescribeTaskList_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a MatchingService_DescribeTaskList_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v MatchingService_DescribeTaskList_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *MatchingService_DescribeTaskList_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _DescribeTaskListResponse_Read(field.Value)
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
	if count != 1 {
		return fmt.Errorf("MatchingService_DescribeTaskList_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a MatchingService_DescribeTaskList_Result
// struct.
func (v *MatchingService_DescribeTaskList_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [4]string
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

	return fmt.Sprintf("MatchingService_DescribeTaskList_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this MatchingService_DescribeTaskList_Result match the
// provided MatchingService_DescribeTaskList_Result.
//
// This function performs a deep comparison.
func (v *MatchingService_DescribeTaskList_Result) Equals(rhs *MatchingService_DescribeTaskList_Result) bool {
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

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "DescribeTaskList" for this struct.
func (v *MatchingService_DescribeTaskList_Result) MethodName() string {
	return "DescribeTaskList"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *MatchingService_DescribeTaskList_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
