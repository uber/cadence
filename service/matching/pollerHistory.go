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

package matching

import (
	"sort"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/types"
)

const (
	pollerHistoryInitSize    = 0
	pollerHistoryInitMaxSize = 5000
	pollerHistoryTTL         = 5 * time.Minute
)

type (
	pollerIdentity string

	pollerInfo struct {
		ratePerSecond  *float64
		isolationGroup string
	}
)

type pollerHistory struct {
	// poller ID -> pollerInfo
	// pollers map[pollerID]pollerInfo
	history cache.Cache

	// OnHistoryUpdatedFunc is a function called when the poller history was updated
	onHistoryUpdatedFunc HistoryUpdatedFunc
}

// HistoryUpdatedFunc is a type for notifying applications when the poller history was updated
type HistoryUpdatedFunc func()

func newPollerHistory(historyUpdatedFunc HistoryUpdatedFunc) *pollerHistory {
	opts := &cache.Options{
		InitialCapacity: pollerHistoryInitSize,
		TTL:             pollerHistoryTTL,
		Pin:             false,
		MaxCount:        pollerHistoryInitMaxSize,
	}

	return &pollerHistory{
		history:              cache.New(opts),
		onHistoryUpdatedFunc: historyUpdatedFunc,
	}
}

func (pollers *pollerHistory) updatePollerInfo(id pollerIdentity, info pollerInfo) {
	pollers.history.Put(id, &info)
	if pollers.onHistoryUpdatedFunc != nil {
		pollers.onHistoryUpdatedFunc()
	}
}

func (pollers *pollerHistory) getPollerInfo(earliestAccessTime time.Time) []*types.PollerInfo {
	var result []*types.PollerInfo

	ite := pollers.history.Iterator()
	defer ite.Close()
	for ite.HasNext() {
		entry := ite.Next()
		key := entry.Key().(pollerIdentity)
		value := entry.Value().(*pollerInfo)
		// TODO add IP, T1396795
		lastAccessTime := entry.CreateTime()
		if earliestAccessTime.Before(lastAccessTime) {
			rps := _defaultTaskDispatchRPS
			if value.ratePerSecond != nil {
				rps = *value.ratePerSecond
			}
			result = append(result, &types.PollerInfo{
				Identity:       string(key),
				LastAccessTime: common.Int64Ptr(lastAccessTime.UnixNano()),
				RatePerSecond:  rps,
			})
		}
	}

	return result
}

func (pollers *pollerHistory) getPollerIsolationGroups(earliestAccessTime time.Time) []string {
	groupSet := make(map[string]struct{})
	ite := pollers.history.Iterator()
	defer ite.Close()
	for ite.HasNext() {
		entry := ite.Next()
		value := entry.Value().(*pollerInfo)
		lastAccessTime := entry.CreateTime()
		if earliestAccessTime.Before(lastAccessTime) {
			if value.isolationGroup != "" {
				groupSet[value.isolationGroup] = struct{}{}
			}
		}
	}
	var result []string
	for k := range groupSet {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
