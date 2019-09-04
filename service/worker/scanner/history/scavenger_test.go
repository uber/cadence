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

package history

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/.gen/go/history/historyservicetest"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	p "github.com/uber/cadence/common/persistence"
	"go.uber.org/zap"
	"testing"
	"time"
)

type (
	ScavengerTestSuite struct {
		suite.Suite
		logger        log.Logger
		metric        metrics.Client
	}
)

func TestScavengerTestSuite(t *testing.T) {
	suite.Run(t, new(ScavengerTestSuite))
}

func (s *ScavengerTestSuite) SetupTest() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		s.Require().NoError(err)
	}
	s.logger = loggerimpl.NewLogger(zapLogger)
	s.metric = metrics.NewClient(tally.NoopScope, metrics.Worker)
}

func (s *ScavengerTestSuite) createTestScavenger( rps int) (*mocks.HistoryV2Manager, *historyservicetest.MockClient, *Scavenger, *gomock.Controller){
	db := &mocks.HistoryV2Manager{}
	controller := gomock.NewController(s.T())
	workflowClient := historyservicetest.NewMockClient(controller)
	scvgr := NewScavenger(db, 100, workflowClient, ScavengerHeartbeatDetails{}, s.metric, s.logger )
	scvgr.isInTest = true
	return db, workflowClient, scvgr, controller
}

func (s *ScavengerTestSuite) TestAllSkipTasksTwoPages() {
	db, _, scvgr, controller := s.createTestScavenger( 100)
	defer controller.Finish()
	db.On("GetAllHistoryTreeBranches", &p.GetAllHistoryTreeBranchesRequest{
		PageSize:pageSize,
	}).Return(&p.GetAllHistoryTreeBranchesResponse{
		NextPageToken:[]byte("page1"),
		Branches:[]p.HistoryBranchDetail{
			{
				TreeID:"treeID1",
				BranchID:"branchID1",
				ForkTime:time.Now(),
				Info:p.BuildHistoryGarbageCleanupInfo("domainID1", "workflowID1", "runID1"),
			},
			{
				TreeID:"treeID2",
				BranchID:"branchID2",
				ForkTime:time.Now(),
				Info:p.BuildHistoryGarbageCleanupInfo("domainID2", "workflowID2", "runID2"),
			},
		},
	}, nil)
	db.On("GetAllHistoryTreeBranches", &p.GetAllHistoryTreeBranchesRequest{
		PageSize:pageSize,
		NextPageToken:[]byte("page1"),
	}).Return(&p.GetAllHistoryTreeBranchesResponse{
		Branches:[]p.HistoryBranchDetail{
			{
				TreeID:"treeID3",
				BranchID:"branchID3",
				ForkTime:time.Now(),
				Info:p.BuildHistoryGarbageCleanupInfo("domainID3", "workflowID3", "runID3"),
			},
			{
				TreeID:"treeID4",
				BranchID:"branchID4",
				ForkTime:time.Now(),
				Info:p.BuildHistoryGarbageCleanupInfo("domainID4", "workflowID4", "runID4"),
			},
		},
	}, nil)

	hbd, err := scvgr.Run(context.Background())
	s.Nil(err)
	s.Equal(4, hbd.SkipCount)
}
