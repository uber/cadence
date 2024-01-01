// Copyright (c) 2020 Uber Technologies, Inc.
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

package execution

import (
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/testing/testdatagen"
	"github.com/uber/cadence/common/types"
)

func TestCopyActivityInfo(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test CopyActivityInfo Mapping", func(t *testing.T) {
			f := fuzz.New().Funcs(testdatagen.GenHistoryEvent).NilChance(0.2)

			d1 := persistence.ActivityInfo{}
			f.Fuzz(&d1)
			d2 := CopyActivityInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestCopyWorkflowExecutionInfo(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test ExecutionInfo Mapping", func(t *testing.T) {
			f := fuzz.New().Funcs(testdatagen.GenHistoryEvent).NilChance(0.2)

			d1 := persistence.WorkflowExecutionInfo{}
			f.Fuzz(&d1)
			d2 := CopyWorkflowExecutionInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestCopyTimerInfoMapping(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test Timer info Mapping", func(t *testing.T) {
			f := fuzz.New().NilChance(0.2)

			d1 := persistence.TimerInfo{}
			f.Fuzz(&d1)
			d2 := CopyTimerInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestChildWorkflowMapping(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test child workflwo info Mapping", func(t *testing.T) {
			f := fuzz.New().Funcs(testdatagen.GenHistoryEvent).NilChance(0.2)

			d1 := persistence.ChildExecutionInfo{}
			f.Fuzz(&d1)
			d2 := CopyChildInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestCopySignalInfo(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test signal info Mapping", func(t *testing.T) {
			f := fuzz.New().NilChance(0.2)

			d1 := persistence.SignalInfo{}
			f.Fuzz(&d1)
			d2 := CopySignalInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestCopyCancellationInfo(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("test signal info Mapping", func(t *testing.T) {
			f := fuzz.New().NilChance(0.2)

			d1 := persistence.RequestCancelInfo{}
			f.Fuzz(&d1)
			d2 := CopyCancellationInfo(&d1)

			assert.Equal(t, &d1, d2)
		})
	}
}

func TestFindAutoResetPoint(t *testing.T) {
	timeSource := clock.NewRealTimeSource()

	// case 1: nil
	_, pt := FindAutoResetPoint(timeSource, nil, nil)
	assert.Nil(t, pt)

	// case 2: empty
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{}, &types.ResetPoints{})
	assert.Nil(t, pt)

	pt0 := &types.ResetPointInfo{
		BinaryChecksum: "abc",
		Resettable:     true,
	}
	pt1 := &types.ResetPointInfo{
		BinaryChecksum: "def",
		Resettable:     true,
	}
	pt3 := &types.ResetPointInfo{
		BinaryChecksum: "ghi",
		Resettable:     false,
	}

	expiredNowNano := time.Now().UnixNano() - int64(time.Hour)
	notExpiredNowNano := time.Now().UnixNano() + int64(time.Hour)
	pt4 := &types.ResetPointInfo{
		BinaryChecksum:   "expired",
		Resettable:       true,
		ExpiringTimeNano: common.Int64Ptr(expiredNowNano),
	}

	pt5 := &types.ResetPointInfo{
		BinaryChecksum:   "notExpired",
		Resettable:       true,
		ExpiringTimeNano: common.Int64Ptr(notExpiredNowNano),
	}

	// case 3: two intersection
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"abc": {},
			"def": {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt0, pt1, pt3,
		},
	})
	assert.Equal(t, pt, pt0)

	// case 4: one intersection
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"none":    {},
			"def":     {},
			"expired": {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt4, pt0, pt1, pt3,
		},
	})
	assert.Equal(t, pt, pt1)

	// case 4: no intersection
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"none1": {},
			"none2": {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt0, pt1, pt3,
		},
	})
	assert.Nil(t, pt)

	// case 5: not resettable
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"none1": {},
			"ghi":   {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt0, pt1, pt3,
		},
	})
	assert.Nil(t, pt)

	// case 6: one intersection of expired
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"none":    {},
			"expired": {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt0, pt1, pt3, pt4, pt5,
		},
	})
	assert.Nil(t, pt)

	// case 7: one intersection of not expired
	_, pt = FindAutoResetPoint(timeSource, &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			"none":       {},
			"notExpired": {},
		},
	}, &types.ResetPoints{
		Points: []*types.ResetPointInfo{
			pt0, pt1, pt3, pt4, pt5,
		},
	})
	assert.Equal(t, pt, pt5)
}
