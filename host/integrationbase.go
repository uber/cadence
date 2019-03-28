// Copyright (c) 2016 Uber Technologies, Inc.
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

package host

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence/persistence-tests"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
)

type (
	// IntegrationBase is a base struct for integration tests
	IntegrationBase struct {
		suite.Suite

		testCluster       *TestCluster
		engine            FrontendClient
		adminClient       AdminClient
		Logger            bark.Logger
		domainName        string
		foreignDomainName string
	}
)

func (s *IntegrationBase) setupSuite(enableGlobalDomain bool, isMasterCluster bool, enableWorker bool, enableArchival bool) {
	s.setupLogger()

	if *frontendAddress != "" {
		s.Logger.WithField("address", *frontendAddress).Info("Running integration test against specified frontend")
		channel, err := tchannel.NewChannelTransport(tchannel.ServiceName("cadence-frontend"))
		s.Require().NoError(err)
		dispatcher := yarpc.NewDispatcher(yarpc.Config{
			Name: "unittest",
			Outbounds: yarpc.Outbounds{
				"cadence-frontend": {Unary: channel.NewSingleOutbound(*frontendAddress)},
			},
		})
		if err := dispatcher.Start(); err != nil {
			s.Logger.WithField("error", err).Fatal("Failed to create outbound transport channel")
		}

		s.engine = NewFrontendClient(dispatcher)
		s.adminClient = NewAdminClient(dispatcher)
	} else {
		s.Logger.Info("Running integration test against test cluster")
		cluster, err := s.setupCadenceHost(enableGlobalDomain, isMasterCluster, enableWorker, enableArchival)
		s.Require().NoError(err)
		s.testCluster = cluster
		s.engine = s.testCluster.GetFrontendClient()
		s.adminClient = s.testCluster.GetAdminClient()
	}

	var retentionDays int32 = 1
	s.domainName = s.randomizeStr("integration-test-domain")
	s.Require().NoError(s.testCluster.GetFrontendClient().RegisterDomain(createContext(), &workflow.RegisterDomainRequest{
		Name: &s.domainName,
		WorkflowExecutionRetentionPeriodInDays:&retentionDays,
	}))

	s.foreignDomainName = s.randomizeStr("integration-foreign-test-domain")
	s.Require().NoError(s.testCluster.GetFrontendClient().RegisterDomain(createContext(), &workflow.RegisterDomainRequest{
		Name: &s.foreignDomainName,
		WorkflowExecutionRetentionPeriodInDays:&retentionDays,
	}))
}

func (s *IntegrationBase) setupLogger() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

	logger := log.New()
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	logger.Formatter = formatter
	s.Logger = bark.NewLoggerFromLogrus(logger)
}

func (s *IntegrationBase) setupCadenceHost(enableGlobalDomain bool, isMasterCluster bool, enableWorker bool, enableArchival bool) (*TestCluster, error) {
	persistOptions := &persistencetests.TestBaseOptions{
		EnableGlobalDomain: enableGlobalDomain,
		IsMasterCluster:    isMasterCluster,
		EnableArchival:     enableArchival,
	}

	options := &TestClusterOptions{
		PersistOptions:   persistOptions,
		EnableWorker:     enableWorker,
		MessagingClient:  mocks.NewMockMessagingClient(&mocks.KafkaProducer{}, nil),
		NumHistoryShards: testNumberOfHistoryShards,
		EnableEventsV2:   *EnableEventsV2,
	}
	return NewCluster(options, s.Logger)
}

func (s *IntegrationBase) tearDownSuite() {
	if s.testCluster != nil {
		s.testCluster.TearDownCluster()
	}
}

func (s *IntegrationBase) registerDomain(domain string, desc string, archivalStatus workflow.ArchivalStatus, archivalBucket string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.engine.RegisterDomain(ctx, &workflow.RegisterDomainRequest{
		Name:               &domain,
		Description:        &desc,
		ArchivalStatus:     &archivalStatus,
		ArchivalBucketName: &archivalBucket,
	})
}

func (s *IntegrationBase) describeDomain(domain string) (*workflow.DescribeDomainResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.engine.DescribeDomain(ctx, &workflow.DescribeDomainRequest{
		Name: &domain,
	})
}

func (s *IntegrationBase) randomizeStr(id string) string {
	return fmt.Sprintf("%v-%v", id, uuid.New())
}

func (s *IntegrationBase) printWorkflowHistory(domain string, execution *workflow.WorkflowExecution) {
	events := s.getHistory(domain, execution)
	history := &workflow.History{}
	history.Events = events
	common.PrettyPrintHistory(history, s.Logger)
}

func (s *IntegrationBase) getHistory(domain string, execution *workflow.WorkflowExecution) []*workflow.HistoryEvent {
	historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &workflow.GetWorkflowExecutionHistoryRequest{
		Domain:          common.StringPtr(domain),
		Execution:       execution,
		MaximumPageSize: common.Int32Ptr(5), // Use small page size to force pagination code path
	})
	s.Require().NoError(err)

	events := historyResponse.History.Events
	for historyResponse.NextPageToken != nil {
		historyResponse, err = s.engine.GetWorkflowExecutionHistory(createContext(), &workflow.GetWorkflowExecutionHistoryRequest{
			Domain:        common.StringPtr(domain),
			Execution:     execution,
			NextPageToken: historyResponse.NextPageToken,
		})
		s.Require().NoError(err)
		events = append(events, historyResponse.History.Events...)
	}

	return events
}
