// Copyright (c) 2019 Uber Technologies, Inc.
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

package ndc

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	test "github.com/uber/cadence/common/testing"
	"github.com/uber/cadence/environment"
	"github.com/uber/cadence/host"
)

type (
	nDCIntegrationTestSuite struct {
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
		suite.Suite
		active     *host.TestCluster
		passive    *host.TestCluster
		generator  test.Generator
		serializer persistence.PayloadSerializer
		logger     log.Logger

		domainName string
		domainID   string
	}
)

var (
	clusterName              = []string{"active", "standby"}
	clusterReplicationConfig = []*workflow.ClusterReplicationConfiguration{
		{ClusterName: common.StringPtr(clusterName[0])},
		{ClusterName: common.StringPtr(clusterName[1])},
	}
)

func TestNDCIntegrationTestSuite(t *testing.T) {

	flag.Parse()
	suite.Run(t, new(nDCIntegrationTestSuite))
}

func (s *nDCIntegrationTestSuite) SetupSuite() {
	zapLogger, err := zap.NewDevelopment()
	// cannot use s.Nil since it is not initialized
	s.Require().NoError(err)
	s.serializer = persistence.NewPayloadSerializer()
	s.logger = loggerimpl.NewLogger(zapLogger)

	fileName := "../testdata/xdc_integration_test_clusters.yaml"
	if host.TestFlags.TestClusterConfigFile != "" {
		fileName = host.TestFlags.TestClusterConfigFile
	}
	environment.SetupEnv()

	confContent, err := ioutil.ReadFile(fileName)
	s.Require().NoError(err)
	confContent = []byte(os.ExpandEnv(string(confContent)))

	var clusterConfigs []*host.TestClusterConfig
	s.Require().NoError(yaml.Unmarshal(confContent, &clusterConfigs))
	clusterConfigs[0].WorkerConfig = &host.WorkerConfig{}
	clusterConfigs[1].WorkerConfig = &host.WorkerConfig{}

	cluster, err := host.NewCluster(clusterConfigs[0], s.logger.WithTags(tag.ClusterName(clusterName[0])))
	s.Require().NoError(err)
	s.active = cluster

	cluster, err = host.NewCluster(clusterConfigs[1], s.logger.WithTags(tag.ClusterName(clusterName[1])))
	s.Require().NoError(err)
	s.passive = cluster

	s.registerDomain()

	s.generator = test.InitializeHistoryEventGenerator(s.domainName)
}

func (s *nDCIntegrationTestSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func (s *nDCIntegrationTestSuite) TearDownSuite() {
	s.active.TearDownCluster()
	s.passive.TearDownCluster()
}

func (s *nDCIntegrationTestSuite) TestSingleBranch() {

	workflowID := uuid.New()

	workflowType := "event-generator-workflow-type"
	tasklist := "event-generator-taskList"

	historyClient := s.active.GetHistoryClient()

	versions := []int64{101, 1, 201}
	for _, version := range versions {
		runID := uuid.New()

		root := &test.NDCTestBranch{
			Batches: make([]test.NDCTestBatch, 0),
		}
		s.generator.Reset()
		fmt.Printf("##########\n")
		for s.generator.HasNextVertex() {
			events := s.generator.GetNextVertices()
			newBatch := test.NDCTestBatch{
				Events: events,
			}
			root.Batches = append(root.Batches, newBatch)
		}

		// TODO temporary code to generate version history
		//  we should generate version as part of modeled based testing
		versionHistory := persistence.NewVersionHistory(nil, nil)
		for _, batch := range root.Batches {
			for _, event := range batch.Events {
				print := func(value interface{}) string {
					bytes, _ := json.MarshalIndent(value, "", "  ")
					return string(bytes)
				}
				fmt.Printf("++++++++++\n")
				fmt.Printf("## SEEING:\n%v\n.", print(event.GetData().(*shared.HistoryEvent)))
				fmt.Printf("++++++++++\n")
				err := versionHistory.AddOrUpdateItem(
					persistence.NewVersionHistoryItem(
						event.GetData().(*shared.HistoryEvent).GetEventId(),
						event.GetData().(*shared.HistoryEvent).GetVersion(),
					))
				s.NoError(err)
			}
		}

		for _, batch := range root.Batches {

			// TODO temporary code to generate first event & version history
			//  we should generate these as part of modeled based testing
			lastEvent := batch.Events[len(batch.Events)-1].GetData().(*shared.HistoryEvent)
			newRunEventBlob, newRunVersionHistory := s.generateNewRunHistory(
				lastEvent, s.domainName, workflowID, runID, version, workflowType, tasklist,
			)

			// must serialize events batch after attempt on continue as new as generateNewRunHistory will
			// modify the NewExecutionRunId attr
			var events []*shared.HistoryEvent
			for _, event := range batch.Events {
				events = append(events, event.GetData().(*shared.HistoryEvent))
			}
			eventBlob, err := s.serializer.SerializeBatchEvents(events, common.EncodingTypeThriftRW)
			s.NoError(err)

			err = historyClient.ReplicateEventsV2(s.createContext(), &history.ReplicateEventsV2Request{
				DomainUUID: common.StringPtr(s.domainID),
				WorkflowExecution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
					RunId:      common.StringPtr(runID),
				},
				VersionHistoryItems:       s.toThriftVersionHistoryItems(versionHistory),
				Events:                    s.toThriftDataBlob(eventBlob),
				NewRunVersionHistoryItems: s.toThriftVersionHistoryItems(newRunVersionHistory),
				NewRunEvents:              s.toThriftDataBlob(newRunEventBlob),
				ResetWorkflow:             common.BoolPtr(false),
			})
			s.Nil(err, "Failed to replicate history event")
		}

		// get replicated history events from passive side
		passiveClient := s.active.GetFrontendClient()
		replicatedHistory, err := passiveClient.GetWorkflowExecutionHistory(
			s.createContext(),
			&shared.GetWorkflowExecutionHistoryRequest{
				Domain: common.StringPtr(s.domainName),
				Execution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
					RunId:      common.StringPtr(runID),
				},
				MaximumPageSize:        common.Int32Ptr(1000),
				NextPageToken:          nil,
				WaitForNewEvent:        common.BoolPtr(false),
				HistoryEventFilterType: shared.HistoryEventFilterTypeAllEvent.Ptr(),
			},
		)
		s.Nil(err, "Failed to get history event from passive side")

		// compare origin events with replicated events
		batchIndex := 0
		batch := root.Batches[batchIndex].Events
		eventIndex := 0
		for _, event := range replicatedHistory.GetHistory().GetEvents() {
			if eventIndex >= len(batch) {
				batchIndex++
				batch = root.Batches[batchIndex].Events
				eventIndex = 0
			}
			originEvent := batch[eventIndex].GetData().(shared.HistoryEvent)
			eventIndex++
			s.Equal(originEvent.GetEventType().String(), event.GetEventType().String(), "The replicated event and the origin event are not the same")
		}
	}
}

func (s *nDCIntegrationTestSuite) registerDomain() {
	s.domainName = "test-simple-workflow-ndc-" + common.GenerateRandomString(5)
	client1 := s.active.GetFrontendClient() // active
	err := client1.RegisterDomain(s.createContext(), &shared.RegisterDomainRequest{
		Name:                                   common.StringPtr(s.domainName),
		IsGlobalDomain:                         common.BoolPtr(true),
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      common.StringPtr(clusterName[0]),
		WorkflowExecutionRetentionPeriodInDays: common.Int32Ptr(1),
	})
	s.Require().NoError(err)

	descReq := &shared.DescribeDomainRequest{
		Name: common.StringPtr(s.domainName),
	}
	resp, err := client1.DescribeDomain(s.createContext(), descReq)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.domainID = resp.GetDomainInfo().GetUUID()
	// Wait for domain cache to pick the change
	time.Sleep(2 * cache.DomainCacheRefreshInterval)

	s.logger.Info(fmt.Sprintf("Domain name: %v - ID: %v", s.domainName, s.domainID))
}

func (s *nDCIntegrationTestSuite) generateNewRunHistory(
	event *shared.HistoryEvent,
	domain string,
	workflowID string,
	runID string,
	version int64,
	workflowType string,
	taskList string,
) (*persistence.DataBlob, *persistence.VersionHistory) {

	// TODO temporary code to generate first event & version history
	//  we should generate these as part of modeled based testing

	if event.GetWorkflowExecutionContinuedAsNewEventAttributes() == nil {
		return nil, nil
	}

	event.WorkflowExecutionContinuedAsNewEventAttributes.NewExecutionRunId = common.StringPtr(uuid.New())

	newRunFirstEvent := &shared.HistoryEvent{
		EventId:   common.Int64Ptr(common.FirstEventID),
		Timestamp: common.Int64Ptr(time.Now().UnixNano()),
		EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionStarted),
		Version:   common.Int64Ptr(version),
		TaskId:    common.Int64Ptr(1),
		WorkflowExecutionStartedEventAttributes: &shared.WorkflowExecutionStartedEventAttributes{
			WorkflowType:         common.WorkflowTypePtr(shared.WorkflowType{Name: common.StringPtr(workflowType)}),
			ParentWorkflowDomain: common.StringPtr(domain),
			ParentWorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: common.StringPtr(uuid.New()),
				RunId:      common.StringPtr(uuid.New()),
			},
			ParentInitiatedEventId: common.Int64Ptr(event.GetEventId()),
			TaskList: common.TaskListPtr(shared.TaskList{
				Name: common.StringPtr(taskList),
				Kind: common.TaskListKindPtr(shared.TaskListKindNormal),
			}),
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
			ContinuedExecutionRunId:             common.StringPtr(runID),
			Initiator:                           shared.ContinueAsNewInitiatorCronSchedule.Ptr(),
			OriginalExecutionRunId:              common.StringPtr(runID),
			Identity:                            common.StringPtr("NDC-test"),
			FirstExecutionRunId:                 common.StringPtr(runID),
			Attempt:                             common.Int32Ptr(0),
			ExpirationTimestamp:                 common.Int64Ptr(time.Now().Add(time.Minute).UnixNano()),
		},
	}

	eventBlob, err := s.serializer.SerializeBatchEvents([]*shared.HistoryEvent{newRunFirstEvent}, common.EncodingTypeThriftRW)
	s.NoError(err)

	newRunVersionHistory := persistence.NewVersionHistory(nil, nil)
	err = newRunVersionHistory.AddOrUpdateItem(persistence.NewVersionHistoryItem(
		newRunFirstEvent.GetEventId(),
		newRunFirstEvent.GetVersion(),
	))
	s.NoError(err)

	return eventBlob, newRunVersionHistory
}

func (s *nDCIntegrationTestSuite) toThriftDataBlob(
	blob *persistence.DataBlob,
) *shared.DataBlob {

	if blob == nil {
		return nil
	}

	var encodingType shared.EncodingType
	switch blob.GetEncoding() {
	case common.EncodingTypeThriftRW:
		encodingType = shared.EncodingTypeThriftRW
	case common.EncodingTypeJSON,
		common.EncodingTypeGob,
		common.EncodingTypeUnknown,
		common.EncodingTypeEmpty:
		panic(fmt.Sprintf("unsupported encoding type: %v", blob.GetEncoding()))
	default:
		panic(fmt.Sprintf("unknown encoding type: %v", blob.GetEncoding()))
	}

	return &shared.DataBlob{
		EncodingType: encodingType.Ptr(),
		Data:         blob.Data,
	}
}

func (s *nDCIntegrationTestSuite) toThriftVersionHistoryItems(
	versionHistory *persistence.VersionHistory,
) []*shared.VersionHistoryItem {
	if versionHistory == nil {
		return nil
	}

	return versionHistory.ToThrift().Items
}

func (s *nDCIntegrationTestSuite) createContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 90*time.Second)
	return ctx
}
