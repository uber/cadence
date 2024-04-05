// Copyright (c) 2024 Uber Technologies, Inc.
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

package domain

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type HandlerSuite struct {
	suite.Suite
	controller           *gomock.Controller
	mockDomainMgr        *persistence.MockDomainManager
	mockClusterMetadata  cluster.Metadata
	mockArchiverProvider provider.ArchiverProvider
	handler              Handler
	context              context.Context
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

func (s *HandlerSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.context = context.Background()
	s.mockDomainMgr = persistence.NewMockDomainManager(s.controller)

	s.mockClusterMetadata = cluster.GetTestClusterMetadata(true)

	mockDC := dynamicconfig.NewCollection(dynamicconfig.NewNopClient(), log.NewNoop())
	testConfig := Config{
		MinRetentionDays:       dynamicconfig.GetIntPropertyFn(1),
		MaxRetentionDays:       dynamicconfig.GetIntPropertyFn(5),
		RequiredDomainDataKeys: nil,
		MaxBadBinaryCount:      nil,
		FailoverCoolDown:       nil,
	}

	domainDefaults := &config.ArchivalDomainDefaults{
		History: config.HistoryArchivalDomainDefaults{
			Status: "Disabled",
			URI:    "https://history.example.com",
		},
		Visibility: config.VisibilityArchivalDomainDefaults{
			Status: "Disabled",
			URI:    "https://visibility.example.com",
		},
	}

	archivalMetadata := archiver.NewArchivalMetadata(
		mockDC,
		"Disabled", // historyStatus
		false,      // historyReadEnabled
		"Disabled", // visibilityStatus
		false,      // visibilityReadEnabled
		domainDefaults,
	)

	s.mockClusterMetadata = cluster.GetTestClusterMetadata(true)
	s.mockArchiverProvider = provider.NewArchiverProvider(nil, nil)

	s.handler = NewHandler(
		testConfig,
		log.NewNoop(),
		s.mockDomainMgr,
		s.mockClusterMetadata,
		nil,
		archivalMetadata,
		s.mockArchiverProvider,
		clock.NewMockedTimeSource(),
	)
}

func (s *HandlerSuite) TestRegisterDomain() {
	tests := []struct {
		name          string
		request       *types.RegisterDomainRequest
		setupMocks    func(t *testing.T, request *types.RegisterDomainRequest)
		expectedError bool
	}{
		{
			name: "register new domain successfully",
			request: &types.RegisterDomainRequest{
				Name:                                   "test-domain",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			setupMocks: func(t *testing.T, request *types.RegisterDomainRequest) {
				// Mocking the behavior when the domain does not exist yet.
				s.mockDomainMgr.EXPECT().
					GetDomain(s.context, &persistence.GetDomainRequest{Name: request.GetName()}).
					Return(nil, &types.EntityNotExistsError{})

				// Mocking successful domain creation.
				s.mockDomainMgr.EXPECT().
					CreateDomain(s.context, gomock.Any()).
					Return(&persistence.CreateDomainResponse{ID: "test-domain-id"}, nil)
			},
			expectedError: false,
		},
		{
			name: "domain already exists",
			request: &types.RegisterDomainRequest{
				Name:                                   "existing-domain",
				WorkflowExecutionRetentionPeriodInDays: 3,
			},
			setupMocks: func(t *testing.T, request *types.RegisterDomainRequest) {
				s.mockDomainMgr.EXPECT().
					GetDomain(s.context, &persistence.GetDomainRequest{Name: request.GetName()}).
					Return(&persistence.GetDomainResponse{}, nil)
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.setupMocks(t, tc.request)

			err := s.handler.RegisterDomain(s.context, tc.request)

			if tc.expectedError {
				s.NotNil(err, "Expected an error but got none")
			} else {
				s.Nil(err, "Expected no error but got one")
			}
		})
	}
}
