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
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	result := history.GetHistory()
	s.Equal(items, result)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_CreateNewItem() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	err := history.Update(VersionHistoryItem{
		eventID: 8,
		version: 5,
	})

	s.NoError(err)
	s.Equal(len(history.GetHistory()), len(items)+1)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_UpdateEventID() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	err := history.Update(VersionHistoryItem{
		eventID: 8,
		version: 4,
	})

	s.NoError(err)
	s.Equal(len(history.GetHistory()), len(items))
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_Failed_VersionNotMatch() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	err := history.Update(VersionHistoryItem{
		eventID: 8,
		version: 3,
	})

	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestUpdateVersionHistory_Failed_EventIDNotIncrease() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	err := history.Update(VersionHistoryItem{
		eventID: 5,
		version: 4,
	})

	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestIsAppendable_True() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	appendItem := VersionHistoryItem{
		eventID: 6,
		version: 4,
	}

	s.True(history.IsAppendable(appendItem))
}

func (s *versionHistoryStoreSuite) TestIsAppendable_False_VersionNotMatch() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	appendItem := VersionHistoryItem{
		eventID: 6,
		version: 7,
	}

	s.False(history.IsAppendable(appendItem))
}

func (s *versionHistoryStoreSuite) TestIsAppendable_False_EventIDNotMatch() {
	items := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 6, version: 4},
	}

	history := NewVersionHistory(items...)
	appendItem := VersionHistoryItem{
		eventID: 7,
		version: 4,
	}

	s.False(history.IsAppendable(appendItem))
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem() {
	localItems := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 5, version: 4},
		{eventID: 7, version: 6},
		{eventID: 9, version: 10},
	}
	remoteItems := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 7, version: 4},
		{eventID: 8, version: 8},
		{eventID: 11, version: 12},
	}
	local := NewVersionHistory(localItems...)
	remote := NewVersionHistory(remoteItems...)
	item, err := local.FindLowestCommonVersionHistoryItem(remote)
	s.NoError(err)
	s.Equal(int64(5), item.eventID)
	s.Equal(int64(4), item.version)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistoryItem_Error() {
	localItems := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 5, version: 4},
		{eventID: 7, version: 6},
		{eventID: 9, version: 10},
	}
	remoteItems := []VersionHistoryItem{
		{eventID: 3, version: 1},
		{eventID: 7, version: 2},
		{eventID: 8, version: 3},
	}
	local := NewVersionHistory(localItems...)
	remote := NewVersionHistory(remoteItems...)
	_, err := local.FindLowestCommonVersionHistoryItem(remote)
	s.Error(err)
}

func (s *versionHistoryStoreSuite) TestFindLowestCommonVersionHistory() {
	localItems1 := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 5, version: 4},
		{eventID: 7, version: 6},
		{eventID: 9, version: 10},
	}
	localItems2 := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 8, version: 4},
		{eventID: 9, version: 6},
	}
	remoteItems := []VersionHistoryItem{
		{eventID: 3, version: 0},
		{eventID: 7, version: 4},
		{eventID: 8, version: 8},
		{eventID: 11, version: 12},
	}
	local1 := NewVersionHistory(localItems1...)
	local2 := NewVersionHistory(localItems2...)
	remote := NewVersionHistory(remoteItems...)
	histories := NewVersionHistories(local1, local2)
	item, history, err := histories.FindLowestCommonVersionHistory(remote)
	s.NoError(err)
	s.Equal(*history, local2)
	s.Equal(int64(7), item.eventID)
	s.Equal(int64(4), item.version)
}
