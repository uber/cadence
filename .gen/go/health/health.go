// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Code generated by thriftrw v1.25.0. DO NOT EDIT.
// @generated

package health

import (
	errors "errors"
	fmt "fmt"
	strings "strings"

	multierr "go.uber.org/multierr"
	thriftreflect "go.uber.org/thriftrw/thriftreflect"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"
)

type HealthStatus struct {
	Ok  bool    `json:"ok,required"`
	Msg *string `json:"msg,omitempty"`
}

// ToWire translates a HealthStatus struct into a Thrift-level intermediate
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
func (v *HealthStatus) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueBool(v.Ok), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.Msg != nil {
		w, err = wire.NewValueString(*(v.Msg)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a HealthStatus struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HealthStatus struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HealthStatus
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HealthStatus) FromWire(w wire.Value) error {
	var err error

	okIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBool {
				v.Ok, err = field.Value.GetBool(), error(nil)
				if err != nil {
					return err
				}
				okIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Msg = &x
				if err != nil {
					return err
				}

			}
		}
	}

	if !okIsSet {
		return errors.New("field Ok of HealthStatus is required")
	}

	return nil
}

// String returns a readable string representation of a HealthStatus
// struct.
func (v *HealthStatus) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	fields[i] = fmt.Sprintf("Ok: %v", v.Ok)
	i++
	if v.Msg != nil {
		fields[i] = fmt.Sprintf("Msg: %v", *(v.Msg))
		i++
	}

	return fmt.Sprintf("HealthStatus{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this HealthStatus match the
// provided HealthStatus.
//
// This function performs a deep comparison.
func (v *HealthStatus) Equals(rhs *HealthStatus) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !(v.Ok == rhs.Ok) {
		return false
	}
	if !_String_EqualsPtr(v.Msg, rhs.Msg) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of HealthStatus.
func (v *HealthStatus) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	enc.AddBool("ok", v.Ok)
	if v.Msg != nil {
		enc.AddString("msg", *v.Msg)
	}
	return err
}

// GetOk returns the value of Ok if it is set or its
// zero value if it is unset.
func (v *HealthStatus) GetOk() (o bool) {
	if v != nil {
		o = v.Ok
	}
	return
}

// GetMsg returns the value of Msg if it is set or its
// zero value if it is unset.
func (v *HealthStatus) GetMsg() (o string) {
	if v != nil && v.Msg != nil {
		return *v.Msg
	}

	return
}

// IsSetMsg returns true if Msg is not nil.
func (v *HealthStatus) IsSetMsg() bool {
	return v != nil && v.Msg != nil
}

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "health",
	Package:  "github.com/uber/cadence/.gen/go/health",
	FilePath: "health.thrift",
	SHA1:     "8d52f05c157e47bef27c86d2133e1cdb475f8024",
	Raw:      rawIDL,
}

const rawIDL = "// Copyright (c) 2017 Uber Technologies, Inc.\n//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\nnamespace java com.uber.cadence\n\n/* ==================== Health Check ==================== */\n\nstruct HealthStatus {\n    1: required bool ok\n    2: optional string msg\n}\n\nservice Meta {\n    HealthStatus health()\n}\n\n"

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
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of Meta_Health_Args.
func (v *Meta_Health_Args) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	return err
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
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.Success == nil && rhs.Success == nil) || (v.Success != nil && rhs.Success != nil && v.Success.Equals(rhs.Success))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of Meta_Health_Result.
func (v *Meta_Health_Result) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Success != nil {
		err = multierr.Append(err, enc.AddObject("success", v.Success))
	}
	return err
}

// GetSuccess returns the value of Success if it is set or its
// zero value if it is unset.
func (v *Meta_Health_Result) GetSuccess() (o *HealthStatus) {
	if v != nil && v.Success != nil {
		return v.Success
	}

	return
}

// IsSetSuccess returns true if Success is not nil.
func (v *Meta_Health_Result) IsSetSuccess() bool {
	return v != nil && v.Success != nil
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
