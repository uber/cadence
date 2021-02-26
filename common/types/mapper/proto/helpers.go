// Copyright (c) 2021 Uber Technologies Inc.
//
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

package proto

import (
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/uber/cadence/common"
)

func fromDoubleValue(v *float64) *wrappers.DoubleValue {
	if v == nil {
		return nil
	}
	return &wrappers.DoubleValue{Value: *v}
}

func toDoubleValue(v *wrappers.DoubleValue) *float64 {
	if v == nil {
		return nil
	}
	return common.Float64Ptr(v.Value)
}

func fromInt64Value(v *int64) *wrappers.Int64Value {
	if v == nil {
		return nil
	}
	return &wrappers.Int64Value{Value: *v}
}

func toInt64Value(v *wrappers.Int64Value) *int64 {
	if v == nil {
		return nil
	}
	return common.Int64Ptr(v.Value)
}

func unixNanoToTime(t *int64) *timestamp.Timestamp {
	if t == nil {
		return nil
	}
	time, err := ptypes.TimestampProto(time.Unix(0, *t))
	if err != nil {
		panic(err)
	}
	return time
}

func timeToUnixNano(t *timestamp.Timestamp) *int64 {
	if t == nil {
		return nil
	}
	return common.Int64Ptr(t.AsTime().UnixNano())
}

func daysToDuration(d *int32) *duration.Duration {
	if d == nil {
		return nil
	}
	return ptypes.DurationProto(common.DaysToDuration(*d))
}

func durationToDays(d *duration.Duration) *int32 {
	if d == nil {
		return nil
	}
	return common.Int32Ptr(common.DurationToDays(d.AsDuration()))
}

func secondsToDuration(d *int32) *duration.Duration {
	if d == nil {
		return nil
	}
	return ptypes.DurationProto(common.SecondsToDuration(int64(*d)))
}

func durationToSeconds(d *duration.Duration) *int32 {
	if d == nil {
		return nil
	}
	return common.Int32Ptr(int32(common.DurationToSeconds(d.AsDuration())))
}

func int32To64(v *int32) *int64 {
	if v == nil {
		return nil
	}
	return common.Int64Ptr(int64(*v))
}

func int64To32(v *int64) *int32 {
	if v == nil {
		return nil
	}
	return common.Int32Ptr(int32(*v))
}

func stringToInt32(s string) int32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return int32(i)
}

func int32ToString(i int32) string {
	s := strconv.Itoa(int(i))
	return s
}

type fieldSet map[string]struct{}

func newFieldSet(mask *fieldmaskpb.FieldMask) fieldSet {
	if mask == nil {
		return nil
	}
	fs := map[string]struct{}{}
	for _, field := range mask.Paths {
		fs[field] = struct{}{}
	}
	return fs
}

func (fs fieldSet) isSet(field string) bool {
	if fs == nil {
		return true
	}
	_, ok := fs[field]
	return ok
}
