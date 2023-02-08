// Copyright (c) 2022 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/collection"
	persistenceutils "github.com/uber/cadence/common/persistence/persistence-utils"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

// GetResurrectedTimers returns a set of timers (timer IDs) that were resurrected.
// Meaning timers that are still pending in mutable state, but were already completed based on event history.
func GetResurrectedTimers(
	ctx context.Context,
	shard shard.Context,
	mutableState MutableState,
) (map[string]struct{}, error) {
	// 1. check if there is any pending timer
	pendingTimerInfos := mutableState.GetPendingTimerInfos()
	if len(pendingTimerInfos) == 0 {
		return map[string]struct{}{}, nil
	}

	// 2. scan history from the beginning and see if any
	// TimerFiredEvent or TimerCancelledEvent matches pending timer
	// NOTE: We can't read from the middle of events branch, because
	// we don't know the last txn id of previous event from the middle.
	// Reading from the middle could get invalid nodes with invalid txn ids.
	resurrectedTimer := make(map[string]struct{})
	branchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return nil, err
	}
	domainID := mutableState.GetExecutionInfo().DomainID
	iter := collection.NewPagingIterator(getHistoryPaginationFn(
		ctx,
		shard,
		1,
		mutableState.GetNextEventID(),
		branchToken,
		domainID,
	))
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			return nil, err
		}
		event := item.(*types.HistoryEvent)
		var timerID string
		switch event.GetEventType() {
		case types.EventTypeTimerFired:
			timerID = event.TimerFiredEventAttributes.TimerID
		case types.EventTypeTimerCanceled:
			timerID = event.TimerCanceledEventAttributes.TimerID
		}
		if _, ok := pendingTimerInfos[timerID]; ok && timerID != "" {
			resurrectedTimer[timerID] = struct{}{}
		}
	}
	return resurrectedTimer, nil
}

// GetResurrectedActivities returns a set of activities (schedule IDs) that were resurrected.
// Meaning activities that are still pending in mutable state, but were already completed based on event history.
func GetResurrectedActivities(
	ctx context.Context,
	shard shard.Context,
	mutableState MutableState,
) (map[int64]struct{}, error) {
	// 1. check if there is any pending activity
	pendingActivityInfos := mutableState.GetPendingActivityInfos()
	if len(pendingActivityInfos) == 0 {
		return map[int64]struct{}{}, nil
	}

	// 2. scan history from the beginning and see if any
	// activity termination events matches pending activity
	// NOTE: We can't read from the middle of events branch, because
	// we don't know the last txn id of previous event from the middle.
	// Reading from the middle could get invalid nodes with invalid txn ids.
	resurrectedActivity := make(map[int64]struct{})
	branchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return nil, err
	}
	domainID := mutableState.GetExecutionInfo().DomainID
	iter := collection.NewPagingIterator(getHistoryPaginationFn(
		ctx,
		shard,
		1,
		mutableState.GetNextEventID(),
		branchToken,
		domainID,
	))
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			return nil, err
		}
		event := item.(*types.HistoryEvent)
		var scheduledID int64
		switch event.GetEventType() {
		case types.EventTypeActivityTaskCompleted:
			scheduledID = event.ActivityTaskCompletedEventAttributes.ScheduledEventID
		case types.EventTypeActivityTaskFailed:
			scheduledID = event.ActivityTaskFailedEventAttributes.ScheduledEventID
		case types.EventTypeActivityTaskTimedOut:
			scheduledID = event.ActivityTaskTimedOutEventAttributes.ScheduledEventID
		case types.EventTypeActivityTaskCanceled:
			scheduledID = event.ActivityTaskCanceledEventAttributes.ScheduledEventID
		}
		if _, ok := pendingActivityInfos[scheduledID]; ok && scheduledID != 0 {
			resurrectedActivity[scheduledID] = struct{}{}
		}
	}
	return resurrectedActivity, nil
}

func getHistoryPaginationFn(
	ctx context.Context,
	shard shard.Context,
	firstEventID int64,
	nextEventID int64,
	branchToken []byte,
	domainID string,
) collection.PaginationFn {
	domainCache := shard.GetDomainCache()
	return func(token []byte) ([]interface{}, []byte, error) {
		historyEvents, _, token, _, err := persistenceutils.PaginateHistory(
			ctx,
			shard.GetHistoryManager(),
			false,
			branchToken,
			firstEventID,
			nextEventID,
			token,
			NDCDefaultPageSize,
			common.IntPtr(shard.GetShardID()),
			domainID,
			domainCache,
		)
		if err != nil {
			return nil, nil, err
		}

		var items []interface{}
		for _, event := range historyEvents {
			items = append(items, event)
		}
		return items, token, nil
	}
}
