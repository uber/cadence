// Code generated by thriftrw v1.12.0. DO NOT EDIT.
// @generated

package health

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

// Meta_Health_Args represents the arguments for the Meta.health function.
//
// The arguments for health are sent and received over the wire as this struct.
type Meta_Health_Args struct {
}

// ToWire translates a Meta_Health_Args struct into a Thrift-level intermediate
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
func (v *Meta_Health_Args) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a Meta_Health_Args struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Meta_Health_Args struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Meta_Health_Args
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Meta_Health_Args) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a Meta_Health_Args
// struct.
func (v *Meta_Health_Args) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("Meta_Health_Args{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Meta_Health_Args match the
// provided Meta_Health_Args.
//
// This function performs a deep comparison.
func (v *Meta_Health_Args) Equals(rhs *Meta_Health_Args) bool {

	return true
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the arguments.
//
// This will always be "health" for this struct.
func (v *Meta_Health_Args) MethodName() string {
	return "health"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Call for this struct.
func (v *Meta_Health_Args) EnvelopeType() wire.EnvelopeType {
	return wire.Call
}

// Meta_Health_Helper provides functions that aid in handling the
// parameters and return values of the Meta.health
// function.
var Meta_Health_Helper = struct {
	// Args accepts the parameters of health in-order and returns
	// the arguments struct for the function.
	Args func() *Meta_Health_Args

	// IsException returns true if the given error can be thrown
	// by health.
	//
	// An error can be thrown by health only if the
	// corresponding exception type was mentioned in the 'throws'
	// section for it in the Thrift file.
	IsException func(error) bool

	// WrapResponse returns the result struct for health
	// given its return value and error.
	//
	// This allows mapping values and errors returned by
	// health into a serializable result struct.
	// WrapResponse returns a non-nil error if the provided
	// error cannot be thrown by health
	//
	//   value, err := health(args)
	//   result, err := Meta_Health_Helper.WrapResponse(value, err)
	//   if err != nil {
	//     return fmt.Errorf("unexpected error from health: %v", err)
	//   }
	//   serialize(result)
	WrapResponse func(*HealthStatus, error) (*Meta_Health_Result, error)

	// UnwrapResponse takes the result struct for health
	// and returns the value or error returned by it.
	//
	// The error is non-nil only if health threw an
	// exception.
	//
	//   result := deserialize(bytes)
	//   value, err := Meta_Health_Helper.UnwrapResponse(result)
	UnwrapResponse func(*Meta_Health_Result) (*HealthStatus, error)
}{}

func init() {
	Meta_Health_Helper.Args = func() *Meta_Health_Args {
		return &Meta_Health_Args{}
	}

	Meta_Health_Helper.IsException = func(err error) bool {
		switch err.(type) {
		default:
			return false
		}
	}

	Meta_Health_Helper.WrapResponse = func(success *HealthStatus, err error) (*Meta_Health_Result, error) {
		if err == nil {
			return &Meta_Health_Result{Success: success}, nil
		}

		return nil, err
	}
	Meta_Health_Helper.UnwrapResponse = func(result *Meta_Health_Result) (success *HealthStatus, err error) {

		if result.Success != nil {
			success = result.Success
			return
		}

		err = errors.New("expected a non-void result")
		return
	}

}

// Meta_Health_Result represents the result of a Meta.health function call.
//
// The result of a health execution is sent and received over the wire as this struct.
//
// Success is set only if the function did not throw an exception.
type Meta_Health_Result struct {
	// Value returned by health after a successful execution.
	Success *HealthStatus `json:"success,omitempty"`
}

// ToWire translates a Meta_Health_Result struct into a Thrift-level intermediate
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
func (v *Meta_Health_Result) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
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

	if i != 1 {
		return wire.Value{}, fmt.Errorf("Meta_Health_Result should have exactly one field: got %v fields", i)
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _HealthStatus_Read(w wire.Value) (*HealthStatus, error) {
	var v HealthStatus
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a Meta_Health_Result struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a Meta_Health_Result struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v Meta_Health_Result
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *Meta_Health_Result) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 0:
			if field.Value.Type() == wire.TStruct {
				v.Success, err = _HealthStatus_Read(field.Value)
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
	if count != 1 {
		return fmt.Errorf("Meta_Health_Result should have exactly one field: got %v fields", count)
	}

	return nil
}

// String returns a readable string representation of a Meta_Health_Result
// struct.
func (v *Meta_Health_Result) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [1]string
	i := 0
	if v.Success != nil {
		fields[i] = fmt.Sprintf("Success: %v", v.Success)
		i++
	}

	return fmt.Sprintf("Meta_Health_Result{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this Meta_Health_Result match the
// provided Meta_Health_Result.
//
// This function performs a deep comparison.
func (v *Meta_Health_Result) Equals(rhs *Meta_Health_Result) bool {
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}

	return true
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *Meta_Health_Result) GetSuccess() (o *HealthStatus) {
	if v.Success != nil {
		return v.Success
	}

	return
}

// MethodName returns the name of the Thrift function as specified in
// the IDL, for which this struct represent the result.
//
// This will always be "health" for this struct.
func (v *Meta_Health_Result) MethodName() string {
	return "health"
}

// EnvelopeType returns the kind of value inside this struct.
//
// This will always be Reply for this struct.
func (v *Meta_Health_Result) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}
