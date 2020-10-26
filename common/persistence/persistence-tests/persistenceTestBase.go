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

package persistencetests

import (
	"context"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/replicator"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/persistence/client"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/service/config"
)

type (
	// TransferTaskIDGenerator generates IDs for transfer tasks written by helper methods
	TransferTaskIDGenerator interface {
		GenerateTransferTaskID() (int64, error)
	}

	// TestBaseOptions options to configure workflow test base.
	TestBaseOptions struct {
		SQLDBPluginName string
		DBName          string
		DBUsername      string
		DBPassword      string
		DBHost          string
		DBPort          int              `yaml:"-"`
		StoreType       string           `yaml:"-"`
		SchemaDir       string           `yaml:"-"`
		ClusterMetadata cluster.Metadata `yaml:"-"`
	}

	// TestBase wraps the base setup needed to create workflows over persistence layer.
	TestBase struct {
		suite.Suite
		ShardMgr               p.ShardManager
		ExecutionMgrFactory    client.Factory
		ExecutionManager       p.ExecutionManager
		TaskMgr                p.TaskManager
		HistoryV2Mgr           p.HistoryManager
		MetadataManager        p.MetadataManager
		VisibilityMgr          p.VisibilityManager
		DomainReplicationQueue p.DomainReplicationQueue
		ShardInfo              *p.ShardInfo
		TaskIDGenerator        TransferTaskIDGenerator
		ClusterMetadata        cluster.Metadata
		ReadLevel              int64
		ReplicationReadLevel   int64
		DefaultTestCluster     PersistenceTestCluster
		VisibilityTestCluster  PersistenceTestCluster
		Logger                 log.Logger
		PayloadSerializer      p.PayloadSerializer
	}

	// PersistenceTestCluster exposes management operations on a database
	PersistenceTestCluster interface {
		SetupTestDatabase()
		TearDownTestDatabase()
		Config() config.Persistence
	}

	// TestTransferTaskIDGenerator helper
	TestTransferTaskIDGenerator struct {
		seqNum int64
	}
)

const (
	defaultScheduleToStartTimeout = 111
)

// NewTestBaseWithCassandra returns a persistence test base backed by cassandra datastore
func NewTestBaseWithCassandra(options *TestBaseOptions) TestBase {
	if options.DBName == "" {
		options.DBName = "test_" + GenerateRandomDBName(10)
	}
	testCluster := cassandra.NewTestCluster(options.DBName, options.DBUsername, options.DBPassword, options.DBHost, options.DBPort, options.SchemaDir)
	return newTestBase(options, testCluster)
}

// NewTestBaseWithSQL returns a new persistence test base backed by SQL
func NewTestBaseWithSQL(options *TestBaseOptions) TestBase {
	if options.DBName == "" {
		options.DBName = "test_" + GenerateRandomDBName(10)
	}
	testCluster := sql.NewTestCluster(options.SQLDBPluginName, options.DBName, options.DBUsername, options.DBPassword, options.DBHost, options.DBPort, options.SchemaDir)
	return newTestBase(options, testCluster)
}

// NewTestBase returns a persistence test base backed by either cassandra or sql
func NewTestBase(options *TestBaseOptions) TestBase {
	switch options.StoreType {
	case config.StoreTypeSQL:
		return NewTestBaseWithSQL(options)
	case config.StoreTypeCassandra:
		return NewTestBaseWithCassandra(options)
	default:
		panic("invalid storeType " + options.StoreType)
	}
}

func newTestBase(options *TestBaseOptions, testCluster PersistenceTestCluster) TestBase {
	metadata := options.ClusterMetadata
	if metadata == nil {
		metadata = cluster.GetTestClusterMetadata(false, false)
	}
	options.ClusterMetadata = metadata
	base := TestBase{
		DefaultTestCluster:    testCluster,
		VisibilityTestCluster: testCluster,
		ClusterMetadata:       metadata,
		PayloadSerializer:     p.NewPayloadSerializer(),
	}
	logger, err := loggerimpl.NewDevelopment()
	if err != nil {
		panic(err)
	}
	base.Logger = logger
	return base
}

// Config returns the persistence configuration for this test
func (s *TestBase) Config() config.Persistence {
	cfg := s.DefaultTestCluster.Config()
	if s.DefaultTestCluster == s.VisibilityTestCluster {
		return cfg
	}
	vCfg := s.VisibilityTestCluster.Config()
	cfg.VisibilityStore = "visibility_ " + vCfg.VisibilityStore
	cfg.DataStores[cfg.VisibilityStore] = vCfg.DataStores[vCfg.VisibilityStore]
	return cfg
}

// Setup sets up the test base, must be called as part of SetupSuite
func (s *TestBase) Setup() {
	var err error
	shardID := 10
	clusterName := s.ClusterMetadata.GetCurrentClusterName()

	s.DefaultTestCluster.SetupTestDatabase()
	if s.VisibilityTestCluster != s.DefaultTestCluster {
		s.VisibilityTestCluster.SetupTestDatabase()
	}

	cfg := s.DefaultTestCluster.Config()
	scope := tally.NewTestScope(common.HistoryServiceName, make(map[string]string))
	metricsClient := metrics.NewClient(scope, service.GetMetricsServiceIdx(common.HistoryServiceName, s.Logger))
	factory := client.NewFactory(&cfg, nil, clusterName, metricsClient, s.Logger)

	s.TaskMgr, err = factory.NewTaskManager()
	s.fatalOnError("NewTaskManager", err)

	s.MetadataManager, err = factory.NewMetadataManager()
	s.fatalOnError("NewMetadataManager", err)

	s.HistoryV2Mgr, err = factory.NewHistoryManager()
	s.fatalOnError("NewHistoryManager", err)

	s.ShardMgr, err = factory.NewShardManager()
	s.fatalOnError("NewShardManager", err)

	s.ExecutionMgrFactory = factory
	s.ExecutionManager, err = factory.NewExecutionManager(shardID)
	s.fatalOnError("NewExecutionManager", err)

	visibilityFactory := factory
	if s.VisibilityTestCluster != s.DefaultTestCluster {
		vCfg := s.VisibilityTestCluster.Config()
		visibilityFactory = client.NewFactory(&vCfg, nil, clusterName, nil, s.Logger)
	}
	// SQL currently doesn't have support for visibility manager
	s.VisibilityMgr, err = visibilityFactory.NewVisibilityManager()
	if err != nil {
		s.fatalOnError("NewVisibilityManager", err)
	}

	s.ReadLevel = 0
	s.ReplicationReadLevel = 0

	domainFilter := &history.DomainFilter{
		DomainIDs:    []string{},
		ReverseMatch: common.BoolPtr(true),
	}
	transferPQSMap := map[string][]*history.ProcessingQueueState{
		s.ClusterMetadata.GetCurrentClusterName(): {
			&history.ProcessingQueueState{
				Level:        common.Int32Ptr(0),
				AckLevel:     common.Int64Ptr(0),
				MaxLevel:     common.Int64Ptr(0),
				DomainFilter: domainFilter,
			},
		},
	}
	transferPQS := history.ProcessingQueueStates{transferPQSMap}
	transferPQSBlob, _ := s.PayloadSerializer.SerializeProcessingQueueStates(
		&transferPQS,
		common.EncodingTypeThriftRW,
	)
	timerPQSMap := map[string][]*history.ProcessingQueueState{
		s.ClusterMetadata.GetCurrentClusterName(): {
			&history.ProcessingQueueState{
				Level:        common.Int32Ptr(0),
				AckLevel:     common.Int64Ptr(time.Now().UnixNano()),
				MaxLevel:     common.Int64Ptr(time.Now().UnixNano()),
				DomainFilter: domainFilter,
			},
		},
	}
	timerPQS := history.ProcessingQueueStates{StatesByCluster: timerPQSMap}
	timerPQSBlob, _ := s.PayloadSerializer.SerializeProcessingQueueStates(
		&timerPQS,
		common.EncodingTypeThriftRW,
	)

	s.ShardInfo = &p.ShardInfo{
		ShardID:                       shardID,
		RangeID:                       0,
		TransferAckLevel:              0,
		ReplicationAckLevel:           0,
		TimerAckLevel:                 time.Time{},
		ClusterTimerAckLevel:          map[string]time.Time{clusterName: time.Time{}},
		ClusterTransferAckLevel:       map[string]int64{clusterName: 0},
		TransferProcessingQueueStates: transferPQSBlob,
		TimerProcessingQueueStates:    timerPQSBlob,
	}

	s.TaskIDGenerator = &TestTransferTaskIDGenerator{}
	err = s.ShardMgr.CreateShard(context.Background(), &p.CreateShardRequest{ShardInfo: s.ShardInfo})
	s.fatalOnError("CreateShard", err)

	queue, err := factory.NewDomainReplicationQueue()
	s.fatalOnError("Create DomainReplicationQueue", err)
	s.DomainReplicationQueue = queue
}

func (s *TestBase) fatalOnError(msg string, err error) {
	if err != nil {
		s.Logger.Fatal(msg, tag.Error(err))
	}
}

// CreateShard is a utility method to create the shard using persistence layer
func (s *TestBase) CreateShard(ctx context.Context, shardID int, owner string, rangeID int64) error {
	info := &p.ShardInfo{
		ShardID: shardID,
		Owner:   owner,
		RangeID: rangeID,
	}

	return s.ShardMgr.CreateShard(ctx, &p.CreateShardRequest{
		ShardInfo: info,
	})
}

// GetShard is a utility method to get the shard using persistence layer
func (s *TestBase) GetShard(ctx context.Context, shardID int) (*p.ShardInfo, error) {
	response, err := s.ShardMgr.GetShard(ctx, &p.GetShardRequest{
		ShardID: shardID,
	})

	if err != nil {
		return nil, err
	}

	return response.ShardInfo, nil
}

// UpdateShard is a utility method to update the shard using persistence layer
func (s *TestBase) UpdateShard(ctx context.Context, updatedInfo *p.ShardInfo, previousRangeID int64) error {
	return s.ShardMgr.UpdateShard(ctx, &p.UpdateShardRequest{
		ShardInfo:       updatedInfo,
		PreviousRangeID: previousRangeID,
	})
}

// CreateWorkflowExecutionWithBranchToken test util function
func (s *TestBase) CreateWorkflowExecutionWithBranchToken(
	ctx context.Context,
	domainID string,
	workflowExecution workflow.WorkflowExecution,
	taskList string,
	wType string,
	wTimeout int32,
	decisionTimeout int32,
	executionContext []byte,
	nextEventID int64,
	lastProcessedEventID int64,
	decisionScheduleID int64,
	branchToken []byte,
	timerTasks []p.Task,
) (*p.CreateWorkflowExecutionResponse, error) {

	versionHistory := p.NewVersionHistory(branchToken, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	response, err := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				TaskList:                    taskList,
				WorkflowTypeName:            wType,
				WorkflowTimeout:             wTimeout,
				DecisionStartToCloseTimeout: decisionTimeout,
				ExecutionContext:            executionContext,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				LastFirstEventID:            common.FirstEventID,
				NextEventID:                 nextEventID,
				LastProcessedEvent:          lastProcessedEventID,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				BranchToken:                 branchToken,
			},
			ExecutionStats: &p.ExecutionStats{},
			TransferTasks: []p.Task{
				&p.DecisionTask{
					TaskID:              s.GetNextSequenceNumber(),
					DomainID:            domainID,
					TaskList:            taskList,
					ScheduleID:          decisionScheduleID,
					VisibilityTimestamp: time.Now(),
				},
			},
			TimerTasks:       timerTasks,
			Checksum:         testWorkflowChecksum,
			VersionHistories: verisonHistories,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	return response, err
}

// CreateWorkflowExecution is a utility method to create workflow executions
func (s *TestBase) CreateWorkflowExecution(
	ctx context.Context,
	domainID string,
	workflowExecution workflow.WorkflowExecution,
	taskList string,
	wType string,
	wTimeout int32,
	decisionTimeout int32,
	executionContext []byte,
	nextEventID int64,
	lastProcessedEventID int64,
	decisionScheduleID int64,
	timerTasks []p.Task,
) (*p.CreateWorkflowExecutionResponse, error) {

	return s.CreateWorkflowExecutionWithBranchToken(ctx, domainID, workflowExecution, taskList, wType, wTimeout, decisionTimeout,
		executionContext, nextEventID, lastProcessedEventID, decisionScheduleID, nil, timerTasks)
}

// CreateWorkflowExecutionWithReplication is a utility method to create workflow executions
func (s *TestBase) CreateWorkflowExecutionWithReplication(
	ctx context.Context,
	domainID string,
	workflowExecution workflow.WorkflowExecution,
	taskList string,
	wType string,
	wTimeout int32,
	decisionTimeout int32,
	nextEventID int64,
	lastProcessedEventID int64,
	decisionScheduleID int64,
	txTasks []p.Task,
) (*p.CreateWorkflowExecutionResponse, error) {

	var transferTasks []p.Task
	var replicationTasks []p.Task
	for _, task := range txTasks {
		switch t := task.(type) {
		case *p.DecisionTask, *p.ActivityTask, *p.CloseExecutionTask, *p.CancelExecutionTask, *p.StartChildExecutionTask, *p.SignalExecutionTask, *p.RecordWorkflowStartedTask:
			transferTasks = append(transferTasks, t)
		case *p.HistoryReplicationTask:
			replicationTasks = append(replicationTasks, t)
		default:
			panic("Unknown transfer task type.")
		}
	}

	transferTasks = append(transferTasks, &p.DecisionTask{
		TaskID:     s.GetNextSequenceNumber(),
		DomainID:   domainID,
		TaskList:   taskList,
		ScheduleID: decisionScheduleID,
	})
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	response, err := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				TaskList:                    taskList,
				WorkflowTypeName:            wType,
				WorkflowTimeout:             wTimeout,
				DecisionStartToCloseTimeout: decisionTimeout,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				LastFirstEventID:            common.FirstEventID,
				NextEventID:                 nextEventID,
				LastProcessedEvent:          lastProcessedEventID,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
			},
			ExecutionStats:   &p.ExecutionStats{},
			TransferTasks:    transferTasks,
			ReplicationTasks: replicationTasks,
			Checksum:         testWorkflowChecksum,
			VersionHistories: verisonHistories,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	return response, err
}

// CreateWorkflowExecutionManyTasks is a utility method to create workflow executions
func (s *TestBase) CreateWorkflowExecutionManyTasks(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution,
	taskList string, executionContext []byte, nextEventID int64, lastProcessedEventID int64,
	decisionScheduleIDs []int64, activityScheduleIDs []int64) (*p.CreateWorkflowExecutionResponse, error) {

	transferTasks := []p.Task{}
	for _, decisionScheduleID := range decisionScheduleIDs {
		transferTasks = append(transferTasks,
			&p.DecisionTask{
				TaskID:     s.GetNextSequenceNumber(),
				DomainID:   domainID,
				TaskList:   taskList,
				ScheduleID: int64(decisionScheduleID),
			})
	}

	for _, activityScheduleID := range activityScheduleIDs {
		transferTasks = append(transferTasks,
			&p.ActivityTask{
				TaskID:     s.GetNextSequenceNumber(),
				DomainID:   domainID,
				TaskList:   taskList,
				ScheduleID: int64(activityScheduleID),
			})
	}

	response, err := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:    uuid.New(),
				DomainID:           domainID,
				WorkflowID:         workflowExecution.GetWorkflowId(),
				RunID:              workflowExecution.GetRunId(),
				TaskList:           taskList,
				ExecutionContext:   executionContext,
				State:              p.WorkflowStateRunning,
				CloseStatus:        p.WorkflowCloseStatusNone,
				LastFirstEventID:   common.FirstEventID,
				NextEventID:        nextEventID,
				LastProcessedEvent: lastProcessedEventID,
				DecisionScheduleID: common.EmptyEventID,
				DecisionStartedID:  common.EmptyEventID,
				DecisionTimeout:    1,
			},
			ExecutionStats: &p.ExecutionStats{},
			TransferTasks:  transferTasks,
			Checksum:       testWorkflowChecksum,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	return response, err
}

// CreateChildWorkflowExecution is a utility method to create child workflow executions
func (s *TestBase) CreateChildWorkflowExecution(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution,
	parentDomainID string, parentExecution workflow.WorkflowExecution, initiatedID int64, taskList, wType string,
	wTimeout int32, decisionTimeout int32, executionContext []byte, nextEventID int64, lastProcessedEventID int64,
	decisionScheduleID int64, timerTasks []p.Task) (*p.CreateWorkflowExecutionResponse, error) {
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	response, err := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				ParentDomainID:              parentDomainID,
				ParentWorkflowID:            parentExecution.GetWorkflowId(),
				ParentRunID:                 parentExecution.GetRunId(),
				InitiatedID:                 initiatedID,
				TaskList:                    taskList,
				WorkflowTypeName:            wType,
				WorkflowTimeout:             wTimeout,
				DecisionStartToCloseTimeout: decisionTimeout,
				ExecutionContext:            executionContext,
				State:                       p.WorkflowStateCreated,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				LastFirstEventID:            common.FirstEventID,
				NextEventID:                 nextEventID,
				LastProcessedEvent:          lastProcessedEventID,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
			},
			ExecutionStats: &p.ExecutionStats{},
			TransferTasks: []p.Task{
				&p.DecisionTask{
					TaskID:     s.GetNextSequenceNumber(),
					DomainID:   domainID,
					TaskList:   taskList,
					ScheduleID: decisionScheduleID,
				},
			},
			TimerTasks:       timerTasks,
			VersionHistories: verisonHistories,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	return response, err
}

// GetWorkflowExecutionInfoWithStats is a utility method to retrieve execution info with size stats
func (s *TestBase) GetWorkflowExecutionInfoWithStats(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution) (
	*p.MutableStateStats, *p.WorkflowMutableState, error) {
	response, err := s.ExecutionManager.GetWorkflowExecution(ctx, &p.GetWorkflowExecutionRequest{
		DomainID:  domainID,
		Execution: workflowExecution,
	})
	if err != nil {
		return nil, nil, err
	}

	return response.MutableStateStats, response.State, nil
}

// GetWorkflowExecutionInfo is a utility method to retrieve execution info
func (s *TestBase) GetWorkflowExecutionInfo(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution) (
	*p.WorkflowMutableState, error) {
	response, err := s.ExecutionManager.GetWorkflowExecution(ctx, &p.GetWorkflowExecutionRequest{
		DomainID:  domainID,
		Execution: workflowExecution,
	})
	if err != nil {
		return nil, err
	}
	return response.State, nil
}

// GetCurrentWorkflowRunID returns the workflow run ID for the given params
func (s *TestBase) GetCurrentWorkflowRunID(ctx context.Context, domainID, workflowID string) (string, error) {
	response, err := s.ExecutionManager.GetCurrentExecution(ctx, &p.GetCurrentExecutionRequest{
		DomainID:   domainID,
		WorkflowID: workflowID,
	})

	if err != nil {
		return "", err
	}

	return response.RunID, nil
}

// ContinueAsNewExecution is a utility method to create workflow executions
func (s *TestBase) ContinueAsNewExecution(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	condition int64,
	newExecution workflow.WorkflowExecution,
	nextEventID, decisionScheduleID int64,
	prevResetPoints *workflow.ResetPoints,
) error {

	newdecisionTask := &p.DecisionTask{
		TaskID:     s.GetNextSequenceNumber(),
		DomainID:   updatedInfo.DomainID,
		TaskList:   updatedInfo.TaskList,
		ScheduleID: int64(decisionScheduleID),
	}
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)

	req := &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:       updatedInfo,
			ExecutionStats:      updatedStats,
			TransferTasks:       []p.Task{newdecisionTask},
			TimerTasks:          nil,
			Condition:           condition,
			UpsertActivityInfos: nil,
			DeleteActivityInfos: nil,
			UpsertTimerInfos:    nil,
			DeleteTimerInfos:    nil,
			VersionHistories:    verisonHistories,
		},
		NewWorkflowSnapshot: &p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    updatedInfo.DomainID,
				WorkflowID:                  newExecution.GetWorkflowId(),
				RunID:                       newExecution.GetRunId(),
				TaskList:                    updatedInfo.TaskList,
				WorkflowTypeName:            updatedInfo.WorkflowTypeName,
				WorkflowTimeout:             updatedInfo.WorkflowTimeout,
				DecisionStartToCloseTimeout: updatedInfo.DecisionStartToCloseTimeout,
				ExecutionContext:            nil,
				State:                       updatedInfo.State,
				CloseStatus:                 updatedInfo.CloseStatus,
				LastFirstEventID:            common.FirstEventID,
				NextEventID:                 nextEventID,
				LastProcessedEvent:          common.EmptyEventID,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				AutoResetPoints:             prevResetPoints,
			},
			ExecutionStats:   updatedStats,
			TransferTasks:    nil,
			TimerTasks:       nil,
			VersionHistories: verisonHistories,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	}
	req.UpdateWorkflowMutation.ExecutionInfo.State = p.WorkflowStateCompleted
	req.UpdateWorkflowMutation.ExecutionInfo.CloseStatus = p.WorkflowCloseStatusContinuedAsNew
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, req)
	return err
}

// UpdateWorkflowExecution is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecution(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	decisionScheduleIDs []int64, activityScheduleIDs []int64, condition int64, timerTasks []p.Task,
	upsertActivityInfos []*p.ActivityInfo, deleteActivityInfos []int64,
	upsertTimerInfos []*p.TimerInfo, deleteTimerInfos []string) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, decisionScheduleIDs, activityScheduleIDs,
		s.ShardInfo.RangeID, condition, timerTasks, upsertActivityInfos, deleteActivityInfos,
		upsertTimerInfos, deleteTimerInfos, nil, nil, nil, nil,
		nil, nil, nil, "")
}

// UpdateWorkflowExecutionAndFinish is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionAndFinish(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	condition int64,
	versionHistories *p.VersionHistories,
) error {
	transferTasks := []p.Task{}
	transferTasks = append(transferTasks, &p.CloseExecutionTask{TaskID: s.GetNextSequenceNumber()})
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		RangeID: s.ShardInfo.RangeID,
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:       updatedInfo,
			ExecutionStats:      updatedStats,
			TransferTasks:       transferTasks,
			TimerTasks:          nil,
			Condition:           condition,
			UpsertActivityInfos: nil,
			DeleteActivityInfos: nil,
			UpsertTimerInfos:    nil,
			DeleteTimerInfos:    nil,
			VersionHistories:    versionHistories,
		},
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpsertChildExecutionsState is a utility method to update mutable state of workflow execution
func (s *TestBase) UpsertChildExecutionsState(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	updatedVersionHistories *p.VersionHistories,
	condition int64,
	upsertChildInfos []*p.ChildExecutionInfo,
) error {

	return s.UpdateWorkflowExecutionWithRangeID(
		ctx,
		updatedInfo,
		updatedStats,
		updatedVersionHistories,
		nil,
		nil,
		s.ShardInfo.RangeID,
		condition,
		nil,
		nil,
		nil,
		nil,
		nil,
		upsertChildInfos,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
	)
}

// UpsertRequestCancelState is a utility method to update mutable state of workflow execution
func (s *TestBase) UpsertRequestCancelState(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	updatedVersionHistories *p.VersionHistories,
	condition int64,
	upsertCancelInfos []*p.RequestCancelInfo,
) error {

	return s.UpdateWorkflowExecutionWithRangeID(
		ctx,
		updatedInfo,
		updatedStats,
		updatedVersionHistories,
		nil,
		nil,
		s.ShardInfo.RangeID,
		condition,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		upsertCancelInfos,
		nil,
		nil,
		nil,
		nil,
		"",
	)
}

// UpsertSignalInfoState is a utility method to update mutable state of workflow execution
func (s *TestBase) UpsertSignalInfoState(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	updatedVersionHistories *p.VersionHistories,
	condition int64,
	upsertSignalInfos []*p.SignalInfo,
) error {

	return s.UpdateWorkflowExecutionWithRangeID(
		ctx,
		updatedInfo,
		updatedStats,
		updatedVersionHistories,
		nil,
		nil,
		s.ShardInfo.RangeID,
		condition,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		upsertSignalInfos,
		nil,
		nil,
		"",
	)
}

// UpsertSignalsRequestedState is a utility method to update mutable state of workflow execution
func (s *TestBase) UpsertSignalsRequestedState(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	condition int64, upsertSignalsRequested []string) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, nil, nil,
		s.ShardInfo.RangeID, condition, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		nil, nil, upsertSignalsRequested, "")
}

// DeleteChildExecutionsState is a utility method to delete child execution from mutable state
func (s *TestBase) DeleteChildExecutionsState(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	condition int64, deleteChildInfo int64) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, nil, nil,
		s.ShardInfo.RangeID, condition, nil, nil, nil,
		nil, nil, nil, &deleteChildInfo, nil, nil,
		nil, nil, nil, "")
}

// DeleteCancelState is a utility method to delete request cancel state from mutable state
func (s *TestBase) DeleteCancelState(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	condition int64, deleteCancelInfo int64) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, nil, nil,
		s.ShardInfo.RangeID, condition, nil, nil, nil,
		nil, nil, nil, nil, nil, &deleteCancelInfo,
		nil, nil, nil, "")
}

// DeleteSignalState is a utility method to delete request cancel state from mutable state
func (s *TestBase) DeleteSignalState(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	condition int64, deleteSignalInfo int64) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, nil, nil,
		s.ShardInfo.RangeID, condition, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		nil, &deleteSignalInfo, nil, "")
}

// DeleteSignalsRequestedState is a utility method to delete mutable state of workflow execution
func (s *TestBase) DeleteSignalsRequestedState(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	condition int64, deleteSignalsRequestedID string) error {
	return s.UpdateWorkflowExecutionWithRangeID(ctx, updatedInfo, updatedStats, updatedVersionHistories, nil, nil,
		s.ShardInfo.RangeID, condition, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		nil, nil, nil, deleteSignalsRequestedID)
}

// UpdateWorklowStateAndReplication is a utility method to update workflow execution
func (s *TestBase) UpdateWorklowStateAndReplication(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	updatedVersionHistories *p.VersionHistories,
	condition int64,
	txTasks []p.Task,
) error {

	return s.UpdateWorkflowExecutionWithReplication(
		ctx,
		updatedInfo,
		updatedStats,
		updatedVersionHistories,
		nil,
		nil,
		s.ShardInfo.RangeID,
		condition,
		nil,
		txTasks,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
	)
}

// UpdateWorkflowExecutionWithRangeID is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionWithRangeID(ctx context.Context, updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, updatedVersionHistories *p.VersionHistories,
	decisionScheduleIDs []int64, activityScheduleIDs []int64, rangeID, condition int64, timerTasks []p.Task,
	upsertActivityInfos []*p.ActivityInfo, deleteActivityInfos []int64, upsertTimerInfos []*p.TimerInfo,
	deleteTimerInfos []string, upsertChildInfos []*p.ChildExecutionInfo, deleteChildInfo *int64,
	upsertCancelInfos []*p.RequestCancelInfo, deleteCancelInfo *int64,
	upsertSignalInfos []*p.SignalInfo, deleteSignalInfo *int64,
	upsertSignalRequestedIDs []string, deleteSignalRequestedID string) error {
	return s.UpdateWorkflowExecutionWithReplication(
		ctx,
		updatedInfo,
		updatedStats,
		updatedVersionHistories,
		decisionScheduleIDs,
		activityScheduleIDs,
		rangeID,
		condition,
		timerTasks,
		[]p.Task{},
		upsertActivityInfos,
		deleteActivityInfos,
		upsertTimerInfos,
		deleteTimerInfos,
		upsertChildInfos,
		deleteChildInfo,
		upsertCancelInfos,
		deleteCancelInfo,
		upsertSignalInfos,
		deleteSignalInfo,
		upsertSignalRequestedIDs,
		deleteSignalRequestedID,
	)
}

// UpdateWorkflowExecutionWithReplication is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionWithReplication(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	updatedVersionHistories *p.VersionHistories,
	decisionScheduleIDs []int64,
	activityScheduleIDs []int64,
	rangeID int64,
	condition int64,
	timerTasks []p.Task,
	txTasks []p.Task,
	upsertActivityInfos []*p.ActivityInfo,
	deleteActivityInfos []int64,
	upsertTimerInfos []*p.TimerInfo,
	deleteTimerInfos []string,
	upsertChildInfos []*p.ChildExecutionInfo,
	deleteChildInfo *int64,
	upsertCancelInfos []*p.RequestCancelInfo,
	deleteCancelInfo *int64,
	upsertSignalInfos []*p.SignalInfo,
	deleteSignalInfo *int64,
	upsertSignalRequestedIDs []string,
	deleteSignalRequestedID string,
) error {

	var transferTasks []p.Task
	var replicationTasks []p.Task
	for _, task := range txTasks {
		switch t := task.(type) {
		case *p.DecisionTask, *p.ActivityTask, *p.CloseExecutionTask, *p.CancelExecutionTask, *p.StartChildExecutionTask, *p.SignalExecutionTask, *p.RecordWorkflowStartedTask:
			transferTasks = append(transferTasks, t)
		case *p.HistoryReplicationTask, *p.SyncActivityTask:
			replicationTasks = append(replicationTasks, t)
		default:
			panic("Unknown transfer task type.")
		}
	}
	for _, decisionScheduleID := range decisionScheduleIDs {
		transferTasks = append(transferTasks, &p.DecisionTask{
			TaskID:     s.GetNextSequenceNumber(),
			DomainID:   updatedInfo.DomainID,
			TaskList:   updatedInfo.TaskList,
			ScheduleID: int64(decisionScheduleID)})
	}

	for _, activityScheduleID := range activityScheduleIDs {
		transferTasks = append(transferTasks, &p.ActivityTask{
			TaskID:     s.GetNextSequenceNumber(),
			DomainID:   updatedInfo.DomainID,
			TaskList:   updatedInfo.TaskList,
			ScheduleID: int64(activityScheduleID)})
	}
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		RangeID: rangeID,
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:    updatedInfo,
			ExecutionStats:   updatedStats,
			VersionHistories: updatedVersionHistories,

			UpsertActivityInfos:       upsertActivityInfos,
			DeleteActivityInfos:       deleteActivityInfos,
			UpsertTimerInfos:          upsertTimerInfos,
			DeleteTimerInfos:          deleteTimerInfos,
			UpsertChildExecutionInfos: upsertChildInfos,
			DeleteChildExecutionInfo:  deleteChildInfo,
			UpsertRequestCancelInfos:  upsertCancelInfos,
			DeleteRequestCancelInfo:   deleteCancelInfo,
			UpsertSignalInfos:         upsertSignalInfos,
			DeleteSignalInfo:          deleteSignalInfo,
			UpsertSignalRequestedIDs:  upsertSignalRequestedIDs,
			DeleteSignalRequestedID:   deleteSignalRequestedID,

			TransferTasks:    transferTasks,
			ReplicationTasks: replicationTasks,
			TimerTasks:       timerTasks,

			Condition: condition,
			Checksum:  testWorkflowChecksum,
		},
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateWorkflowExecutionWithTransferTasks is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionWithTransferTasks(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	condition int64,
	transferTasks []p.Task,
	upsertActivityInfo []*p.ActivityInfo,
	versionHistories *p.VersionHistories,
) error {

	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:       updatedInfo,
			ExecutionStats:      updatedStats,
			TransferTasks:       transferTasks,
			Condition:           condition,
			UpsertActivityInfos: upsertActivityInfo,
			VersionHistories:    versionHistories,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateWorkflowExecutionForChildExecutionsInitiated is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionForChildExecutionsInitiated(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, condition int64, transferTasks []p.Task, childInfos []*p.ChildExecutionInfo) error {
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:             updatedInfo,
			ExecutionStats:            updatedStats,
			TransferTasks:             transferTasks,
			Condition:                 condition,
			UpsertChildExecutionInfos: childInfos,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateWorkflowExecutionForRequestCancel is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionForRequestCancel(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, condition int64, transferTasks []p.Task,
	upsertRequestCancelInfo []*p.RequestCancelInfo) error {
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:            updatedInfo,
			ExecutionStats:           updatedStats,
			TransferTasks:            transferTasks,
			Condition:                condition,
			UpsertRequestCancelInfos: upsertRequestCancelInfo,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateWorkflowExecutionForSignal is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionForSignal(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo, updatedStats *p.ExecutionStats, condition int64, transferTasks []p.Task,
	upsertSignalInfos []*p.SignalInfo) error {
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:     updatedInfo,
			ExecutionStats:    updatedStats,
			TransferTasks:     transferTasks,
			Condition:         condition,
			UpsertSignalInfos: upsertSignalInfos,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateWorkflowExecutionForBufferEvents is a utility method to update workflow execution
func (s *TestBase) UpdateWorkflowExecutionForBufferEvents(
	ctx context.Context,
	updatedInfo *p.WorkflowExecutionInfo,
	updatedStats *p.ExecutionStats,
	condition int64,
	bufferEvents []*workflow.HistoryEvent,
	clearBufferedEvents bool,
	versionHistories *p.VersionHistories,
) error {

	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:       updatedInfo,
			ExecutionStats:      updatedStats,
			NewBufferedEvents:   bufferEvents,
			Condition:           condition,
			ClearBufferedEvents: clearBufferedEvents,
			VersionHistories:    versionHistories,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})
	return err
}

// UpdateAllMutableState is a utility method to update workflow execution
func (s *TestBase) UpdateAllMutableState(ctx context.Context, updatedMutableState *p.WorkflowMutableState, condition int64) error {
	var aInfos []*p.ActivityInfo
	for _, ai := range updatedMutableState.ActivityInfos {
		aInfos = append(aInfos, ai)
	}

	var tInfos []*p.TimerInfo
	for _, ti := range updatedMutableState.TimerInfos {
		tInfos = append(tInfos, ti)
	}

	var cInfos []*p.ChildExecutionInfo
	for _, ci := range updatedMutableState.ChildExecutionInfos {
		cInfos = append(cInfos, ci)
	}

	var rcInfos []*p.RequestCancelInfo
	for _, rci := range updatedMutableState.RequestCancelInfos {
		rcInfos = append(rcInfos, rci)
	}

	var sInfos []*p.SignalInfo
	for _, si := range updatedMutableState.SignalInfos {
		sInfos = append(sInfos, si)
	}

	var srIDs []string
	for id := range updatedMutableState.SignalRequestedIDs {
		srIDs = append(srIDs, id)
	}
	_, err := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		RangeID: s.ShardInfo.RangeID,
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:             updatedMutableState.ExecutionInfo,
			ExecutionStats:            updatedMutableState.ExecutionStats,
			Condition:                 condition,
			UpsertActivityInfos:       aInfos,
			UpsertTimerInfos:          tInfos,
			UpsertChildExecutionInfos: cInfos,
			UpsertRequestCancelInfos:  rcInfos,
			UpsertSignalInfos:         sInfos,
			UpsertSignalRequestedIDs:  srIDs,
			VersionHistories:          updatedMutableState.VersionHistories,
		},
		Encoding: pickRandomEncoding(),
	})
	return err
}

// ConflictResolveWorkflowExecution is  utility method to reset mutable state
func (s *TestBase) ConflictResolveWorkflowExecution(
	ctx context.Context,
	info *p.WorkflowExecutionInfo,
	stats *p.ExecutionStats,
	nextEventID int64,
	activityInfos []*p.ActivityInfo,
	timerInfos []*p.TimerInfo,
	childExecutionInfos []*p.ChildExecutionInfo,
	requestCancelInfos []*p.RequestCancelInfo,
	signalInfos []*p.SignalInfo,
	ids []string,
	versionHistories *p.VersionHistories,
) error {

	return s.ExecutionManager.ConflictResolveWorkflowExecution(ctx, &p.ConflictResolveWorkflowExecutionRequest{
		RangeID: s.ShardInfo.RangeID,
		ResetWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo:       info,
			ExecutionStats:      stats,
			Condition:           nextEventID,
			ActivityInfos:       activityInfos,
			TimerInfos:          timerInfos,
			ChildExecutionInfos: childExecutionInfos,
			RequestCancelInfos:  requestCancelInfos,
			SignalInfos:         signalInfos,
			SignalRequestedIDs:  ids,
			Checksum:            testWorkflowChecksum,
			VersionHistories:    versionHistories,
		},
		Encoding: pickRandomEncoding(),
	})
}

// ResetWorkflowExecution is  utility method to reset WF
func (s *TestBase) ResetWorkflowExecution(
	ctx context.Context,
	condition int64,
	info *p.WorkflowExecutionInfo,
	executionStats *p.ExecutionStats,
	activityInfos []*p.ActivityInfo,
	timerInfos []*p.TimerInfo,
	childExecutionInfos []*p.ChildExecutionInfo,
	requestCancelInfos []*p.RequestCancelInfo,
	signalInfos []*p.SignalInfo,
	ids []string,
	trasTasks []p.Task,
	timerTasks []p.Task,
	replTasks []p.Task,
	updateCurr bool,
	currInfo *p.WorkflowExecutionInfo,
	currExecutionStats *p.ExecutionStats,
	currTrasTasks,
	currTimerTasks []p.Task,
	forkRunID string,
	forkRunNextEventID int64,
) error {

	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{info.LastProcessedEvent, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)

	req := &p.ResetWorkflowExecutionRequest{
		RangeID: s.ShardInfo.RangeID,

		BaseRunID:          forkRunID,
		BaseRunNextEventID: forkRunNextEventID,

		CurrentRunID:          currInfo.RunID,
		CurrentRunNextEventID: condition,

		CurrentWorkflowMutation: nil,

		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo:  info,
			ExecutionStats: executionStats,

			ActivityInfos:       activityInfos,
			TimerInfos:          timerInfos,
			ChildExecutionInfos: childExecutionInfos,
			RequestCancelInfos:  requestCancelInfos,
			SignalInfos:         signalInfos,
			SignalRequestedIDs:  ids,

			TransferTasks:    trasTasks,
			ReplicationTasks: replTasks,
			TimerTasks:       timerTasks,
			VersionHistories: verisonHistories,
		},
		Encoding: pickRandomEncoding(),
	}

	if updateCurr {
		req.CurrentWorkflowMutation = &p.WorkflowMutation{
			ExecutionInfo:  currInfo,
			ExecutionStats: currExecutionStats,

			TransferTasks: currTrasTasks,
			TimerTasks:    currTimerTasks,

			Condition:        condition,
			VersionHistories: verisonHistories,
		}
	}

	return s.ExecutionManager.ResetWorkflowExecution(ctx, req)
}

// DeleteWorkflowExecution is a utility method to delete a workflow execution
func (s *TestBase) DeleteWorkflowExecution(ctx context.Context, info *p.WorkflowExecutionInfo) error {
	return s.ExecutionManager.DeleteWorkflowExecution(ctx, &p.DeleteWorkflowExecutionRequest{
		DomainID:   info.DomainID,
		WorkflowID: info.WorkflowID,
		RunID:      info.RunID,
	})
}

// DeleteCurrentWorkflowExecution is a utility method to delete the workflow current execution
func (s *TestBase) DeleteCurrentWorkflowExecution(ctx context.Context, info *p.WorkflowExecutionInfo) error {
	return s.ExecutionManager.DeleteCurrentWorkflowExecution(ctx, &p.DeleteCurrentWorkflowExecutionRequest{
		DomainID:   info.DomainID,
		WorkflowID: info.WorkflowID,
		RunID:      info.RunID,
	})
}

// GetTransferTasks is a utility method to get tasks from transfer task queue
func (s *TestBase) GetTransferTasks(ctx context.Context, batchSize int, getAll bool) ([]*p.TransferTaskInfo, error) {
	result := []*p.TransferTaskInfo{}
	var token []byte

Loop:
	for {
		response, err := s.ExecutionManager.GetTransferTasks(ctx, &p.GetTransferTasksRequest{
			ReadLevel:     s.GetTransferReadLevel(),
			MaxReadLevel:  int64(math.MaxInt64),
			BatchSize:     batchSize,
			NextPageToken: token,
		})
		if err != nil {
			return nil, err
		}

		token = response.NextPageToken
		result = append(result, response.Tasks...)
		if len(token) == 0 || !getAll {
			break Loop
		}
	}

	for _, task := range result {
		atomic.StoreInt64(&s.ReadLevel, task.TaskID)
	}

	return result, nil
}

// GetReplicationTasks is a utility method to get tasks from replication task queue
func (s *TestBase) GetReplicationTasks(ctx context.Context, batchSize int, getAll bool) ([]*p.ReplicationTaskInfo, error) {
	result := []*p.ReplicationTaskInfo{}
	var token []byte

Loop:
	for {
		response, err := s.ExecutionManager.GetReplicationTasks(ctx, &p.GetReplicationTasksRequest{
			ReadLevel:     s.GetReplicationReadLevel(),
			MaxReadLevel:  int64(math.MaxInt64),
			BatchSize:     batchSize,
			NextPageToken: token,
		})
		if err != nil {
			return nil, err
		}

		token = response.NextPageToken
		result = append(result, response.Tasks...)
		if len(token) == 0 || !getAll {
			break Loop
		}
	}

	for _, task := range result {
		atomic.StoreInt64(&s.ReplicationReadLevel, task.TaskID)
	}

	return result, nil
}

// RangeCompleteReplicationTask is a utility method to complete a range of replication tasks
func (s *TestBase) RangeCompleteReplicationTask(ctx context.Context, inclusiveEndTaskID int64) error {
	return s.ExecutionManager.RangeCompleteReplicationTask(ctx, &p.RangeCompleteReplicationTaskRequest{
		InclusiveEndTaskID: inclusiveEndTaskID,
	})
}

// PutReplicationTaskToDLQ is a utility method to insert a replication task info
func (s *TestBase) PutReplicationTaskToDLQ(
	ctx context.Context,
	sourceCluster string,
	taskInfo *p.ReplicationTaskInfo,
) error {

	return s.ExecutionManager.PutReplicationTaskToDLQ(ctx, &p.PutReplicationTaskToDLQRequest{
		SourceClusterName: sourceCluster,
		TaskInfo:          taskInfo,
	})
}

// GetReplicationTasksFromDLQ is a utility method to read replication task info
func (s *TestBase) GetReplicationTasksFromDLQ(
	ctx context.Context,
	sourceCluster string,
	readLevel int64,
	maxReadLevel int64,
	pageSize int,
	pageToken []byte,
) (*p.GetReplicationTasksFromDLQResponse, error) {

	return s.ExecutionManager.GetReplicationTasksFromDLQ(ctx, &p.GetReplicationTasksFromDLQRequest{
		SourceClusterName: sourceCluster,
		GetReplicationTasksRequest: p.GetReplicationTasksRequest{
			ReadLevel:     readLevel,
			MaxReadLevel:  maxReadLevel,
			BatchSize:     pageSize,
			NextPageToken: pageToken,
		},
	})
}

// GetReplicationDLQSize is a utility method to read replication dlq size
func (s *TestBase) GetReplicationDLQSize(
	ctx context.Context,
	sourceCluster string,
) (*p.GetReplicationDLQSizeResponse, error) {

	return s.ExecutionManager.GetReplicationDLQSize(ctx, &p.GetReplicationDLQSizeRequest{
		SourceClusterName: sourceCluster,
	})
}

// DeleteReplicationTaskFromDLQ is a utility method to delete a replication task info
func (s *TestBase) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	sourceCluster string,
	taskID int64,
) error {

	return s.ExecutionManager.DeleteReplicationTaskFromDLQ(ctx, &p.DeleteReplicationTaskFromDLQRequest{
		SourceClusterName: sourceCluster,
		TaskID:            taskID,
	})
}

// RangeDeleteReplicationTaskFromDLQ is a utility method to delete  replication task info
func (s *TestBase) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	sourceCluster string,
	beginTaskID int64,
	endTaskID int64,
) error {

	return s.ExecutionManager.RangeDeleteReplicationTaskFromDLQ(ctx, &p.RangeDeleteReplicationTaskFromDLQRequest{
		SourceClusterName:    sourceCluster,
		ExclusiveBeginTaskID: beginTaskID,
		InclusiveEndTaskID:   endTaskID,
	})
}

// CreateFailoverMarkers is a utility method to create failover markers
func (s *TestBase) CreateFailoverMarkers(
	ctx context.Context,
	markers []*p.FailoverMarkerTask,
) error {

	return s.ExecutionManager.CreateFailoverMarkerTasks(ctx, &p.CreateFailoverMarkersRequest{
		RangeID: s.ShardInfo.RangeID,
		Markers: markers,
	})
}

// CompleteTransferTask is a utility method to complete a transfer task
func (s *TestBase) CompleteTransferTask(ctx context.Context, taskID int64) error {

	return s.ExecutionManager.CompleteTransferTask(ctx, &p.CompleteTransferTaskRequest{
		TaskID: taskID,
	})
}

// RangeCompleteTransferTask is a utility method to complete a range of transfer tasks
func (s *TestBase) RangeCompleteTransferTask(ctx context.Context, exclusiveBeginTaskID int64, inclusiveEndTaskID int64) error {
	return s.ExecutionManager.RangeCompleteTransferTask(ctx, &p.RangeCompleteTransferTaskRequest{
		ExclusiveBeginTaskID: exclusiveBeginTaskID,
		InclusiveEndTaskID:   inclusiveEndTaskID,
	})
}

// CompleteReplicationTask is a utility method to complete a replication task
func (s *TestBase) CompleteReplicationTask(ctx context.Context, taskID int64) error {

	return s.ExecutionManager.CompleteReplicationTask(ctx, &p.CompleteReplicationTaskRequest{
		TaskID: taskID,
	})
}

// GetTimerIndexTasks is a utility method to get tasks from transfer task queue
func (s *TestBase) GetTimerIndexTasks(ctx context.Context, batchSize int, getAll bool) ([]*p.TimerTaskInfo, error) {
	result := []*p.TimerTaskInfo{}
	var token []byte

Loop:
	for {
		response, err := s.ExecutionManager.GetTimerIndexTasks(ctx, &p.GetTimerIndexTasksRequest{
			MinTimestamp:  time.Time{},
			MaxTimestamp:  time.Unix(0, math.MaxInt64),
			BatchSize:     batchSize,
			NextPageToken: token,
		})
		if err != nil {
			return nil, err
		}

		token = response.NextPageToken
		result = append(result, response.Timers...)
		if len(token) == 0 || !getAll {
			break Loop
		}
	}

	return result, nil
}

// CompleteTimerTask is a utility method to complete a timer task
func (s *TestBase) CompleteTimerTask(ctx context.Context, ts time.Time, taskID int64) error {
	return s.ExecutionManager.CompleteTimerTask(ctx, &p.CompleteTimerTaskRequest{
		VisibilityTimestamp: ts,
		TaskID:              taskID,
	})
}

// RangeCompleteTimerTask is a utility method to complete a range of timer tasks
func (s *TestBase) RangeCompleteTimerTask(ctx context.Context, inclusiveBeginTimestamp time.Time, exclusiveEndTimestamp time.Time) error {
	return s.ExecutionManager.RangeCompleteTimerTask(ctx, &p.RangeCompleteTimerTaskRequest{
		InclusiveBeginTimestamp: inclusiveBeginTimestamp,
		ExclusiveEndTimestamp:   exclusiveEndTimestamp,
	})
}

// CreateDecisionTask is a utility method to create a task
func (s *TestBase) CreateDecisionTask(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution, taskList string,
	decisionScheduleID int64) (int64, error) {
	leaseResponse, err := s.TaskMgr.LeaseTaskList(ctx, &p.LeaseTaskListRequest{
		DomainID: domainID,
		TaskList: taskList,
		TaskType: p.TaskListTypeDecision,
	})
	if err != nil {
		return 0, err
	}

	taskID := s.GetNextSequenceNumber()
	tasks := []*p.CreateTaskInfo{
		{
			TaskID:    taskID,
			Execution: workflowExecution,
			Data: &p.TaskInfo{
				DomainID:   domainID,
				WorkflowID: *workflowExecution.WorkflowId,
				RunID:      *workflowExecution.RunId,
				TaskID:     taskID,
				ScheduleID: decisionScheduleID,
			},
		},
	}

	_, err = s.TaskMgr.CreateTasks(ctx, &p.CreateTasksRequest{
		TaskListInfo: leaseResponse.TaskListInfo,
		Tasks:        tasks,
	})

	if err != nil {
		return 0, err
	}

	return taskID, err
}

// CreateActivityTasks is a utility method to create tasks
func (s *TestBase) CreateActivityTasks(ctx context.Context, domainID string, workflowExecution workflow.WorkflowExecution,
	activities map[int64]string) ([]int64, error) {

	taskLists := make(map[string]*p.TaskListInfo)
	for _, tl := range activities {
		_, ok := taskLists[tl]
		if !ok {
			resp, err := s.TaskMgr.LeaseTaskList(
				ctx,
				&p.LeaseTaskListRequest{DomainID: domainID, TaskList: tl, TaskType: p.TaskListTypeActivity})
			if err != nil {
				return []int64{}, err
			}
			taskLists[tl] = resp.TaskListInfo
		}
	}

	var taskIDs []int64
	for activityScheduleID, taskList := range activities {
		taskID := s.GetNextSequenceNumber()
		tasks := []*p.CreateTaskInfo{
			{
				TaskID:    taskID,
				Execution: workflowExecution,
				Data: &p.TaskInfo{
					DomainID:               domainID,
					WorkflowID:             *workflowExecution.WorkflowId,
					RunID:                  *workflowExecution.RunId,
					TaskID:                 taskID,
					ScheduleID:             activityScheduleID,
					ScheduleToStartTimeout: defaultScheduleToStartTimeout,
				},
			},
		}
		_, err := s.TaskMgr.CreateTasks(ctx, &p.CreateTasksRequest{
			TaskListInfo: taskLists[taskList],
			Tasks:        tasks,
		})
		if err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

// GetTasks is a utility method to get tasks from persistence
func (s *TestBase) GetTasks(ctx context.Context, domainID, taskList string, taskType int, batchSize int) (*p.GetTasksResponse, error) {
	response, err := s.TaskMgr.GetTasks(ctx, &p.GetTasksRequest{
		DomainID:     domainID,
		TaskList:     taskList,
		TaskType:     taskType,
		BatchSize:    batchSize,
		MaxReadLevel: common.Int64Ptr(math.MaxInt64),
	})

	if err != nil {
		return nil, err
	}

	return &p.GetTasksResponse{Tasks: response.Tasks}, nil
}

// CompleteTask is a utility method to complete a task
func (s *TestBase) CompleteTask(ctx context.Context, domainID, taskList string, taskType int, taskID int64, ackLevel int64) error {
	return s.TaskMgr.CompleteTask(ctx, &p.CompleteTaskRequest{
		TaskList: &p.TaskListInfo{
			DomainID: domainID,
			AckLevel: ackLevel,
			TaskType: taskType,
			Name:     taskList,
		},
		TaskID: taskID,
	})
}

// TearDownWorkflowStore to cleanup
func (s *TestBase) TearDownWorkflowStore() {
	s.ExecutionMgrFactory.Close()
	// TODO VisibilityMgr/Store is created with a separated code path, this is incorrect and may cause leaking connection
	// And Postgres requires all connection to be closed before dropping a database
	// https://github.com/uber/cadence/issues/2854
	// Remove the below line after the issue is fix
	s.VisibilityMgr.Close()

	s.DefaultTestCluster.TearDownTestDatabase()
}

// GetNextSequenceNumber generates a unique sequence number for can be used for transfer queue taskId
func (s *TestBase) GetNextSequenceNumber() int64 {
	taskID, _ := s.TaskIDGenerator.GenerateTransferTaskID()
	return taskID
}

// GetTransferReadLevel returns the current read level for shard
func (s *TestBase) GetTransferReadLevel() int64 {
	return atomic.LoadInt64(&s.ReadLevel)
}

// GetReplicationReadLevel returns the current read level for shard
func (s *TestBase) GetReplicationReadLevel() int64 {
	return atomic.LoadInt64(&s.ReplicationReadLevel)
}

// ClearTasks completes all transfer tasks and replication tasks
func (s *TestBase) ClearTasks() {
	s.ClearTransferQueue()
	s.ClearReplicationQueue()
}

// ClearTransferQueue completes all tasks in transfer queue
func (s *TestBase) ClearTransferQueue() {
	s.Logger.Info("Clearing transfer tasks", tag.ShardRangeID(s.ShardInfo.RangeID), tag.ReadLevel(s.GetTransferReadLevel()))
	tasks, err := s.GetTransferTasks(context.Background(), 100, true)
	if err != nil {
		s.Logger.Fatal("Error during cleanup", tag.Error(err))
	}

	counter := 0
	for _, t := range tasks {
		s.Logger.Info("Deleting transfer task with ID", tag.TaskID(t.TaskID))
		s.NoError(s.CompleteTransferTask(context.Background(), t.TaskID))
		counter++
	}

	s.Logger.Info("Deleted transfer tasks.", tag.Counter(counter))
	atomic.StoreInt64(&s.ReadLevel, 0)
}

// ClearReplicationQueue completes all tasks in replication queue
func (s *TestBase) ClearReplicationQueue() {
	s.Logger.Info("Clearing replication tasks", tag.ShardRangeID(s.ShardInfo.RangeID), tag.ReadLevel(s.GetReplicationReadLevel()))
	tasks, err := s.GetReplicationTasks(context.Background(), 100, true)
	if err != nil {
		s.Logger.Fatal("Error during cleanup", tag.Error(err))
	}

	counter := 0
	for _, t := range tasks {
		s.Logger.Info("Deleting replication task with ID", tag.TaskID(t.TaskID))
		s.NoError(s.CompleteReplicationTask(context.Background(), t.TaskID))
		counter++
	}

	s.Logger.Info("Deleted replication tasks.", tag.Counter(counter))
	atomic.StoreInt64(&s.ReplicationReadLevel, 0)
}

// EqualTimesWithPrecision assertion that two times are equal within precision
func (s *TestBase) EqualTimesWithPrecision(t1, t2 time.Time, precision time.Duration) {
	s.True(timeComparator(t1, t2, precision),
		"Not equal: \n"+
			"expected: %s\n"+
			"actual  : %s%s", t1, t2,
	)
}

// EqualTimes assertion that two times are equal within two millisecond precision
func (s *TestBase) EqualTimes(t1, t2 time.Time) {
	s.EqualTimesWithPrecision(t1, t2, TimePrecision)
}

func (s *TestBase) validateTimeRange(t time.Time, expectedDuration time.Duration) bool {
	currentTime := time.Now()
	diff := time.Duration(currentTime.UnixNano() - t.UnixNano())
	if diff > expectedDuration {
		s.Logger.Info("Check Current time, Application time, Differenrce", tag.Timestamp(t), tag.CursorTimestamp(currentTime), tag.Number(int64(diff)))
		return false
	}
	return true
}

// GenerateTransferTaskID helper
func (g *TestTransferTaskIDGenerator) GenerateTransferTaskID() (int64, error) {
	return atomic.AddInt64(&g.seqNum, 1), nil
}

// Publish is a utility method to add messages to the queue
func (s *TestBase) Publish(
	ctx context.Context,
	message interface{},
) error {

	retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	retryPolicy.SetBackoffCoefficient(1.5)
	retryPolicy.SetMaximumAttempts(5)

	return backoff.Retry(
		func() error {
			return s.DomainReplicationQueue.Publish(ctx, message)
		},
		retryPolicy,
		func(e error) bool {
			return common.IsPersistenceTransientError(e) || isMessageIDConflictError(e)
		})
}

func isMessageIDConflictError(err error) bool {
	_, ok := err.(*p.ConditionFailedError)
	return ok
}

// GetReplicationMessages is a utility method to get messages from the queue
func (s *TestBase) GetReplicationMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*replicator.ReplicationTask, int64, error) {

	return s.DomainReplicationQueue.GetReplicationMessages(ctx, lastMessageID, maxCount)
}

// UpdateAckLevel updates replication queue ack level
func (s *TestBase) UpdateAckLevel(
	ctx context.Context,
	lastProcessedMessageID int64,
	clusterName string,
) error {

	return s.DomainReplicationQueue.UpdateAckLevel(ctx, lastProcessedMessageID, clusterName)
}

// GetAckLevels returns replication queue ack levels
func (s *TestBase) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	return s.DomainReplicationQueue.GetAckLevels(ctx)
}

// PublishToDomainDLQ is a utility method to add messages to the domain DLQ
func (s *TestBase) PublishToDomainDLQ(
	ctx context.Context,
	message interface{},
) error {

	retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	retryPolicy.SetBackoffCoefficient(1.5)
	retryPolicy.SetMaximumAttempts(5)

	return backoff.Retry(
		func() error {
			return s.DomainReplicationQueue.PublishToDLQ(ctx, message)
		},
		retryPolicy,
		func(e error) bool {
			return common.IsPersistenceTransientError(e) || isMessageIDConflictError(e)
		})
}

// GetMessagesFromDomainDLQ is a utility method to get messages from the domain DLQ
func (s *TestBase) GetMessagesFromDomainDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*replicator.ReplicationTask, []byte, error) {

	return s.DomainReplicationQueue.GetMessagesFromDLQ(
		ctx,
		firstMessageID,
		lastMessageID,
		pageSize,
		pageToken,
	)
}

// UpdateDomainDLQAckLevel updates domain dlq ack level
func (s *TestBase) UpdateDomainDLQAckLevel(
	ctx context.Context,
	lastProcessedMessageID int64,
) error {

	return s.DomainReplicationQueue.UpdateDLQAckLevel(ctx, lastProcessedMessageID)
}

// GetDomainDLQAckLevel returns domain dlq ack level
func (s *TestBase) GetDomainDLQAckLevel(
	ctx context.Context,
) (int64, error) {
	return s.DomainReplicationQueue.GetDLQAckLevel(ctx)
}

// DeleteMessageFromDomainDLQ deletes one message from domain DLQ
func (s *TestBase) DeleteMessageFromDomainDLQ(
	ctx context.Context,
	messageID int64,
) error {

	return s.DomainReplicationQueue.DeleteMessageFromDLQ(ctx, messageID)
}

// RangeDeleteMessagesFromDomainDLQ deletes messages from domain DLQ
func (s *TestBase) RangeDeleteMessagesFromDomainDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {

	return s.DomainReplicationQueue.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
}

// GenerateTransferTaskIDs helper
func (g *TestTransferTaskIDGenerator) GenerateTransferTaskIDs(number int) ([]int64, error) {
	result := []int64{}
	for i := 0; i < number; i++ {
		id, err := g.GenerateTransferTaskID()
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

// GenerateRandomDBName helper
func GenerateRandomDBName(n int) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("workflow")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func pickRandomEncoding() common.EncodingType {
	// randomly pick json/thriftrw/empty as encoding type
	var encoding common.EncodingType
	i := rand.Intn(3)
	switch i {
	case 0:
		encoding = common.EncodingTypeJSON
	case 1:
		encoding = common.EncodingTypeThriftRW
	case 2:
		encoding = common.EncodingType("")
	}
	return encoding
}

func int64Ptr(i int64) *int64 {
	return &i
}
