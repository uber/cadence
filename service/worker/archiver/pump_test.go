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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
)

var (
	pumpTestMetrics *mmocks.Scope
	pumpTestLogger  *log.MockLogger
)

type pumpSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestPumpSuite(t *testing.T) {
	suite.Run(t, new(pumpSuite))
}

func (s *pumpSuite) SetupSuite() {
	workflow.Register(carryoverSatisfiesLimitWorkflow)
	workflow.Register(pumpWorkflow)
	workflow.Register(signalChClosePumpWorkflow)
	workflow.Register(signalAndCarryoverPumpWorkflow)
}

func (s *pumpSuite) SetupTest() {
	pumpTestMetrics = &mmocks.Scope{}
	pumpTestMetrics.On("StartTimer", mock.Anything).Return(metrics.NewTestStopwatch()).Once()
	pumpTestLogger = &log.MockLogger{}
}

func (s *pumpSuite) TearDownTest() {
	pumpTestMetrics.AssertExpectations(s.T())
	pumpTestLogger.AssertExpectations(s.T())
}

func (s *pumpSuite) TestPumpRun_CarryoverLargerThanLimit() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(1)).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(carryoverSatisfiesLimitWorkflow, 10, 11)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_CarryoverExactlyMatchesLimit() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(carryoverSatisfiesLimitWorkflow, 10, 10)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_TimeoutWithoutSignals() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpTimeoutCount).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpTimeoutWithoutSignalsCount).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(pumpWorkflow, 10, 0)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_TimeoutWithSignals() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpTimeoutCount).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(pumpWorkflow, 10, 5)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_SignalsGottenSatisfyLimit() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpSignalThresholdCount).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(pumpWorkflow, 10, 10)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_SignalsAndCarryover() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpSignalThresholdCount).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(signalAndCarryoverPumpWorkflow, 10, 5, 5)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *pumpSuite) TestPumpRun_SignalChannelClosedUnexpectedly() {
	pumpTestMetrics.On("UpdateGauge", metrics.ArchiverBacklogSizeGauge, float64(0)).Once()
	pumpTestMetrics.On("IncCounter", metrics.ArchiverPumpSignalChannelClosedCount).Once()
	pumpTestLogger.On("Error", mock.Anything, mock.Anything).Once()

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(signalChClosePumpWorkflow, 10, 5)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func carryoverSatisfiesLimitWorkflow(ctx workflow.Context, requestLimit int, carryoverSize int) error {
	unhandledCarryoverSize := carryoverSize - requestLimit
	carryover, carryoverHashes := randomCarryover(carryoverSize)
	requestCh := workflow.NewBufferedChannel(ctx, requestLimit)
	pump := NewPump(ctx, pumpTestLogger, pumpTestMetrics, carryover, time.Nanosecond, requestLimit, requestCh, nil, nil)
	actual := pump.Run()
	expected := PumpResult{
		PumpedHashes:          carryoverHashes[:len(carryoverHashes)-unhandledCarryoverSize],
		UnhandledCarryover:    carryover[len(carryover)-unhandledCarryoverSize:],
		TimeoutWithoutSignals: false,
	}
	if !pumpResultsEqual(expected, actual) {
		return errors.New("did not get expected pump result")
	}
	if !channelContainsExpected(ctx, requestCh, carryover[:len(carryover)-unhandledCarryoverSize]) {
		return errors.New("request channel was not populated with expected values")
	}
	return nil
}

func pumpWorkflow(ctx workflow.Context, requestLimit int, numRequests int) error {
	signalCh := workflow.NewBufferedChannel(ctx, requestLimit)
	signalsSent, signalHashes := sendRequestsToChannel(ctx, signalCh, numRequests)
	requestCh := workflow.NewBufferedChannel(ctx, requestLimit)
	pump := NewPump(ctx, pumpTestLogger, pumpTestMetrics, nil, time.Nanosecond, requestLimit, requestCh, signalCh, NewHistoryRequestReceiver())
	actual := pump.Run()
	expected := PumpResult{
		PumpedHashes:          signalHashes,
		UnhandledCarryover:    nil,
		TimeoutWithoutSignals: numRequests == 0,
	}
	if !pumpResultsEqual(expected, actual) {
		return errors.New("did not get expected pump result")
	}
	if !channelContainsExpected(ctx, requestCh, signalsSent) {
		return errors.New("request channel was not populated with expected values")
	}
	return nil
}

func signalChClosePumpWorkflow(ctx workflow.Context, requestLimit int, numRequests int) error {
	signalCh := workflow.NewBufferedChannel(ctx, requestLimit)
	signalsSent, signalHashes := sendRequestsToChannelBlocking(ctx, signalCh, numRequests)
	signalCh.Close()
	requestCh := workflow.NewBufferedChannel(ctx, requestLimit)
	pump := NewPump(ctx, pumpTestLogger, pumpTestMetrics, nil, time.Nanosecond, requestLimit, requestCh, signalCh, NewHistoryRequestReceiver())
	actual := pump.Run()
	expected := PumpResult{
		PumpedHashes:          signalHashes,
		UnhandledCarryover:    nil,
		TimeoutWithoutSignals: numRequests == 0,
	}
	if !pumpResultsEqual(expected, actual) {
		return errors.New("did not get expected pump result")
	}
	if !channelContainsExpected(ctx, requestCh, signalsSent) {
		return errors.New("request channel was not populated with expected values")
	}
	return nil
}

func signalAndCarryoverPumpWorkflow(ctx workflow.Context, requestLimit int, carryoverSize, numSignals int) error {
	signalCh := workflow.NewBufferedChannel(ctx, requestLimit)
	signalsSent, signalHashes := sendRequestsToChannel(ctx, signalCh, numSignals)
	carryover, carryoverHashes := randomCarryover(carryoverSize)
	requestCh := workflow.NewBufferedChannel(ctx, requestLimit)
	pump := NewPump(ctx, pumpTestLogger, pumpTestMetrics, carryover, time.Nanosecond, requestLimit, requestCh, signalCh, NewHistoryRequestReceiver())
	actual := pump.Run()
	expected := PumpResult{
		PumpedHashes:          append(carryoverHashes, signalHashes...),
		UnhandledCarryover:    nil,
		TimeoutWithoutSignals: false,
	}
	if !pumpResultsEqual(expected, actual) {
		return errors.New("did not get expected pump result")
	}
	if !channelContainsExpected(ctx, requestCh, append(carryover, signalsSent...)) {
		return errors.New("request channel was not populated with expected values")
	}
	return nil
}

func sendRequestsToChannel(ctx workflow.Context, ch workflow.Channel, numRequests int) ([]interface{}, []uint64) {
	requests := make([]interface{}, numRequests, numRequests)
	hashes := make([]uint64, numRequests, numRequests)
	workflow.Go(ctx, func(ctx workflow.Context) {
		for i := 0; i < numRequests; i++ {
			requests[i], hashes[i] = randomArchiveRequest()
			ch.Send(ctx, requests[i])
		}
	})
	return requests, hashes
}

func sendRequestsToChannelBlocking(ctx workflow.Context, ch workflow.Channel, numRequests int) ([]interface{}, []uint64) {
	requests := make([]interface{}, numRequests, numRequests)
	hashes := make([]uint64, numRequests, numRequests)
	for i := 0; i < numRequests; i++ {
		requests[i], hashes[i] = randomArchiveRequest()
		ch.Send(ctx, requests[i])
	}
	return requests, hashes
}

func channelContainsExpected(ctx workflow.Context, ch workflow.Channel, expected []interface{}) bool {
	for i := 0; i < len(expected); i++ {
		var actual ArchiveHistoryRequest
		if !ch.Receive(ctx, &actual) {
			return false
		}
		if hash(expected[i]) != hash(actual) {
			return false
		}
	}
	if ch.Receive(ctx, nil) {
		return false
	}
	return true
}

func randomCarryover(count int) ([]interface{}, []uint64) {
	carryover := make([]interface{}, count, count)
	hashes := make([]uint64, count, count)
	for i := 0; i < count; i++ {
		carryover[i], hashes[i] = randomArchiveRequest()
	}
	return carryover, hashes
}

func pumpResultsEqual(expected PumpResult, actual PumpResult) bool {
	return expected.TimeoutWithoutSignals == actual.TimeoutWithoutSignals &&
		requestsEqual(expected.UnhandledCarryover, actual.UnhandledCarryover) &&
		hashesEqual(expected.PumpedHashes, actual.PumpedHashes)
}

func requestsEqual(expected []interface{}, actual []interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := 0; i < len(expected); i++ {
		if hash(expected[i]) != hash(actual[i]) {
			return false
		}
	}
	return true
}
