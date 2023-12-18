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

package archiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/types"
)

type UtilSuite struct {
	*require.Assertions
	suite.Suite
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *UtilSuite) TestHistoryMutated() {
	testCases := []struct {
		historyBatches []*types.History
		request        *ArchiveHistoryRequest
		isLast         bool
		isMutated      bool
	}{
		{
			historyBatches: []*types.History{
				{
					Events: []*types.HistoryEvent{
						{
							Version: 15,
						},
					},
				},
			},
			request: &ArchiveHistoryRequest{
				CloseFailoverVersion: 3,
			},
			isMutated: true,
		},
		{
			historyBatches: []*types.History{
				{
					Events: []*types.HistoryEvent{
						{
							ID:      33,
							Version: 10,
						},
					},
				},
				{
					Events: []*types.HistoryEvent{
						{
							ID:      49,
							Version: 10,
						},
						{
							ID:      50,
							Version: 10,
						},
					},
				},
			},
			request: &ArchiveHistoryRequest{
				CloseFailoverVersion: 10,
				NextEventID:          34,
			},
			isLast:    true,
			isMutated: true,
		},
		{
			historyBatches: []*types.History{
				{
					Events: []*types.HistoryEvent{
						{
							Version: 9,
						},
					},
				},
			},
			request: &ArchiveHistoryRequest{
				CloseFailoverVersion: 10,
			},
			isLast:    true,
			isMutated: true,
		},
		{
			historyBatches: []*types.History{
				{
					Events: []*types.HistoryEvent{
						{
							ID:      20,
							Version: 10,
						},
					},
				},
				{
					Events: []*types.HistoryEvent{
						{
							ID:      33,
							Version: 10,
						},
					},
				},
			},
			request: &ArchiveHistoryRequest{
				CloseFailoverVersion: 10,
				NextEventID:          34,
			},
			isLast:    true,
			isMutated: false,
		},
	}
	for _, tc := range testCases {
		s.Equal(tc.isMutated, IsHistoryMutated(tc.request, tc.historyBatches, tc.isLast, testlogger.New(s.T())))
	}
}
