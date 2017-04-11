package frontend

import (
	"time"

	"golang.org/x/net/context"

	m "github.com/uber/cadence/.gen/go/cadence"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	tchannel "github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
)

var _ Client = (*clientImpl)(nil)

type clientImpl struct {
	connection *tchannel.Channel
	client     m.TChanWorkflowService
}

// NewClient creates a new frontend TChannel client
func NewClient(ch *tchannel.Channel, hostPort string) (Client, error) {
	var opts *thrift.ClientOptions
	if hostPort != "" {
		opts = &thrift.ClientOptions{
			HostPort: hostPort,
		}
	}
	tClient := thrift.NewClient(ch, common.FrontendServiceName, opts)

	client := &clientImpl{
		connection: ch,
		client:     m.NewTChanWorkflowServiceClient(tClient),
	}
	return client, nil
}

func (c *clientImpl) createContext() (thrift.Context, context.CancelFunc) {
	// TODO: make timeout configurable
	return thrift.NewContext(time.Minute * 3)
}

func (c *clientImpl) RegisterDomain(registerRequest *workflow.RegisterDomainRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RegisterDomain(ctx, registerRequest)
}

func (c *clientImpl) DescribeDomain(
	describeRequest *workflow.DescribeDomainRequest) (*workflow.DescribeDomainResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.DescribeDomain(ctx, describeRequest)
}

func (c *clientImpl) UpdateDomain(
	updateRequest *workflow.UpdateDomainRequest) (*workflow.UpdateDomainResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.UpdateDomain(ctx, updateRequest)
}

func (c *clientImpl) DeprecateDomain(deprecateRequest *workflow.DeprecateDomainRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.DeprecateDomain(ctx, deprecateRequest)
}

func (c *clientImpl) StartWorkflowExecution(request *workflow.StartWorkflowExecutionRequest) (*workflow.StartWorkflowExecutionResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.StartWorkflowExecution(ctx, request)
}

func (c *clientImpl) GetWorkflowExecutionHistory(
	request *workflow.GetWorkflowExecutionHistoryRequest) (*workflow.GetWorkflowExecutionHistoryResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.GetWorkflowExecutionHistory(ctx, request)
}

func (c *clientImpl) PollForActivityTask(pollRequest *workflow.PollForActivityTaskRequest) (*workflow.PollForActivityTaskResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.PollForActivityTask(ctx, pollRequest)
}

func (c *clientImpl) PollForDecisionTask(pollRequest *workflow.PollForDecisionTaskRequest) (*workflow.PollForDecisionTaskResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.PollForDecisionTask(ctx, pollRequest)
}

func (c *clientImpl) RecordActivityTaskHeartbeat(heartbeatRequest *workflow.RecordActivityTaskHeartbeatRequest) (*workflow.RecordActivityTaskHeartbeatResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RecordActivityTaskHeartbeat(ctx, heartbeatRequest)
}

func (c *clientImpl) RespondDecisionTaskCompleted(request *workflow.RespondDecisionTaskCompletedRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RespondDecisionTaskCompleted(ctx, request)
}

func (c *clientImpl) RespondActivityTaskCompleted(request *workflow.RespondActivityTaskCompletedRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RespondActivityTaskCompleted(ctx, request)
}

func (c *clientImpl) RespondActivityTaskFailed(request *workflow.RespondActivityTaskFailedRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RespondActivityTaskFailed(ctx, request)
}

func (c *clientImpl) RespondActivityTaskCanceled(request *workflow.RespondActivityTaskCanceledRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RespondActivityTaskCanceled(ctx, request)
}

func (c *clientImpl) RequestCancelWorkflowExecution(cancelRequest *workflow.RequestCancelWorkflowExecutionRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.RequestCancelWorkflowExecution(ctx, cancelRequest)
}

func (c *clientImpl) TerminateWorkflowExecution(request *workflow.TerminateWorkflowExecutionRequest) error {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.TerminateWorkflowExecution(ctx, request)
}

func (c *clientImpl) ListOpenWorkflowExecutions(
	listRequest *workflow.ListOpenWorkflowExecutionsRequest) (*workflow.ListOpenWorkflowExecutionsResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.ListOpenWorkflowExecutions(ctx, listRequest)
}

func (c *clientImpl) ListClosedWorkflowExecutions(
	listRequest *workflow.ListClosedWorkflowExecutionsRequest) (*workflow.ListClosedWorkflowExecutionsResponse, error) {
	ctx, cancel := c.createContext()
	defer cancel()
	return c.client.ListClosedWorkflowExecutions(ctx, listRequest)
}
