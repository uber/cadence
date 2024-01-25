// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package execution

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/shard"
)

func TestReplicateDecisionTaskCompletedEvent(t *testing.T) {
	mockShard := shard.NewTestContext(
		t,
		gomock.NewController(t),
		&persistence.ShardInfo{
			ShardID:          0,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config.NewForTest(),
	)
	mockShard.GetConfig().MutableStateChecksumGenProbability = func(domain string) int { return 100 }
	mockShard.GetConfig().MutableStateChecksumVerifyProbability = func(domain string) int { return 100 }
	logger := mockShard.GetLogger()
	mockShard.Resource.DomainCache.EXPECT().GetDomainID(constants.TestDomainName).Return(constants.TestDomainID, nil).AnyTimes()

	m := &mutableStateDecisionTaskManagerImpl{
		msb: newMutableStateBuilder(mockShard, logger, constants.TestLocalDomainEntry),
	}
	eventType := types.EventTypeActivityTaskCompleted
	e := &types.HistoryEvent{
		ID:        1,
		EventType: &eventType,
	}
	err := m.ReplicateDecisionTaskCompletedEvent(e)
	assert.NoError(t, err)

	// test when domainEntry is missed
	m.msb.domainEntry = nil
	err = m.ReplicateDecisionTaskCompletedEvent(e)
	assert.NoError(t, err)

	// test when config is nil
	m.msb = newMutableStateBuilder(mockShard, logger, constants.TestLocalDomainEntry)
	m.msb.config = nil
	err = m.ReplicateDecisionTaskCompletedEvent(e)
	assert.NoError(t, err)
}
