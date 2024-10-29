// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package batcher

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

type workflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	workflowEnv *testsuite.TestWorkflowEnvironment
	activityEnv *testsuite.TestActivityEnvironment
}

func TestWorkflowSuite(t *testing.T) {
	suite.Run(t, new(workflowSuite))
}

func (s *workflowSuite) SetupTest() {
	s.workflowEnv = s.NewTestWorkflowEnvironment()
	s.workflowEnv.RegisterWorkflow(BatchWorkflow)

	s.activityEnv = s.NewTestActivityEnvironment()
	s.activityEnv.RegisterActivity(BatchActivity)
}

func (s *workflowSuite) TestWorkflow() {
	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeCancel,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}
	activityHeartBeatDeatils := HeartBeatDetails{}
	s.workflowEnv.OnActivity(batchActivityName, mock.Anything, mock.Anything).Return(activityHeartBeatDeatils, nil)
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.NoError(s.workflowEnv.GetWorkflowError())
}

func (s *workflowSuite) TestActivity_BatchCancel() {
	batcher, _ := setuptest(s.T())
	ctx := context.WithValue(context.Background(), batcherContextKey, batcher)
	workerOpts := worker.Options{
		MetricsScope:              tally.TestScope(nil),
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	s.activityEnv.SetWorkerOptions(workerOpts)
	s.activityEnv.SetHeartbeatDetails(HeartBeatDetails{})

	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeCancel,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}

	_, err := s.activityEnv.ExecuteActivity(BatchActivity, params)
	s.NoError(err)
}

func (s *workflowSuite) TestActivity_BatchTerminate() {
	batcher, _ := setuptest(s.T())
	ctx := context.WithValue(context.Background(), batcherContextKey, batcher)
	workerOpts := worker.Options{
		MetricsScope:              tally.TestScope(nil),
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	s.activityEnv.SetWorkerOptions(workerOpts)

	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeTerminate,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}

	_, err := s.activityEnv.ExecuteActivity(BatchActivity, params)
	s.NoError(err)
}

func (s *workflowSuite) TestActivity_BatchSignal() {
	batcher, _ := setuptest(s.T())
	ctx := context.WithValue(context.Background(), batcherContextKey, batcher)
	workerOpts := worker.Options{
		MetricsScope:              tally.TestScope(nil),
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	s.activityEnv.SetWorkerOptions(workerOpts)

	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeSignal,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}

	_, err := s.activityEnv.ExecuteActivity(BatchActivity, params)
	s.NoError(err)
}

func (s *workflowSuite) TestActivity_BatchReplicate() {
	batcher, _ := setuptest(s.T())
	ctx := context.WithValue(context.Background(), batcherContextKey, batcher)
	workerOpts := worker.Options{
		MetricsScope:              tally.TestScope(nil),
		BackgroundActivityContext: ctx,
		Tracer:                    opentracing.GlobalTracer(),
	}
	s.activityEnv.SetWorkerOptions(workerOpts)

	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeReplicate,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}

	_, err := s.activityEnv.ExecuteActivity(BatchActivity, params)
	s.NoError(err)
}

func (s *workflowSuite) TestWorkflow_ValidationError() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "",
		Reason:     "unit-test",
		BatchType:  BatchTypeCancel,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide required parameters: BatchType/Reason/DomainName/Query")
}

func (s *workflowSuite) TestWorkflow_UnsupportedBatchType() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  "invalid",
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "not supported batch type")
}

func (s *workflowSuite) TestWorkflow_BatchTypeSignalValidation() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeSignal,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide signal name")
}

func (s *workflowSuite) TestWorkflow_BatchTypeReplicateSourceCLusterValidation() {
	paramsWithoutSourceCluster := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeReplicate,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, paramsWithoutSourceCluster)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide source cluster")
}

func (s *workflowSuite) TestWorkflow_BatchTypeReplicateTargetCLusterValidation() {
	paramsWithSourceCluster := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeReplicate,
		ReplicateParams: ReplicateParams{
			SourceCluster: "test-dca",
			TargetCluster: "",
		},
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, paramsWithSourceCluster)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide target cluster")
}

func (s *workflowSuite) TearDownTest() {
	s.workflowEnv.AssertExpectations(s.T())
}
