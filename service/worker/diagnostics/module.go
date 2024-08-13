package diagnostics

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

type DiagnosticsWorkflow interface {
	Start() error
	Stop()
}

type dw struct {
	svcClient     workflowserviceclient.Interface
	clientBean    client.Bean
	metricsClient metrics.Client
	tallyScope    tally.Scope
	worker        worker.Worker
}

type Params struct {
	ServiceClient workflowserviceclient.Interface
	ClientBean    client.Bean
	MetricsClient metrics.Client
	TallyScope    tally.Scope
}

// New creates a new diagnostics workflow.
func New(params Params) DiagnosticsWorkflow {
	return &dw{
		svcClient:     params.ServiceClient,
		metricsClient: params.MetricsClient,
		tallyScope:    params.TallyScope,
		clientBean:    params.ClientBean,
	}
}

// Start starts the worker
func (w *dw) Start() error {
	workerOpts := worker.Options{
		MetricsScope:                     w.tallyScope,
		BackgroundActivityContext:        context.Background(),
		Tracer:                           opentracing.GlobalTracer(),
		MaxConcurrentActivityTaskPollers: 10,
		MaxConcurrentDecisionTaskPollers: 10,
	}
	newWorker := worker.New(w.svcClient, common.SystemLocalDomainName, tasklist, workerOpts)
	newWorker.RegisterWorkflowWithOptions(w.DiagnosticsWorkflow, workflow.RegisterOptions{Name: diagnosticsWorkflow})
	newWorker.RegisterActivityWithOptions(w.retrieveExecutionHistory, activity.RegisterOptions{Name: retrieveWfExecutionHistoryActivity})
	newWorker.RegisterActivityWithOptions(w.identifyTimeouts, activity.RegisterOptions{Name: identifyTimeoutsActivity})
	w.worker = newWorker
	return newWorker.Start()
}

func (w *dw) Stop() {
	w.worker.Stop()
}
