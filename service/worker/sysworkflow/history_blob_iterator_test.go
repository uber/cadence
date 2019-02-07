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

package sysworkflow

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	metricsMocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"testing"
)

const (
	testDomainID          = "test-domain-id"
	testWorkflowID        = "test-workflow-id"
	testRunID             = "test-run-id"
	testEventStoreVersion = "test-event-store-version"
	testBranchToken       = "test-branch-token"
	testLastFirstEventID  = "test-last-first-event-id"
	testDomain            = "test-domain"
	testClusterName       = "test-cluster-name"
)

type HistoryBlobIteratorSuite struct {
	*require.Assertions
	suite.Suite
	logger               bark.Logger
	metricsClient        *metricsMocks.Client
	mockHistoryManager   *mocks.HistoryManager
	mockHistoryV2Manager *mocks.HistoryV2Manager
}

func TestHistoryBlobIteratorSuite(t *testing.T) {
	suite.Run(t, new(HistoryBlobIteratorSuite))
}

func (s *HistoryBlobIteratorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.logger = bark.NewNopLogger()
	s.metricsClient = &metricsMocks.Client{}
	s.mockHistoryManager = &mocks.HistoryManager{}
	s.mockHistoryV2Manager = &mocks.HistoryV2Manager{}
}

func (s *HistoryBlobIteratorSuite) TestTest() {
	s.True(true)
}

func (s *HistoryBlobIteratorSuite) constructHistoryEvent(eventId int64) *shared.HistoryEvent {
	return &shared.HistoryEvent{
		EventId: common.Int64Ptr(eventId),
	}
}

func (s *HistoryBlobIteratorSuite) constructConfig(historyPageSize, targetArchivalBlobSize int) *Config {
	return &Config{
		HistoryPageSize:        dynamicconfig.GetIntPropertyFilteredByDomain(historyPageSize),
		TargetArchivalBlobSize: dynamicconfig.GetIntPropertyFilteredByDomain(targetArchivalBlobSize),
	}
}
