package canary

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
)

const (
	cadenceConfig = "cadence"
)

// cadenceClient is an abstraction on top of
// the cadence library client that serves as
// a union of all the client interfaces that
// the library exposes
type cadenceClient struct {
	client.Client
	// domainClient only exposes domain API
	client.DomainClient
	// this is the service needed to start the workers
	Service workflowserviceclient.Interface
}

// createDomain creates a cadence domain with the given name and description
// if the domain already exist, this method silently returns success
func (client *cadenceClient) createDomain(name string, desc string, owner string, archivalStatus *shared.ArchivalStatus) error {
	emitMetric := true
	retention := int32(workflowRetentionDays)
	if archivalStatus != nil && *archivalStatus == shared.ArchivalStatusEnabled {
		retention = int32(0)
	}
	req := &shared.RegisterDomainRequest{
		Name:                                   &name,
		Description:                            &desc,
		OwnerEmail:                             &owner,
		WorkflowExecutionRetentionPeriodInDays: &retention,
		EmitMetric:                             &emitMetric,
		HistoryArchivalStatus:                  archivalStatus,
	}
	err := client.Register(context.Background(), req)
	if err != nil {
		if _, ok := err.(*shared.DomainAlreadyExistsError); !ok {
			return err
		}
	}
	return nil
}

// newCadenceClient builds a cadenceClient from the runtimeContext
func newCadenceClient(domain string, runtime *RuntimeContext) cadenceClient {
	tracer := opentracing.GlobalTracer()
	cclient := client.NewClient(
		runtime.service,
		domain,
		&client.Options{
			MetricsScope: runtime.metrics,
			Tracer:       tracer,
		},
	)
	domainClient := client.NewDomainClient(
		runtime.service,
		&client.Options{
			MetricsScope: runtime.metrics,
			Tracer:       tracer,
		},
	)
	return cadenceClient{
		Client:       cclient,
		DomainClient: domainClient,
		Service:      runtime.service,
	}
}

// newWorkflowOptions builds workflowOptions with defaults for everything except startToCloseTimeout
func newWorkflowOptions(id string, executionTimeout time.Duration) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:                              id,
		TaskList:                        taskListName,
		ExecutionStartToCloseTimeout:    executionTimeout,
		DecisionTaskStartToCloseTimeout: decisionTaskTimeout,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
}

// newActivityOptions builds and returns activityOptions with reasonable defaults
func newActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		TaskList:               taskListName,
		StartToCloseTimeout:    activityTaskTimeout,
		ScheduleToStartTimeout: scheduleToStartTimeout,
		ScheduleToCloseTimeout: scheduleToStartTimeout + activityTaskTimeout,
	}
}

// newChildWorkflowOptions builds and returns childWorkflowOptions for given domain
func newChildWorkflowOptions(domain string, wfID string) workflow.ChildWorkflowOptions {
	return workflow.ChildWorkflowOptions{
		Domain:                       domain,
		WorkflowID:                   wfID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: childWorkflowTimeout,
		TaskStartToCloseTimeout:      decisionTaskTimeout,
		WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
	}
}

// registerWorkflow registers a workflow function with a given friendly name
func registerWorkflow(workflowFunc interface{}, name string) {
	opts := workflow.RegisterOptions{Name: name}
	workflow.RegisterWithOptions(workflowFunc, opts)
}

// registerActivity registers an activity function with a given friendly name
func registerActivity(activityFunc interface{}, name string) {
	opts := activity.RegisterOptions{Name: name}
	activity.RegisterWithOptions(activityFunc, opts)
}
