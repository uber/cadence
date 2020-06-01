// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package failover

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/.gen/go/history/historyservicetest"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/resource"
)

type (
	coordinatorSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockResource        *resource.Test
		mockMetadataManager *mmocks.MetadataManager
		historyClient       *historyservicetest.MockClient
		config              *config.Config
		coordinator         *coordinatorImpl
	}
)

func TestCoordinatorSuite(t *testing.T) {
	s := new(coordinatorSuite)
	suite.Run(t, s)
}

func (s *coordinatorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.controller, metrics.History)
	s.mockMetadataManager = s.mockResource.MetadataMgr
	s.historyClient = s.mockResource.HistoryClient
	s.config = config.NewForTest()
	s.config.NumberOfShards = 1
	s.config.FailoverMarkerHeartbeatInterval = dynamicconfig.GetDurationPropertyFn(10 * time.Millisecond)
	s.config.FailoverMarkerHeartbeatTimerJitterCoefficient = dynamicconfig.GetFloatPropertyFn(0.01)

	s.coordinator = NewCoordinator(
		s.mockMetadataManager,
		s.historyClient,
		s.mockResource.GetTimeSource(),
		s.config,
		s.mockResource.GetMetricsClient(),
		s.mockResource.GetLogger(),
	).(*coordinatorImpl)
	s.coordinator.Start()
}

func (s *coordinatorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockResource.Finish(s.T())
	s.coordinator.Stop()
}

func (s *coordinatorSuite) TestHeartbeatFailoverMarkers() {
	respCh := s.coordinator.HeartbeatFailoverMarkers(
		1,
		[]*replicator.FailoverMarkerAttributes{
			{
				DomainID:        common.StringPtr(uuid.New()),
				FailoverVersion: common.Int64Ptr(1),
				CreationTime:    common.Int64Ptr(1),
			},
		},
	)

	err := <-respCh
	s.NoError(err)
}
