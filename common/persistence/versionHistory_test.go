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

package persistence

import (
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"testing"
)

type (
	versionHistoryStoreSuite struct {
		suite.Suite
	}
)

func TestVersionHistoryStore(t *testing.T) {
	s := new(versionHistoryStoreSuite)
	suite.Run(t, s)
}

func (s *versionHistoryStoreSuite) TestNewVersionHistory() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	result := history.History
	s.Equal(items, result)
}

func (s *versionHistoryStoreSuite) TestNewVersionHistory_Panic() {
	items := []*shared.VersionHistoryItem{}

	expectedPanic := func() {
		NewVersionHistory(items)
	}
	s.Panics(expectedPanic)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_CreateNewItem() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	err := UpdateVersionHistory(history, shared.VersionHistoryItem{
		EventID: common.Int64Ptr(8),
		Version: common.Int64Ptr(5),
	})

	s.NoError(err)
	s.Equal(len(history.History), len(items)+1)
	s.Equal(int64(8), history.History[2].EventID)
	s.Equal(int64(5), history.History[2].Version)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_UpdateEventID() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	err := UpdateVersionHistory(history, shared.VersionHistoryItem{
		EventID: common.Int64Ptr(8),
		Version: common.Int64Ptr(4),
	})

	s.NoError(err)
	s.Equal(len(history.History), len(items))
	s.Equal(int64(8), history.History[1].EventID)
	s.Equal(int64(4), history.History[1].Version)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_Failed_LowerVersion() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	err := UpdateVersionHistory(history, shared.VersionHistoryItem{
		EventID: common.Int64Ptr(8),
		Version: common.Int64Ptr(3),
	})

	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_Failed_EventIDNotIncrease() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	err := UpdateVersionHistory(history, shared.VersionHistoryItem{
		EventID: common.Int64Ptr(5),
		Version: common.Int64Ptr(4),
	})

	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_Failed_EventIDMatch_VersionNotMatch() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	err := UpdateVersionHistory(history, shared.VersionHistoryItem{
		EventID: common.Int64Ptr(6),
		Version: common.Int64Ptr(7),
	})

	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestIsAppendable_True() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	appendItem := shared.VersionHistoryItem{
		EventID: common.Int64Ptr(6),
		Version: common.Int64Ptr(4),
	}

	s.True(IsAppendable(history, appendItem))
}

func (s *versionHistoryStoreSuite) TestIsAppendable_False_VersionNotMatch() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	appendItem := shared.VersionHistoryItem{
		EventID: common.Int64Ptr(6),
		Version: common.Int64Ptr(7),
	}

	s.False(IsAppendable(history, appendItem))
}

func (s *versionHistoryStoreSuite) TestIsAppendable_False_EventIDNotMatch() {
	items := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(4)},
	}

	history := NewVersionHistory(items)
	appendItem := shared.VersionHistoryItem{
		EventID: common.Int64Ptr(7),
		Version: common.Int64Ptr(4),
	}

	s.False(IsAppendable(history, appendItem))
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_ReturnLocal() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(8)},
		{EventID: common.Int64Ptr(11), Version: common.Int64Ptr(12)},
	}
	local := NewVersionHistory(localItems)
	remote := NewVersionHistory(remoteItems)
	item, err := FindLowestCommonVersionHistoryItem(local, remote)
	s.NoError(err)
	s.Equal(int64(5), item.EventID)
	s.Equal(int64(4), item.Version)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_ReturnRemote() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(11), Version: common.Int64Ptr(12)},
	}
	local := NewVersionHistory(localItems)
	remote := NewVersionHistory(remoteItems)
	item, err := FindLowestCommonVersionHistoryItem(local, remote)
	s.NoError(err)
	s.Equal(int64(6), item.EventID)
	s.Equal(int64(6), item.Version)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_Error_NoLCA() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(1)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(2)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(3)},
	}
	local := NewVersionHistory(localItems)
	remote := NewVersionHistory(remoteItems)
	_, err := FindLowestCommonVersionHistoryItem(local, remote)
	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_Error_InvalidInput() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(1)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(2)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(3)},
	}
	local := NewVersionHistory(localItems)
	remote := NewVersionHistory(remoteItems)
	_, err := FindLowestCommonVersionHistoryItem(local, remote)
	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_Error_NilInput() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	local := NewVersionHistory(localItems)
	remote := NewVersionHistory([]*shared.VersionHistoryItem{})
	_, err := FindLowestCommonVersionHistoryItem(local, remote)
	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestNewVersionHistories_Panic() {
	expectedPanic := func() { NewVersionHistories([]*shared.VersionHistory{}) }
	s.Panics(expectedPanic)
}

func (s *versionHistoryStoreSuite) TestNewVersionHistories() {
	localItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	local := NewVersionHistory(localItems)
	histories := NewVersionHistories([]*shared.VersionHistory{&local})
	s.NotNil(histories)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistory_UpdateExistingHistory() {
	localItems1 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	localItems2 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(6)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(10), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(11), Version: common.Int64Ptr(12)},
	}
	local1 := NewVersionHistory(localItems1)
	local2 := NewVersionHistory(localItems2)
	remote := NewVersionHistory(remoteItems)
	histories := NewVersionHistories([]*shared.VersionHistory{&local1, &local2})
	item, history, err := FindLowestCommonVersionHistory(histories, remote)
	s.NoError(err)
	s.Equal(history, local2)
	s.Equal(int64(9), item.EventID)
	s.Equal(int64(6), item.Version)

	err = AddHistory(histories, *item, *history, remote)
	s.NoError(err)
	s.Equal(histories.Histories[1], remote)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistory_ForkNewHistory() {
	localItems1 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	localItems2 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(6)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(7)},
		{EventID: common.Int64Ptr(10), Version: common.Int64Ptr(12)},
	}
	local1 := NewVersionHistory(localItems1)
	local2 := NewVersionHistory(localItems2)
	remote := NewVersionHistory(remoteItems)
	histories := NewVersionHistories([]*shared.VersionHistory{&local1, &local2})
	item, history, err := FindLowestCommonVersionHistory(histories, remote)
	s.NoError(err)
	s.Equal(int64(3), item.EventID)
	s.Equal(int64(0), item.Version)

	err = AddHistory(histories, *item, *history, remote)
	s.NoError(err)
	s.Equal(3, len(histories.Histories))
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistory_Error() {
	localItems1 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(5), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(7), Version: common.Int64Ptr(6)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(10)},
	}
	localItems2 := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(0)},
		{EventID: common.Int64Ptr(8), Version: common.Int64Ptr(4)},
		{EventID: common.Int64Ptr(9), Version: common.Int64Ptr(6)},
	}
	remoteItems := []*shared.VersionHistoryItem{
		{EventID: common.Int64Ptr(3), Version: common.Int64Ptr(1)},
		{EventID: common.Int64Ptr(6), Version: common.Int64Ptr(7)},
		{EventID: common.Int64Ptr(10), Version: common.Int64Ptr(12)},
	}
	local1 := NewVersionHistory(localItems1)
	local2 := NewVersionHistory(localItems2)
	remote := NewVersionHistory(remoteItems)
	histories := NewVersionHistories([]*shared.VersionHistory{&local1, &local2})
	_, _, err := FindLowestCommonVersionHistory(histories, remote)
	s.Error(err)
}
