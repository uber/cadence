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

package poller

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/types"
)

func TestNewPollerHistory(t *testing.T) {
	p := NewPollerHistory(nil, nil)
	assert.NotNil(t, p)
	assert.NotNil(t, p.history)
}

func TestUpdatePollerInfo(t *testing.T) {
	updated := false
	updateFn := func() {
		updated = true
	}
	mockCtrl := gomock.NewController(t)
	mockCache := cache.NewMockCache(mockCtrl)
	mockCache.EXPECT().Put(Identity("test"), &Info{IsolationGroup: "dca1"}).Return(nil)
	p := &History{
		onHistoryUpdatedFunc: updateFn,
		history:              mockCache,
	}
	p.UpdatePollerInfo(Identity("test"), Info{IsolationGroup: "dca1"})
	assert.True(t, updated)
}

func TestGetPollerInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockCache := cache.NewMockCache(mockCtrl)
	mockIter := cache.NewMockIterator(mockCtrl)
	mockEntry := cache.NewMockEntry(mockCtrl)

	mockCache.EXPECT().Iterator().Return(mockIter)
	gomock.InOrder(
		mockIter.EXPECT().HasNext().Return(true),
		mockIter.EXPECT().Next().Return(mockEntry),
		mockEntry.EXPECT().Key().Return(Identity("test0")),
		mockEntry.EXPECT().Value().Return(&Info{IsolationGroup: "dca1", RatePerSecond: 1.0}),
		mockEntry.EXPECT().CreateTime().Return(time.UnixMilli(1000)),
		mockIter.EXPECT().HasNext().Return(true),
		mockIter.EXPECT().Next().Return(mockEntry),
		mockEntry.EXPECT().Key().Return(Identity("test1")),
		mockEntry.EXPECT().Value().Return(&Info{IsolationGroup: "dca2", RatePerSecond: 2.0}),
		mockEntry.EXPECT().CreateTime().Return(time.UnixMilli(0)),
		mockIter.EXPECT().HasNext().Return(false),
		mockIter.EXPECT().Close(),
	)
	p := &History{
		history: mockCache,
	}
	info := p.GetPollerInfo(time.UnixMilli(500))
	assert.Equal(t, []*types.PollerInfo{
		{
			Identity:       "test0",
			LastAccessTime: common.Ptr(time.UnixMilli(1000).UnixNano()),
			RatePerSecond:  1.0,
		},
	}, info)
}

func TestGetPollerIsolationGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockCache := cache.NewMockCache(mockCtrl)
	mockIter := cache.NewMockIterator(mockCtrl)
	mockEntry := cache.NewMockEntry(mockCtrl)

	mockCache.EXPECT().Iterator().Return(mockIter)
	gomock.InOrder(
		mockIter.EXPECT().HasNext().Return(true),
		mockIter.EXPECT().Next().Return(mockEntry),
		mockEntry.EXPECT().Value().Return(&Info{IsolationGroup: "dca1", RatePerSecond: 1.0}),
		mockEntry.EXPECT().CreateTime().Return(time.UnixMilli(1000)),
		mockIter.EXPECT().HasNext().Return(true),
		mockIter.EXPECT().Next().Return(mockEntry),
		mockEntry.EXPECT().Value().Return(&Info{IsolationGroup: "dca2", RatePerSecond: 2.0}),
		mockEntry.EXPECT().CreateTime().Return(time.UnixMilli(0)),
		mockIter.EXPECT().HasNext().Return(false),
		mockIter.EXPECT().Close(),
	)
	p := &History{
		history: mockCache,
	}
	groups := p.GetPollerIsolationGroups(time.UnixMilli(500))
	assert.Equal(t, map[string]struct{}{"dca1": struct{}{}}, groups)
}
