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

// Code generated by thriftrw v1.29.2. DO NOT EDIT.
// @generated

package config

import (
	fmt "fmt"
	strings "strings"

	multierr "go.uber.org/multierr"
	stream "go.uber.org/thriftrw/protocol/stream"
	thriftreflect "go.uber.org/thriftrw/thriftreflect"
	wire "go.uber.org/thriftrw/wire"
	zapcore "go.uber.org/zap/zapcore"

	shared "github.com/uber/cadence/.gen/go/shared"
)

type DynamicConfigBlob struct {
	SchemaVersion *int64                `json:"schemaVersion,omitempty"`
	Entries       []*DynamicConfigEntry `json:"entries,omitempty"`
}

type _List_DynamicConfigEntry_ValueList []*DynamicConfigEntry

func (v _List_DynamicConfigEntry_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigEntry', index [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_DynamicConfigEntry_ValueList) Size() int {
	return len(v)
}

func (_List_DynamicConfigEntry_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_DynamicConfigEntry_ValueList) Close() {}

// ToWire translates a DynamicConfigBlob struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//	x, err := v.ToWire()
//	if err != nil {
//	  return err
//	}
//
//	if err := binaryProtocol.Encode(x, writer); err != nil {
//	  return err
//	}
func (v *DynamicConfigBlob) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.SchemaVersion != nil {
		w, err = wire.NewValueI64(*(v.SchemaVersion)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.Entries != nil {
		w, err = wire.NewValueList(_List_DynamicConfigEntry_ValueList(v.Entries)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DynamicConfigEntry_Read(w wire.Value) (*DynamicConfigEntry, error) {
	var v DynamicConfigEntry
	err := v.FromWire(w)
	return &v, err
}

func _List_DynamicConfigEntry_Read(l wire.ValueList) ([]*DynamicConfigEntry, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*DynamicConfigEntry, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _DynamicConfigEntry_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a DynamicConfigBlob struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a DynamicConfigBlob struct
// from the provided intermediate representation.
//
//	x, err := binaryProtocol.Decode(reader, wire.TStruct)
//	if err != nil {
//	  return nil, err
//	}
//
//	var v DynamicConfigBlob
//	if err := v.FromWire(x); err != nil {
//	  return nil, err
//	}
//	return &v, nil
func (v *DynamicConfigBlob) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TI64 {
				var x int64
				x, err = field.Value.GetI64(), error(nil)
				v.SchemaVersion = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TList {
				v.Entries, err = _List_DynamicConfigEntry_Read(field.Value.GetList())
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

func _List_DynamicConfigEntry_Encode(val []*DynamicConfigEntry, sw stream.Writer) error {

	lh := stream.ListHeader{
		Type:   wire.TStruct,
		Length: len(val),
	}
	if err := sw.WriteListBegin(lh); err != nil {
		return err
	}

	for i, v := range val {
		if v == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigEntry', index [%v]: value is nil", i)
		}
		if err := v.Encode(sw); err != nil {
			return err
		}
	}
	return sw.WriteListEnd()
}

// Encode serializes a DynamicConfigBlob struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a DynamicConfigBlob struct could not be encoded.
func (v *DynamicConfigBlob) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	if v.SchemaVersion != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 10, Type: wire.TI64}); err != nil {
			return err
		}
		if err := sw.WriteInt64(*(v.SchemaVersion)); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	if v.Entries != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 20, Type: wire.TList}); err != nil {
			return err
		}
		if err := _List_DynamicConfigEntry_Encode(v.Entries, sw); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	return sw.WriteStructEnd()
}

func _DynamicConfigEntry_Decode(sr stream.Reader) (*DynamicConfigEntry, error) {
	var v DynamicConfigEntry
	err := v.Decode(sr)
	return &v, err
}

func _List_DynamicConfigEntry_Decode(sr stream.Reader) ([]*DynamicConfigEntry, error) {
	lh, err := sr.ReadListBegin()
	if err != nil {
		return nil, err
	}

	if lh.Type != wire.TStruct {
		for i := 0; i < lh.Length; i++ {
			if err := sr.Skip(lh.Type); err != nil {
				return nil, err
			}
		}
		return nil, sr.ReadListEnd()
	}

	o := make([]*DynamicConfigEntry, 0, lh.Length)
	for i := 0; i < lh.Length; i++ {
		v, err := _DynamicConfigEntry_Decode(sr)
		if err != nil {
			return nil, err
		}
		o = append(o, v)
	}

	if err = sr.ReadListEnd(); err != nil {
		return nil, err
	}
	return o, err
}

// Decode deserializes a DynamicConfigBlob struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a DynamicConfigBlob struct could not be generated from the wire
// representation.
func (v *DynamicConfigBlob) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		case fh.ID == 10 && fh.Type == wire.TI64:
			var x int64
			x, err = sr.ReadInt64()
			v.SchemaVersion = &x
			if err != nil {
				return err
			}

		case fh.ID == 20 && fh.Type == wire.TList:
			v.Entries, err = _List_DynamicConfigEntry_Decode(sr)
			if err != nil {
				return err
			}

		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a DynamicConfigBlob
// struct.
func (v *DynamicConfigBlob) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.SchemaVersion != nil {
		fields[i] = fmt.Sprintf("SchemaVersion: %v", *(v.SchemaVersion))
		i++
	}
	if v.Entries != nil {
		fields[i] = fmt.Sprintf("Entries: %v", v.Entries)
		i++
	}

	return fmt.Sprintf("DynamicConfigBlob{%v}", strings.Join(fields[:i], ", "))
}

func _I64_EqualsPtr(lhs, rhs *int64) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _List_DynamicConfigEntry_Equals(lhs, rhs []*DynamicConfigEntry) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this DynamicConfigBlob match the
// provided DynamicConfigBlob.
//
// This function performs a deep comparison.
func (v *DynamicConfigBlob) Equals(rhs *DynamicConfigBlob) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !_I64_EqualsPtr(v.SchemaVersion, rhs.SchemaVersion) {
		return false
	}
	if !((v.Entries == nil && rhs.Entries == nil) || (v.Entries != nil && rhs.Entries != nil && _List_DynamicConfigEntry_Equals(v.Entries, rhs.Entries))) {
		return false
	}

	return true
}

type _List_DynamicConfigEntry_Zapper []*DynamicConfigEntry

// MarshalLogArray implements zapcore.ArrayMarshaler, enabling
// fast logging of _List_DynamicConfigEntry_Zapper.
func (l _List_DynamicConfigEntry_Zapper) MarshalLogArray(enc zapcore.ArrayEncoder) (err error) {
	for _, v := range l {
		err = multierr.Append(err, enc.AppendObject(v))
	}
	return err
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of DynamicConfigBlob.
func (v *DynamicConfigBlob) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.SchemaVersion != nil {
		enc.AddInt64("schemaVersion", *v.SchemaVersion)
	}
	if v.Entries != nil {
		err = multierr.Append(err, enc.AddArray("entries", (_List_DynamicConfigEntry_Zapper)(v.Entries)))
	}
	return err
}

// GetSchemaVersion returns the value of SchemaVersion if it is set or its
// zero value if it is unset.
func (v *DynamicConfigBlob) GetSchemaVersion() (o int64) {
	if v != nil && v.SchemaVersion != nil {
		return *v.SchemaVersion
	}

	return
}

// IsSetSchemaVersion returns true if SchemaVersion is not nil.
func (v *DynamicConfigBlob) IsSetSchemaVersion() bool {
	return v != nil && v.SchemaVersion != nil
}

// GetEntries returns the value of Entries if it is set or its
// zero value if it is unset.
func (v *DynamicConfigBlob) GetEntries() (o []*DynamicConfigEntry) {
	if v != nil && v.Entries != nil {
		return v.Entries
	}

	return
}

// IsSetEntries returns true if Entries is not nil.
func (v *DynamicConfigBlob) IsSetEntries() bool {
	return v != nil && v.Entries != nil
}

type DynamicConfigEntry struct {
	Name   *string               `json:"name,omitempty"`
	Values []*DynamicConfigValue `json:"values,omitempty"`
}

type _List_DynamicConfigValue_ValueList []*DynamicConfigValue

func (v _List_DynamicConfigValue_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigValue', index [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_DynamicConfigValue_ValueList) Size() int {
	return len(v)
}

func (_List_DynamicConfigValue_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_DynamicConfigValue_ValueList) Close() {}

// ToWire translates a DynamicConfigEntry struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//	x, err := v.ToWire()
//	if err != nil {
//	  return err
//	}
//
//	if err := binaryProtocol.Encode(x, writer); err != nil {
//	  return err
//	}
func (v *DynamicConfigEntry) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Name != nil {
		w, err = wire.NewValueString(*(v.Name)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.Values != nil {
		w, err = wire.NewValueList(_List_DynamicConfigValue_ValueList(v.Values)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DynamicConfigValue_Read(w wire.Value) (*DynamicConfigValue, error) {
	var v DynamicConfigValue
	err := v.FromWire(w)
	return &v, err
}

func _List_DynamicConfigValue_Read(l wire.ValueList) ([]*DynamicConfigValue, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*DynamicConfigValue, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _DynamicConfigValue_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a DynamicConfigEntry struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a DynamicConfigEntry struct
// from the provided intermediate representation.
//
//	x, err := binaryProtocol.Decode(reader, wire.TStruct)
//	if err != nil {
//	  return nil, err
//	}
//
//	var v DynamicConfigEntry
//	if err := v.FromWire(x); err != nil {
//	  return nil, err
//	}
//	return &v, nil
func (v *DynamicConfigEntry) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Name = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TList {
				v.Values, err = _List_DynamicConfigValue_Read(field.Value.GetList())
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

func _List_DynamicConfigValue_Encode(val []*DynamicConfigValue, sw stream.Writer) error {

	lh := stream.ListHeader{
		Type:   wire.TStruct,
		Length: len(val),
	}
	if err := sw.WriteListBegin(lh); err != nil {
		return err
	}

	for i, v := range val {
		if v == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigValue', index [%v]: value is nil", i)
		}
		if err := v.Encode(sw); err != nil {
			return err
		}
	}
	return sw.WriteListEnd()
}

// Encode serializes a DynamicConfigEntry struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a DynamicConfigEntry struct could not be encoded.
func (v *DynamicConfigEntry) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	if v.Name != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 10, Type: wire.TBinary}); err != nil {
			return err
		}
		if err := sw.WriteString(*(v.Name)); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	if v.Values != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 20, Type: wire.TList}); err != nil {
			return err
		}
		if err := _List_DynamicConfigValue_Encode(v.Values, sw); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	return sw.WriteStructEnd()
}

func _DynamicConfigValue_Decode(sr stream.Reader) (*DynamicConfigValue, error) {
	var v DynamicConfigValue
	err := v.Decode(sr)
	return &v, err
}

func _List_DynamicConfigValue_Decode(sr stream.Reader) ([]*DynamicConfigValue, error) {
	lh, err := sr.ReadListBegin()
	if err != nil {
		return nil, err
	}

	if lh.Type != wire.TStruct {
		for i := 0; i < lh.Length; i++ {
			if err := sr.Skip(lh.Type); err != nil {
				return nil, err
			}
		}
		return nil, sr.ReadListEnd()
	}

	o := make([]*DynamicConfigValue, 0, lh.Length)
	for i := 0; i < lh.Length; i++ {
		v, err := _DynamicConfigValue_Decode(sr)
		if err != nil {
			return nil, err
		}
		o = append(o, v)
	}

	if err = sr.ReadListEnd(); err != nil {
		return nil, err
	}
	return o, err
}

// Decode deserializes a DynamicConfigEntry struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a DynamicConfigEntry struct could not be generated from the wire
// representation.
func (v *DynamicConfigEntry) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		case fh.ID == 10 && fh.Type == wire.TBinary:
			var x string
			x, err = sr.ReadString()
			v.Name = &x
			if err != nil {
				return err
			}

		case fh.ID == 20 && fh.Type == wire.TList:
			v.Values, err = _List_DynamicConfigValue_Decode(sr)
			if err != nil {
				return err
			}

		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a DynamicConfigEntry
// struct.
func (v *DynamicConfigEntry) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.Name != nil {
		fields[i] = fmt.Sprintf("Name: %v", *(v.Name))
		i++
	}
	if v.Values != nil {
		fields[i] = fmt.Sprintf("Values: %v", v.Values)
		i++
	}

	return fmt.Sprintf("DynamicConfigEntry{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _List_DynamicConfigValue_Equals(lhs, rhs []*DynamicConfigValue) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this DynamicConfigEntry match the
// provided DynamicConfigEntry.
//
// This function performs a deep comparison.
func (v *DynamicConfigEntry) Equals(rhs *DynamicConfigEntry) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !_String_EqualsPtr(v.Name, rhs.Name) {
		return false
	}
	if !((v.Values == nil && rhs.Values == nil) || (v.Values != nil && rhs.Values != nil && _List_DynamicConfigValue_Equals(v.Values, rhs.Values))) {
		return false
	}

	return true
}

type _List_DynamicConfigValue_Zapper []*DynamicConfigValue

// MarshalLogArray implements zapcore.ArrayMarshaler, enabling
// fast logging of _List_DynamicConfigValue_Zapper.
func (l _List_DynamicConfigValue_Zapper) MarshalLogArray(enc zapcore.ArrayEncoder) (err error) {
	for _, v := range l {
		err = multierr.Append(err, enc.AppendObject(v))
	}
	return err
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of DynamicConfigEntry.
func (v *DynamicConfigEntry) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Name != nil {
		enc.AddString("name", *v.Name)
	}
	if v.Values != nil {
		err = multierr.Append(err, enc.AddArray("values", (_List_DynamicConfigValue_Zapper)(v.Values)))
	}
	return err
}

// GetName returns the value of Name if it is set or its
// zero value if it is unset.
func (v *DynamicConfigEntry) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}

	return
}

// IsSetName returns true if Name is not nil.
func (v *DynamicConfigEntry) IsSetName() bool {
	return v != nil && v.Name != nil
}

// GetValues returns the value of Values if it is set or its
// zero value if it is unset.
func (v *DynamicConfigEntry) GetValues() (o []*DynamicConfigValue) {
	if v != nil && v.Values != nil {
		return v.Values
	}

	return
}

// IsSetValues returns true if Values is not nil.
func (v *DynamicConfigEntry) IsSetValues() bool {
	return v != nil && v.Values != nil
}

type DynamicConfigFilter struct {
	Name  *string          `json:"name,omitempty"`
	Value *shared.DataBlob `json:"value,omitempty"`
}

// ToWire translates a DynamicConfigFilter struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//	x, err := v.ToWire()
//	if err != nil {
//	  return err
//	}
//
//	if err := binaryProtocol.Encode(x, writer); err != nil {
//	  return err
//	}
func (v *DynamicConfigFilter) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Name != nil {
		w, err = wire.NewValueString(*(v.Name)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.Value != nil {
		w, err = v.Value.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DataBlob_Read(w wire.Value) (*shared.DataBlob, error) {
	var v shared.DataBlob
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a DynamicConfigFilter struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a DynamicConfigFilter struct
// from the provided intermediate representation.
//
//	x, err := binaryProtocol.Decode(reader, wire.TStruct)
//	if err != nil {
//	  return nil, err
//	}
//
//	var v DynamicConfigFilter
//	if err := v.FromWire(x); err != nil {
//	  return nil, err
//	}
//	return &v, nil
func (v *DynamicConfigFilter) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Name = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TStruct {
				v.Value, err = _DataBlob_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// Encode serializes a DynamicConfigFilter struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a DynamicConfigFilter struct could not be encoded.
func (v *DynamicConfigFilter) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	if v.Name != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 10, Type: wire.TBinary}); err != nil {
			return err
		}
		if err := sw.WriteString(*(v.Name)); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	if v.Value != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 20, Type: wire.TStruct}); err != nil {
			return err
		}
		if err := v.Value.Encode(sw); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	return sw.WriteStructEnd()
}

func _DataBlob_Decode(sr stream.Reader) (*shared.DataBlob, error) {
	var v shared.DataBlob
	err := v.Decode(sr)
	return &v, err
}

// Decode deserializes a DynamicConfigFilter struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a DynamicConfigFilter struct could not be generated from the wire
// representation.
func (v *DynamicConfigFilter) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		case fh.ID == 10 && fh.Type == wire.TBinary:
			var x string
			x, err = sr.ReadString()
			v.Name = &x
			if err != nil {
				return err
			}

		case fh.ID == 20 && fh.Type == wire.TStruct:
			v.Value, err = _DataBlob_Decode(sr)
			if err != nil {
				return err
			}

		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a DynamicConfigFilter
// struct.
func (v *DynamicConfigFilter) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.Name != nil {
		fields[i] = fmt.Sprintf("Name: %v", *(v.Name))
		i++
	}
	if v.Value != nil {
		fields[i] = fmt.Sprintf("Value: %v", v.Value)
		i++
	}

	return fmt.Sprintf("DynamicConfigFilter{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this DynamicConfigFilter match the
// provided DynamicConfigFilter.
//
// This function performs a deep comparison.
func (v *DynamicConfigFilter) Equals(rhs *DynamicConfigFilter) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !_String_EqualsPtr(v.Name, rhs.Name) {
		return false
	}
	if !((v.Value == nil && rhs.Value == nil) || (v.Value != nil && rhs.Value != nil && v.Value.Equals(rhs.Value))) {
		return false
	}

	return true
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of DynamicConfigFilter.
func (v *DynamicConfigFilter) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Name != nil {
		enc.AddString("name", *v.Name)
	}
	if v.Value != nil {
		err = multierr.Append(err, enc.AddObject("value", v.Value))
	}
	return err
}

// GetName returns the value of Name if it is set or its
// zero value if it is unset.
func (v *DynamicConfigFilter) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}

	return
}

// IsSetName returns true if Name is not nil.
func (v *DynamicConfigFilter) IsSetName() bool {
	return v != nil && v.Name != nil
}

// GetValue returns the value of Value if it is set or its
// zero value if it is unset.
func (v *DynamicConfigFilter) GetValue() (o *shared.DataBlob) {
	if v != nil && v.Value != nil {
		return v.Value
	}

	return
}

// IsSetValue returns true if Value is not nil.
func (v *DynamicConfigFilter) IsSetValue() bool {
	return v != nil && v.Value != nil
}

type DynamicConfigValue struct {
	Value   *shared.DataBlob       `json:"value,omitempty"`
	Filters []*DynamicConfigFilter `json:"filters,omitempty"`
}

type _List_DynamicConfigFilter_ValueList []*DynamicConfigFilter

func (v _List_DynamicConfigFilter_ValueList) ForEach(f func(wire.Value) error) error {
	for i, x := range v {
		if x == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigFilter', index [%v]: value is nil", i)
		}
		w, err := x.ToWire()
		if err != nil {
			return err
		}
		err = f(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v _List_DynamicConfigFilter_ValueList) Size() int {
	return len(v)
}

func (_List_DynamicConfigFilter_ValueList) ValueType() wire.Type {
	return wire.TStruct
}

func (_List_DynamicConfigFilter_ValueList) Close() {}

// ToWire translates a DynamicConfigValue struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// An error is returned if the struct or any of its fields failed to
// validate.
//
//	x, err := v.ToWire()
//	if err != nil {
//	  return err
//	}
//
//	if err := binaryProtocol.Encode(x, writer); err != nil {
//	  return err
//	}
func (v *DynamicConfigValue) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.Value != nil {
		w, err = v.Value.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.Filters != nil {
		w, err = wire.NewValueList(_List_DynamicConfigFilter_ValueList(v.Filters)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DynamicConfigFilter_Read(w wire.Value) (*DynamicConfigFilter, error) {
	var v DynamicConfigFilter
	err := v.FromWire(w)
	return &v, err
}

func _List_DynamicConfigFilter_Read(l wire.ValueList) ([]*DynamicConfigFilter, error) {
	if l.ValueType() != wire.TStruct {
		return nil, nil
	}

	o := make([]*DynamicConfigFilter, 0, l.Size())
	err := l.ForEach(func(x wire.Value) error {
		i, err := _DynamicConfigFilter_Read(x)
		if err != nil {
			return err
		}
		o = append(o, i)
		return nil
	})
	l.Close()
	return o, err
}

// FromWire deserializes a DynamicConfigValue struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a DynamicConfigValue struct
// from the provided intermediate representation.
//
//	x, err := binaryProtocol.Decode(reader, wire.TStruct)
//	if err != nil {
//	  return nil, err
//	}
//
//	var v DynamicConfigValue
//	if err := v.FromWire(x); err != nil {
//	  return nil, err
//	}
//	return &v, nil
func (v *DynamicConfigValue) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TStruct {
				v.Value, err = _DataBlob_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TList {
				v.Filters, err = _List_DynamicConfigFilter_Read(field.Value.GetList())
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

func _List_DynamicConfigFilter_Encode(val []*DynamicConfigFilter, sw stream.Writer) error {

	lh := stream.ListHeader{
		Type:   wire.TStruct,
		Length: len(val),
	}
	if err := sw.WriteListBegin(lh); err != nil {
		return err
	}

	for i, v := range val {
		if v == nil {
			return fmt.Errorf("invalid list '[]*DynamicConfigFilter', index [%v]: value is nil", i)
		}
		if err := v.Encode(sw); err != nil {
			return err
		}
	}
	return sw.WriteListEnd()
}

// Encode serializes a DynamicConfigValue struct directly into bytes, without going
// through an intermediary type.
//
// An error is returned if a DynamicConfigValue struct could not be encoded.
func (v *DynamicConfigValue) Encode(sw stream.Writer) error {
	if err := sw.WriteStructBegin(); err != nil {
		return err
	}

	if v.Value != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 10, Type: wire.TStruct}); err != nil {
			return err
		}
		if err := v.Value.Encode(sw); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	if v.Filters != nil {
		if err := sw.WriteFieldBegin(stream.FieldHeader{ID: 20, Type: wire.TList}); err != nil {
			return err
		}
		if err := _List_DynamicConfigFilter_Encode(v.Filters, sw); err != nil {
			return err
		}
		if err := sw.WriteFieldEnd(); err != nil {
			return err
		}
	}

	return sw.WriteStructEnd()
}

func _DynamicConfigFilter_Decode(sr stream.Reader) (*DynamicConfigFilter, error) {
	var v DynamicConfigFilter
	err := v.Decode(sr)
	return &v, err
}

func _List_DynamicConfigFilter_Decode(sr stream.Reader) ([]*DynamicConfigFilter, error) {
	lh, err := sr.ReadListBegin()
	if err != nil {
		return nil, err
	}

	if lh.Type != wire.TStruct {
		for i := 0; i < lh.Length; i++ {
			if err := sr.Skip(lh.Type); err != nil {
				return nil, err
			}
		}
		return nil, sr.ReadListEnd()
	}

	o := make([]*DynamicConfigFilter, 0, lh.Length)
	for i := 0; i < lh.Length; i++ {
		v, err := _DynamicConfigFilter_Decode(sr)
		if err != nil {
			return nil, err
		}
		o = append(o, v)
	}

	if err = sr.ReadListEnd(); err != nil {
		return nil, err
	}
	return o, err
}

// Decode deserializes a DynamicConfigValue struct directly from its Thrift-level
// representation, without going through an intemediary type.
//
// An error is returned if a DynamicConfigValue struct could not be generated from the wire
// representation.
func (v *DynamicConfigValue) Decode(sr stream.Reader) error {

	if err := sr.ReadStructBegin(); err != nil {
		return err
	}

	fh, ok, err := sr.ReadFieldBegin()
	if err != nil {
		return err
	}

	for ok {
		switch {
		case fh.ID == 10 && fh.Type == wire.TStruct:
			v.Value, err = _DataBlob_Decode(sr)
			if err != nil {
				return err
			}

		case fh.ID == 20 && fh.Type == wire.TList:
			v.Filters, err = _List_DynamicConfigFilter_Decode(sr)
			if err != nil {
				return err
			}

		default:
			if err := sr.Skip(fh.Type); err != nil {
				return err
			}
		}

		if err := sr.ReadFieldEnd(); err != nil {
			return err
		}

		if fh, ok, err = sr.ReadFieldBegin(); err != nil {
			return err
		}
	}

	if err := sr.ReadStructEnd(); err != nil {
		return err
	}

	return nil
}

// String returns a readable string representation of a DynamicConfigValue
// struct.
func (v *DynamicConfigValue) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [2]string
	i := 0
	if v.Value != nil {
		fields[i] = fmt.Sprintf("Value: %v", v.Value)
		i++
	}
	if v.Filters != nil {
		fields[i] = fmt.Sprintf("Filters: %v", v.Filters)
		i++
	}

	return fmt.Sprintf("DynamicConfigValue{%v}", strings.Join(fields[:i], ", "))
}

func _List_DynamicConfigFilter_Equals(lhs, rhs []*DynamicConfigFilter) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lv := range lhs {
		rv := rhs[i]
		if !lv.Equals(rv) {
			return false
		}
	}

	return true
}

// Equals returns true if all the fields of this DynamicConfigValue match the
// provided DynamicConfigValue.
//
// This function performs a deep comparison.
func (v *DynamicConfigValue) Equals(rhs *DynamicConfigValue) bool {
	if v == nil {
		return rhs == nil
	} else if rhs == nil {
		return false
	}
	if !((v.Value == nil && rhs.Value == nil) || (v.Value != nil && rhs.Value != nil && v.Value.Equals(rhs.Value))) {
		return false
	}
	if !((v.Filters == nil && rhs.Filters == nil) || (v.Filters != nil && rhs.Filters != nil && _List_DynamicConfigFilter_Equals(v.Filters, rhs.Filters))) {
		return false
	}

	return true
}

type _List_DynamicConfigFilter_Zapper []*DynamicConfigFilter

// MarshalLogArray implements zapcore.ArrayMarshaler, enabling
// fast logging of _List_DynamicConfigFilter_Zapper.
func (l _List_DynamicConfigFilter_Zapper) MarshalLogArray(enc zapcore.ArrayEncoder) (err error) {
	for _, v := range l {
		err = multierr.Append(err, enc.AppendObject(v))
	}
	return err
}

// MarshalLogObject implements zapcore.ObjectMarshaler, enabling
// fast logging of DynamicConfigValue.
func (v *DynamicConfigValue) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	if v == nil {
		return nil
	}
	if v.Value != nil {
		err = multierr.Append(err, enc.AddObject("value", v.Value))
	}
	if v.Filters != nil {
		err = multierr.Append(err, enc.AddArray("filters", (_List_DynamicConfigFilter_Zapper)(v.Filters)))
	}
	return err
}

// GetValue returns the value of Value if it is set or its
// zero value if it is unset.
func (v *DynamicConfigValue) GetValue() (o *shared.DataBlob) {
	if v != nil && v.Value != nil {
		return v.Value
	}

	return
}

// IsSetValue returns true if Value is not nil.
func (v *DynamicConfigValue) IsSetValue() bool {
	return v != nil && v.Value != nil
}

// GetFilters returns the value of Filters if it is set or its
// zero value if it is unset.
func (v *DynamicConfigValue) GetFilters() (o []*DynamicConfigFilter) {
	if v != nil && v.Filters != nil {
		return v.Filters
	}

	return
}

// IsSetFilters returns true if Filters is not nil.
func (v *DynamicConfigValue) IsSetFilters() bool {
	return v != nil && v.Filters != nil
}

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "config",
	Package:  "github.com/uber/cadence/.gen/go/config",
	FilePath: "config.thrift",
	SHA1:     "cbc9d97e2a2f4820452fc2d273d5102e4721cf34",
	Includes: []*thriftreflect.ThriftModule{
		shared.ThriftModule,
	},
	Raw: rawIDL,
}

const rawIDL = "//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\nnamespace java com.uber.cadence.config\n\ninclude \"shared.thrift\"\n\nstruct DynamicConfigBlob {\n\t10: optional i64 schemaVersion\n\t20: optional list<DynamicConfigEntry> entries\n}\n\nstruct DynamicConfigEntry {\n  10: optional string name\n  20: optional list<DynamicConfigValue> values\n}\n\nstruct DynamicConfigValue {\n  10: optional shared.DataBlob value\n  20: optional list<DynamicConfigFilter> filters\n}\n\nstruct DynamicConfigFilter {\n  10: optional string name\n  20: optional shared.DataBlob value\n}\n"
