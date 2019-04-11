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

package tag

import (
	"time"
)

// Tag is the interface for logging system
type Tag interface {
	GetKey() string
	GetValueType() valueTypeEnum

	GetString() string
	GetInteger() int64
	GetDouble() float64
	GetBool() bool
	GetError() error
	GetDuration() time.Duration
	GetTime() time.Time
	GetObject() interface{}
}

var _ Tag = (*tagImpl)(nil)

// the tag value types supported
const (
	// ValueTypeString is string value type
	ValueTypeString valueTypeEnum = 1
	// ValueTypeInteger is any value of the types(int8,int16,int32,unit8,uint32,uint64) will be converted into int64
	ValueTypeInteger valueTypeEnum = 2
	// ValueTypeDouble is for both float and double will be converted into double
	ValueTypeDouble valueTypeEnum = 3
	// ValueTypeBool is bool value type
	ValueTypeBool valueTypeEnum = 4
	// ValueTypeError is error type value
	ValueTypeError valueTypeEnum = 5
	// ValueTypeDuration is duration type value
	ValueTypeDuration valueTypeEnum = 6
	// ValueTypeTime is time type value
	ValueTypeTime valueTypeEnum = 7
	// ValueTypeObject will be converted into string by fmt.Sprintf("%+v")
	ValueTypeObject valueTypeEnum = 8
)

// keep this module private
type valueTypeEnum int

// keep this module private
type tagImpl struct {
	key       string
	valueType valueTypeEnum

	valueString   string        //1
	valueInteger  int64         //2
	valueDouble   float64       //3
	valueBool     bool          //4
	valueError    error         //5
	valueDuration time.Duration //6
	valueTime     time.Time     //7
	valueObject   interface{}   //8
}

func (t *tagImpl) GetKey() string {
	return t.key
}

func (t *tagImpl) GetValueType() valueTypeEnum {
	return t.valueType
}

func (t *tagImpl) GetString() string {
	return t.valueString
}

func (t *tagImpl) GetInteger() int64 {
	return t.valueInteger
}

func (t *tagImpl) GetDouble() float64 {
	return t.valueDouble
}

func (t *tagImpl) GetBool() bool {
	return t.valueBool
}

func (t *tagImpl) GetError() error {
	return t.valueError
}

func (t *tagImpl) GetDuration() time.Duration {
	return t.valueDuration
}

func (t *tagImpl) GetTime() time.Time {
	return t.valueTime
}

func (t *tagImpl) GetObject() interface{} {
	return t.valueObject
}

func newStringTag(key string, value string) Tag {
	return &tagImpl{
		key:         key,
		valueType:   ValueTypeString,
		valueString: value,
	}
}

func newIntegerTag(key string, value int64) Tag {
	return &tagImpl{
		key:          key,
		valueType:    ValueTypeInteger,
		valueInteger: value,
	}
}

func newDoubleTag(key string, value float64) Tag {
	return &tagImpl{
		key:         key,
		valueType:   ValueTypeDouble,
		valueDouble: value,
	}
}

func newBoolTag(key string, value bool) Tag {
	return &tagImpl{
		key:       key,
		valueType: ValueTypeBool,
		valueBool: value,
	}
}

func newErrorTag(key string, value error) Tag {
	return &tagImpl{
		key:        key,
		valueType:  ValueTypeError,
		valueError: value,
	}
}

func newDurationTag(key string, value time.Duration) Tag {
	return &tagImpl{
		key:           key,
		valueType:     ValueTypeDuration,
		valueDuration: value,
	}
}

func newTimeTag(key string, value time.Time) Tag {
	return &tagImpl{
		key:       key,
		valueType: ValueTypeTime,
		valueTime: value,
	}
}

func newObjectTag(key string, value interface{}) Tag {
	return &tagImpl{
		key:         key,
		valueType:   ValueTypeObject,
		valueObject: value,
	}
}

func newPredefinedStringTag(key string, value string) Tag {
	return &tagImpl{
		key:         key,
		valueType:   ValueTypeString,
		valueString: value,
	}
}
