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

package replicator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"math"
	"strconv"
	"strings"
)

type DomainOperation int32

const (
	DomainOperationCreate DomainOperation = 0
	DomainOperationUpdate DomainOperation = 1
)

// DomainOperation_Values returns all recognized values of DomainOperation.
func DomainOperation_Values() []DomainOperation {
	return []DomainOperation{
		DomainOperationCreate,
		DomainOperationUpdate,
	}
}

// UnmarshalText tries to decode DomainOperation from a byte slice
// containing its name.
//
//   var v DomainOperation
//   err := v.UnmarshalText([]byte("Create"))
func (v *DomainOperation) UnmarshalText(value []byte) error {
	switch string(value) {
	case "Create":
		*v = DomainOperationCreate
		return nil
	case "Update":
		*v = DomainOperationUpdate
		return nil
	default:
		return fmt.Errorf("unknown enum value %q for %q", value, "DomainOperation")
	}
}

// Ptr returns a pointer to this enum value.
func (v DomainOperation) Ptr() *DomainOperation {
	return &v
}

// ToWire translates DomainOperation into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// Enums are represented as 32-bit integers over the wire.
func (v DomainOperation) ToWire() (wire.Value, error) {
	return wire.NewValueI32(int32(v)), nil
}

// FromWire deserializes DomainOperation from its Thrift-level
// representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TI32)
//   if err != nil {
//     return DomainOperation(0), err
//   }
//
//   var v DomainOperation
//   if err := v.FromWire(x); err != nil {
//     return DomainOperation(0), err
//   }
//   return v, nil
func (v *DomainOperation) FromWire(w wire.Value) error {
	*v = (DomainOperation)(w.GetI32())
	return nil
}

// String returns a readable string representation of DomainOperation.
func (v DomainOperation) String() string {
	w := int32(v)
	switch w {
	case 0:
		return "Create"
	case 1:
		return "Update"
	}
	return fmt.Sprintf("DomainOperation(%d)", w)
}

// Equals returns true if this DomainOperation value matches the provided
// value.
func (v DomainOperation) Equals(rhs DomainOperation) bool {
	return v == rhs
}

// MarshalJSON serializes DomainOperation into JSON.
//
// If the enum value is recognized, its name is returned. Otherwise,
// its integer value is returned.
//
// This implements json.Marshaler.
func (v DomainOperation) MarshalJSON() ([]byte, error) {
	switch int32(v) {
	case 0:
		return ([]byte)("\"Create\""), nil
	case 1:
		return ([]byte)("\"Update\""), nil
	}
	return ([]byte)(strconv.FormatInt(int64(v), 10)), nil
}

// UnmarshalJSON attempts to decode DomainOperation from its JSON
// representation.
//
// This implementation supports both, numeric and string inputs. If a
// string is provided, it must be a known enum name.
//
// This implements json.Unmarshaler.
func (v *DomainOperation) UnmarshalJSON(text []byte) error {
	d := json.NewDecoder(bytes.NewReader(text))
	d.UseNumber()
	t, err := d.Token()
	if err != nil {
		return err
	}

	switch w := t.(type) {
	case json.Number:
		x, err := w.Int64()
		if err != nil {
			return err
		}
		if x > math.MaxInt32 {
			return fmt.Errorf("enum overflow from JSON %q for %q", text, "DomainOperation")
		}
		if x < math.MinInt32 {
			return fmt.Errorf("enum underflow from JSON %q for %q", text, "DomainOperation")
		}
		*v = (DomainOperation)(x)
		return nil
	case string:
		return v.UnmarshalText([]byte(w))
	default:
		return fmt.Errorf("invalid JSON value %q (%T) to unmarshal into %q", t, t, "DomainOperation")
	}
}

type DomainTaskAttributes struct {
	DomainOperation   *DomainOperation                       `json:"domainOperation,omitempty"`
	ID                *string                                `json:"id,omitempty"`
	Info              *shared.DomainInfo                     `json:"info,omitempty"`
	Config            *shared.DomainConfiguration            `json:"config,omitempty"`
	ReplicationConfig *shared.DomainReplicationConfiguration `json:"replicationConfig,omitempty"`
	FailoverVersion   *int64                                 `json:"failoverVersion,omitempty"`
}

// ToWire translates a DomainTaskAttributes struct into a Thrift-level intermediate
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
func (v *DomainTaskAttributes) ToWire() (wire.Value, error) {
	var (
		fields [6]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.DomainOperation != nil {
		w, err = v.DomainOperation.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 5, Value: w}
		i++
	}
	if v.ID != nil {
		w, err = wire.NewValueString(*(v.ID)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.Info != nil {
		w, err = v.Info.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}
	if v.Config != nil {
		w, err = v.Config.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 30, Value: w}
		i++
	}
	if v.ReplicationConfig != nil {
		w, err = v.ReplicationConfig.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 40, Value: w}
		i++
	}
	if v.FailoverVersion != nil {
		w, err = wire.NewValueI64(*(v.FailoverVersion)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 50, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _DomainOperation_Read(w wire.Value) (DomainOperation, error) {
	var v DomainOperation
	err := v.FromWire(w)
	return v, err
}

func _DomainInfo_Read(w wire.Value) (*shared.DomainInfo, error) {
	var v shared.DomainInfo
	err := v.FromWire(w)
	return &v, err
}

func _DomainConfiguration_Read(w wire.Value) (*shared.DomainConfiguration, error) {
	var v shared.DomainConfiguration
	err := v.FromWire(w)
	return &v, err
}

func _DomainReplicationConfiguration_Read(w wire.Value) (*shared.DomainReplicationConfiguration, error) {
	var v shared.DomainReplicationConfiguration
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a DomainTaskAttributes struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a DomainTaskAttributes struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v DomainTaskAttributes
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *DomainTaskAttributes) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 5:
			if field.Value.Type() == wire.TI32 {
				var x DomainOperation
				x, err = _DomainOperation_Read(field.Value)
				v.DomainOperation = &x
				if err != nil {
					return err
				}

			}
		case 10:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.ID = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TStruct {
				v.Info, err = _DomainInfo_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 30:
			if field.Value.Type() == wire.TStruct {
				v.Config, err = _DomainConfiguration_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 40:
			if field.Value.Type() == wire.TStruct {
				v.ReplicationConfig, err = _DomainReplicationConfiguration_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 50:
			if field.Value.Type() == wire.TI64 {
				var x int64
				x, err = field.Value.GetI64(), error(nil)
				v.FailoverVersion = &x
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a DomainTaskAttributes
// struct.
func (v *DomainTaskAttributes) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [6]string
	i := 0
	if v.DomainOperation != nil {
		fields[i] = fmt.Sprintf("DomainOperation: %v", *(v.DomainOperation))
		i++
	}
	if v.ID != nil {
		fields[i] = fmt.Sprintf("ID: %v", *(v.ID))
		i++
	}
	if v.Info != nil {
		fields[i] = fmt.Sprintf("Info: %v", v.Info)
		i++
	}
	if v.Config != nil {
		fields[i] = fmt.Sprintf("Config: %v", v.Config)
		i++
	}
	if v.ReplicationConfig != nil {
		fields[i] = fmt.Sprintf("ReplicationConfig: %v", v.ReplicationConfig)
		i++
	}
	if v.FailoverVersion != nil {
		fields[i] = fmt.Sprintf("FailoverVersion: %v", *(v.FailoverVersion))
		i++
	}

	return fmt.Sprintf("DomainTaskAttributes{%v}", strings.Join(fields[:i], ", "))
}

func _DomainOperation_EqualsPtr(lhs, rhs *DomainOperation) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return x.Equals(y)
	}
	return lhs == nil && rhs == nil
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func _I64_EqualsPtr(lhs, rhs *int64) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this DomainTaskAttributes match the
// provided DomainTaskAttributes.
//
// This function performs a deep comparison.
func (v *DomainTaskAttributes) Equals(rhs *DomainTaskAttributes) bool {
	if !_DomainOperation_EqualsPtr(v.DomainOperation, rhs.DomainOperation) {
		return false
	}
	if !_String_EqualsPtr(v.ID, rhs.ID) {
		return false
	}
	if !((v.Info == nil && rhs.Info == nil) || (v.Info != nil && rhs.Info != nil && v.Info.Equals(rhs.Info))) {
		return false
	}
	if !((v.Config == nil && rhs.Config == nil) || (v.Config != nil && rhs.Config != nil && v.Config.Equals(rhs.Config))) {
		return false
	}
	if !((v.ReplicationConfig == nil && rhs.ReplicationConfig == nil) || (v.ReplicationConfig != nil && rhs.ReplicationConfig != nil && v.ReplicationConfig.Equals(rhs.ReplicationConfig))) {
		return false
	}
	if !_I64_EqualsPtr(v.FailoverVersion, rhs.FailoverVersion) {
		return false
	}

	return true
}

// GetDomainOperation returns the value of DomainOperation if it is set or its
// zero value if it is unset.
func (v *DomainTaskAttributes) GetDomainOperation() (o DomainOperation) {
	if v.DomainOperation != nil {
		return *v.DomainOperation
	}

	return
}

// GetID returns the value of ID if it is set or its
// zero value if it is unset.
func (v *DomainTaskAttributes) GetID() (o string) {
	if v.ID != nil {
		return *v.ID
	}

	return
}

// GetFailoverVersion returns the value of FailoverVersion if it is set or its
// zero value if it is unset.
func (v *DomainTaskAttributes) GetFailoverVersion() (o int64) {
	if v.FailoverVersion != nil {
		return *v.FailoverVersion
	}

	return
}

type HistoryTaskAttributes struct {
}

// ToWire translates a HistoryTaskAttributes struct into a Thrift-level intermediate
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
func (v *HistoryTaskAttributes) ToWire() (wire.Value, error) {
	var (
		fields [0]wire.Field
		i      int = 0
	)

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

// FromWire deserializes a HistoryTaskAttributes struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a HistoryTaskAttributes struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v HistoryTaskAttributes
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *HistoryTaskAttributes) FromWire(w wire.Value) error {

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		}
	}

	return nil
}

// String returns a readable string representation of a HistoryTaskAttributes
// struct.
func (v *HistoryTaskAttributes) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [0]string
	i := 0

	return fmt.Sprintf("HistoryTaskAttributes{%v}", strings.Join(fields[:i], ", "))
}

// Equals returns true if all the fields of this HistoryTaskAttributes match the
// provided HistoryTaskAttributes.
//
// This function performs a deep comparison.
func (v *HistoryTaskAttributes) Equals(rhs *HistoryTaskAttributes) bool {

	return true
}

type ReplicationTask struct {
	TaskType              *ReplicationTaskType   `json:"taskType,omitempty"`
	DomainTaskAttributes  *DomainTaskAttributes  `json:"domainTaskAttributes,omitempty"`
	HistoryTaskAttributes *HistoryTaskAttributes `json:"historyTaskAttributes,omitempty"`
}

// ToWire translates a ReplicationTask struct into a Thrift-level intermediate
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
func (v *ReplicationTask) ToWire() (wire.Value, error) {
	var (
		fields [3]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)

	if v.TaskType != nil {
		w, err = v.TaskType.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 10, Value: w}
		i++
	}
	if v.DomainTaskAttributes != nil {
		w, err = v.DomainTaskAttributes.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 20, Value: w}
		i++
	}
	if v.HistoryTaskAttributes != nil {
		w, err = v.HistoryTaskAttributes.ToWire()
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 30, Value: w}
		i++
	}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func _ReplicationTaskType_Read(w wire.Value) (ReplicationTaskType, error) {
	var v ReplicationTaskType
	err := v.FromWire(w)
	return v, err
}

func _DomainTaskAttributes_Read(w wire.Value) (*DomainTaskAttributes, error) {
	var v DomainTaskAttributes
	err := v.FromWire(w)
	return &v, err
}

func _HistoryTaskAttributes_Read(w wire.Value) (*HistoryTaskAttributes, error) {
	var v HistoryTaskAttributes
	err := v.FromWire(w)
	return &v, err
}

// FromWire deserializes a ReplicationTask struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
//
// An error is returned if we were unable to build a ReplicationTask struct
// from the provided intermediate representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TStruct)
//   if err != nil {
//     return nil, err
//   }
//
//   var v ReplicationTask
//   if err := v.FromWire(x); err != nil {
//     return nil, err
//   }
//   return &v, nil
func (v *ReplicationTask) FromWire(w wire.Value) error {
	var err error

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 10:
			if field.Value.Type() == wire.TI32 {
				var x ReplicationTaskType
				x, err = _ReplicationTaskType_Read(field.Value)
				v.TaskType = &x
				if err != nil {
					return err
				}

			}
		case 20:
			if field.Value.Type() == wire.TStruct {
				v.DomainTaskAttributes, err = _DomainTaskAttributes_Read(field.Value)
				if err != nil {
					return err
				}

			}
		case 30:
			if field.Value.Type() == wire.TStruct {
				v.HistoryTaskAttributes, err = _HistoryTaskAttributes_Read(field.Value)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// String returns a readable string representation of a ReplicationTask
// struct.
func (v *ReplicationTask) String() string {
	if v == nil {
		return "<nil>"
	}

	var fields [3]string
	i := 0
	if v.TaskType != nil {
		fields[i] = fmt.Sprintf("TaskType: %v", *(v.TaskType))
		i++
	}
	if v.DomainTaskAttributes != nil {
		fields[i] = fmt.Sprintf("DomainTaskAttributes: %v", v.DomainTaskAttributes)
		i++
	}
	if v.HistoryTaskAttributes != nil {
		fields[i] = fmt.Sprintf("HistoryTaskAttributes: %v", v.HistoryTaskAttributes)
		i++
	}

	return fmt.Sprintf("ReplicationTask{%v}", strings.Join(fields[:i], ", "))
}

func _ReplicationTaskType_EqualsPtr(lhs, rhs *ReplicationTaskType) bool {
	if lhs != nil && rhs != nil {

		x := *lhs
		y := *rhs
		return x.Equals(y)
	}
	return lhs == nil && rhs == nil
}

// Equals returns true if all the fields of this ReplicationTask match the
// provided ReplicationTask.
//
// This function performs a deep comparison.
func (v *ReplicationTask) Equals(rhs *ReplicationTask) bool {
	if !_ReplicationTaskType_EqualsPtr(v.TaskType, rhs.TaskType) {
		return false
	}
	if !((v.DomainTaskAttributes == nil && rhs.DomainTaskAttributes == nil) || (v.DomainTaskAttributes != nil && rhs.DomainTaskAttributes != nil && v.DomainTaskAttributes.Equals(rhs.DomainTaskAttributes))) {
		return false
	}
	if !((v.HistoryTaskAttributes == nil && rhs.HistoryTaskAttributes == nil) || (v.HistoryTaskAttributes != nil && rhs.HistoryTaskAttributes != nil && v.HistoryTaskAttributes.Equals(rhs.HistoryTaskAttributes))) {
		return false
	}

	return true
}

// GetTaskType returns the value of TaskType if it is set or its
// zero value if it is unset.
func (v *ReplicationTask) GetTaskType() (o ReplicationTaskType) {
	if v.TaskType != nil {
		return *v.TaskType
	}

	return
}

type ReplicationTaskType int32

const (
	ReplicationTaskTypeDomain  ReplicationTaskType = 0
	ReplicationTaskTypeHistory ReplicationTaskType = 1
)

// ReplicationTaskType_Values returns all recognized values of ReplicationTaskType.
func ReplicationTaskType_Values() []ReplicationTaskType {
	return []ReplicationTaskType{
		ReplicationTaskTypeDomain,
		ReplicationTaskTypeHistory,
	}
}

// UnmarshalText tries to decode ReplicationTaskType from a byte slice
// containing its name.
//
//   var v ReplicationTaskType
//   err := v.UnmarshalText([]byte("Domain"))
func (v *ReplicationTaskType) UnmarshalText(value []byte) error {
	switch string(value) {
	case "Domain":
		*v = ReplicationTaskTypeDomain
		return nil
	case "History":
		*v = ReplicationTaskTypeHistory
		return nil
	default:
		return fmt.Errorf("unknown enum value %q for %q", value, "ReplicationTaskType")
	}
}

// Ptr returns a pointer to this enum value.
func (v ReplicationTaskType) Ptr() *ReplicationTaskType {
	return &v
}

// ToWire translates ReplicationTaskType into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
//
// Enums are represented as 32-bit integers over the wire.
func (v ReplicationTaskType) ToWire() (wire.Value, error) {
	return wire.NewValueI32(int32(v)), nil
}

// FromWire deserializes ReplicationTaskType from its Thrift-level
// representation.
//
//   x, err := binaryProtocol.Decode(reader, wire.TI32)
//   if err != nil {
//     return ReplicationTaskType(0), err
//   }
//
//   var v ReplicationTaskType
//   if err := v.FromWire(x); err != nil {
//     return ReplicationTaskType(0), err
//   }
//   return v, nil
func (v *ReplicationTaskType) FromWire(w wire.Value) error {
	*v = (ReplicationTaskType)(w.GetI32())
	return nil
}

// String returns a readable string representation of ReplicationTaskType.
func (v ReplicationTaskType) String() string {
	w := int32(v)
	switch w {
	case 0:
		return "Domain"
	case 1:
		return "History"
	}
	return fmt.Sprintf("ReplicationTaskType(%d)", w)
}

// Equals returns true if this ReplicationTaskType value matches the provided
// value.
func (v ReplicationTaskType) Equals(rhs ReplicationTaskType) bool {
	return v == rhs
}

// MarshalJSON serializes ReplicationTaskType into JSON.
//
// If the enum value is recognized, its name is returned. Otherwise,
// its integer value is returned.
//
// This implements json.Marshaler.
func (v ReplicationTaskType) MarshalJSON() ([]byte, error) {
	switch int32(v) {
	case 0:
		return ([]byte)("\"Domain\""), nil
	case 1:
		return ([]byte)("\"History\""), nil
	}
	return ([]byte)(strconv.FormatInt(int64(v), 10)), nil
}

// UnmarshalJSON attempts to decode ReplicationTaskType from its JSON
// representation.
//
// This implementation supports both, numeric and string inputs. If a
// string is provided, it must be a known enum name.
//
// This implements json.Unmarshaler.
func (v *ReplicationTaskType) UnmarshalJSON(text []byte) error {
	d := json.NewDecoder(bytes.NewReader(text))
	d.UseNumber()
	t, err := d.Token()
	if err != nil {
		return err
	}

	switch w := t.(type) {
	case json.Number:
		x, err := w.Int64()
		if err != nil {
			return err
		}
		if x > math.MaxInt32 {
			return fmt.Errorf("enum overflow from JSON %q for %q", text, "ReplicationTaskType")
		}
		if x < math.MinInt32 {
			return fmt.Errorf("enum underflow from JSON %q for %q", text, "ReplicationTaskType")
		}
		*v = (ReplicationTaskType)(x)
		return nil
	case string:
		return v.UnmarshalText([]byte(w))
	default:
		return fmt.Errorf("invalid JSON value %q (%T) to unmarshal into %q", t, t, "ReplicationTaskType")
	}
}
