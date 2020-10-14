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

package matching

import (
	"context"
	"sync"
	"time"

	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	m "github.com/uber/cadence/.gen/go/matching"
	"github.com/uber/cadence/.gen/go/matching/matchingserviceserver"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/resource"
)

var _ matchingserviceserver.Interface = (*HandlerImpl)(nil)

type (
	//go:generate mockgen -copyright_file=../../LICENSE -package $GOPACKAGE -source $GOFILE -destination handler_mock.go -package matching github.com/uber/cadence/service/matching Handler

	// Handler interface for matching service
	Handler interface {
		Health(context.Context) (*health.HealthStatus, error)
		AddActivityTask(context.Context, *m.AddActivityTaskRequest) error
		AddDecisionTask(context.Context, *m.AddDecisionTaskRequest) error
		CancelOutstandingPoll(context.Context, *m.CancelOutstandingPollRequest) error
		DescribeTaskList(context.Context, *m.DescribeTaskListRequest) (*gen.DescribeTaskListResponse, error)
		ListTaskListPartitions(context.Context, *m.ListTaskListPartitionsRequest) (*gen.ListTaskListPartitionsResponse, error)
		PollForActivityTask(context.Context, *m.PollForActivityTaskRequest) (*gen.PollForActivityTaskResponse, error)
		PollForDecisionTask(context.Context, *m.PollForDecisionTaskRequest) (*m.PollForDecisionTaskResponse, error)
		QueryWorkflow(context.Context, *m.QueryWorkflowRequest) (*gen.QueryWorkflowResponse, error)
		RespondQueryTaskCompleted(context.Context, *m.RespondQueryTaskCompletedRequest) error
	}

	// HandlerImpl is an implementation for matching service independent of wire protocol
	HandlerImpl struct {
		resource.Resource

		engine        Engine
		config        *Config
		metricsClient metrics.Client
		startWG       sync.WaitGroup
		rateLimiter   quotas.Limiter
	}
)

var (
	errMatchingHostThrottle = &gen.ServiceBusyError{Message: "Matching host rps exceeded"}
)

// NewHandler creates a thrift handler for the history service
func NewHandler(
	resource resource.Resource,
	config *Config,
) *HandlerImpl {
	handler := &HandlerImpl{
		Resource:      resource,
		config:        config,
		metricsClient: resource.GetMetricsClient(),
		rateLimiter: quotas.NewDynamicRateLimiter(func() float64 {
			return float64(config.RPS())
		}),
		engine: NewEngine(
			resource.GetTaskManager(),
			resource.GetHistoryClient(),
			resource.GetMatchingRawClient(), // Use non retry client inside matching
			config,
			resource.GetLogger(),
			resource.GetMetricsClient(),
			resource.GetDomainCache(),
			resource.GetMatchingServiceResolver(),
		),
	}
	// prevent us from trying to serve requests before matching engine is started and ready
	handler.startWG.Add(1)
	return handler
}

// RegisterHandler register this handler, must be called before Start()
func (h *HandlerImpl) RegisterHandler() {
	h.Resource.GetDispatcher().Register(matchingserviceserver.New(h))
	h.Resource.GetDispatcher().Register(metaserver.New(h))
}

// Start starts the handler
func (h *HandlerImpl) Start() {
	h.startWG.Done()
}

// Stop stops the handler
func (h *HandlerImpl) Stop() {
	h.engine.Stop()
}

// Health is for health check
func (h *HandlerImpl) Health(ctx context.Context) (*health.HealthStatus, error) {
	h.startWG.Wait()
	h.GetLogger().Debug("Matching service health check endpoint reached.")
	hs := &health.HealthStatus{Ok: true, Msg: common.StringPtr("matching good")}
	return hs, nil
}

func (h *HandlerImpl) newHandlerContext(
	ctx context.Context,
	domainID string,
	taskList *gen.TaskList,
	scope int,
) *handlerContext {
	return newHandlerContext(
		ctx,
		h.domainName(domainID),
		taskList,
		h.metricsClient,
		scope,
	)
}

// AddActivityTask - adds an activity task.
func (h *HandlerImpl) AddActivityTask(
	ctx context.Context,
	request *m.AddActivityTaskRequest,
) (retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	startT := time.Now()
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetTaskList(),
		metrics.MatchingAddActivityTaskScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if request.GetForwardedFrom() != "" {
		hCtx.scope.IncCounter(metrics.ForwardedPerTaskListCounter)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return hCtx.handleErr(errMatchingHostThrottle)
	}

	syncMatch, err := h.engine.AddActivityTask(hCtx, request)
	if syncMatch {
		hCtx.scope.RecordTimer(metrics.SyncMatchLatencyPerTaskList, time.Since(startT))
	}

	return hCtx.handleErr(err)
}

// AddDecisionTask - adds a decision task.
func (h *HandlerImpl) AddDecisionTask(
	ctx context.Context,
	request *m.AddDecisionTaskRequest,
) (retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	startT := time.Now()
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetTaskList(),
		metrics.MatchingAddDecisionTaskScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if request.GetForwardedFrom() != "" {
		hCtx.scope.IncCounter(metrics.ForwardedPerTaskListCounter)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return hCtx.handleErr(errMatchingHostThrottle)
	}

	syncMatch, err := h.engine.AddDecisionTask(hCtx, request)
	if syncMatch {
		hCtx.scope.RecordTimer(metrics.SyncMatchLatencyPerTaskList, time.Since(startT))
	}
	return hCtx.handleErr(err)
}

// PollForActivityTask - long poll for an activity task.
func (h *HandlerImpl) PollForActivityTask(
	ctx context.Context,
	request *m.PollForActivityTaskRequest,
) (resp *gen.PollForActivityTaskResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetPollRequest().GetTaskList(),
		metrics.MatchingPollForActivityTaskScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if request.GetForwardedFrom() != "" {
		hCtx.scope.IncCounter(metrics.ForwardedPerTaskListCounter)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, hCtx.handleErr(errMatchingHostThrottle)
	}

	if _, err := common.ValidateLongPollContextTimeoutIsSet(
		ctx,
		"PollForActivityTask",
		h.Resource.GetThrottledLogger(),
	); err != nil {
		return nil, hCtx.handleErr(err)
	}

	response, err := h.engine.PollForActivityTask(hCtx, request)
	return response, hCtx.handleErr(err)
}

// PollForDecisionTask - long poll for a decision task.
func (h *HandlerImpl) PollForDecisionTask(
	ctx context.Context,
	request *m.PollForDecisionTaskRequest,
) (resp *m.PollForDecisionTaskResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetPollRequest().GetTaskList(),
		metrics.MatchingPollForDecisionTaskScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if request.GetForwardedFrom() != "" {
		hCtx.scope.IncCounter(metrics.ForwardedPerTaskListCounter)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, hCtx.handleErr(errMatchingHostThrottle)
	}

	if _, err := common.ValidateLongPollContextTimeoutIsSet(
		ctx,
		"PollForDecisionTask",
		h.Resource.GetThrottledLogger(),
	); err != nil {
		return nil, hCtx.handleErr(err)
	}

	response, err := h.engine.PollForDecisionTask(hCtx, request)
	return response, hCtx.handleErr(err)
}

// QueryWorkflow queries a given workflow synchronously and return the query result.
func (h *HandlerImpl) QueryWorkflow(
	ctx context.Context,
	request *m.QueryWorkflowRequest,
) (resp *gen.QueryWorkflowResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetTaskList(),
		metrics.MatchingQueryWorkflowScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if request.GetForwardedFrom() != "" {
		hCtx.scope.IncCounter(metrics.ForwardedPerTaskListCounter)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, hCtx.handleErr(errMatchingHostThrottle)
	}

	response, err := h.engine.QueryWorkflow(hCtx, request)
	return response, hCtx.handleErr(err)
}

// RespondQueryTaskCompleted responds a query task completed
func (h *HandlerImpl) RespondQueryTaskCompleted(
	ctx context.Context,
	request *m.RespondQueryTaskCompletedRequest,
) (retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetTaskList(),
		metrics.MatchingRespondQueryTaskCompletedScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	// Count the request in the RPS, but we still accept it even if RPS is exceeded
	h.rateLimiter.Allow()

	err := h.engine.RespondQueryTaskCompleted(hCtx, request)
	return hCtx.handleErr(err)
}

// CancelOutstandingPoll is used to cancel outstanding pollers
func (h *HandlerImpl) CancelOutstandingPoll(ctx context.Context,
	request *m.CancelOutstandingPollRequest) (retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetTaskList(),
		metrics.MatchingCancelOutstandingPollScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	// Count the request in the RPS, but we still accept it even if RPS is exceeded
	h.rateLimiter.Allow()

	err := h.engine.CancelOutstandingPoll(hCtx, request)
	return hCtx.handleErr(err)
}

// DescribeTaskList returns information about the target tasklist, right now this API returns the
// pollers which polled this tasklist in last few minutes. If includeTaskListStatus field is true,
// it will also return status of tasklist's ackManager (readLevel, ackLevel, backlogCountHint and taskIDBlock).
func (h *HandlerImpl) DescribeTaskList(
	ctx context.Context,
	request *m.DescribeTaskListRequest,
) (resp *gen.DescribeTaskListResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := h.newHandlerContext(
		ctx,
		request.GetDomainUUID(),
		request.GetDescRequest().GetTaskList(),
		metrics.MatchingDescribeTaskListScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, hCtx.handleErr(errMatchingHostThrottle)
	}

	response, err := h.engine.DescribeTaskList(hCtx, request)
	return response, hCtx.handleErr(err)
}

// ListTaskListPartitions returns information about partitions for a taskList
func (h *HandlerImpl) ListTaskListPartitions(
	ctx context.Context,
	request *m.ListTaskListPartitionsRequest,
) (resp *gen.ListTaskListPartitionsResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	hCtx := newHandlerContext(
		ctx,
		request.GetDomain(),
		request.GetTaskList(),
		h.metricsClient,
		metrics.MatchingListTaskListPartitionsScope,
	)

	sw := hCtx.startProfiling(&h.startWG)
	defer sw.Stop()

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, hCtx.handleErr(errMatchingHostThrottle)
	}

	response, err := h.engine.ListTaskListPartitions(hCtx, request)
	return response, hCtx.handleErr(err)
}

func (h *HandlerImpl) domainName(id string) string {
	entry, err := h.GetDomainCache().GetDomainByID(id)
	if err != nil {
		return ""
	}
	return entry.GetInfo().Name
}
