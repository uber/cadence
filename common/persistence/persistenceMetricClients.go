package persistence

import (
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/metrics"
)

type (
	shardPersistenceClient struct {
		m3Client    metrics.Client
		persistence ShardManager
	}

	workflowExecutionPersistenceClient struct {
		m3Client    metrics.Client
		persistence ExecutionManager
	}

	taskPersistenceClient struct {
		m3Client    metrics.Client
		persistence TaskManager
	}

	historyPersistenceClient struct {
		m3Client    metrics.Client
		persistence HistoryManager
	}

	metadataPersistenceClient struct {
		m3Client    metrics.Client
		persistence MetadataManager
	}
)

var _ ShardManager = (*shardPersistenceClient)(nil)
var _ ExecutionManager = (*workflowExecutionPersistenceClient)(nil)
var _ TaskManager = (*taskPersistenceClient)(nil)
var _ HistoryManager = (*historyPersistenceClient)(nil)
var _ MetadataManager = (*metadataPersistenceClient)(nil)

// NewShardPersistenceClient creates a client to manage shards
func NewShardPersistenceClient(persistence ShardManager, m3Client metrics.Client) ShardManager {
	return &shardPersistenceClient{
		persistence: persistence,
		m3Client:    m3Client,
	}
}

// NewWorkflowExecutionPersistenceClient creates a client to manage executions
func NewWorkflowExecutionPersistenceClient(persistence ExecutionManager, m3Client metrics.Client) ExecutionManager {
	return &workflowExecutionPersistenceClient{
		persistence: persistence,
		m3Client:    m3Client,
	}
}

// NewTaskPersistenceClient creates a client to manage tasks
func NewTaskPersistenceClient(persistence TaskManager, m3Client metrics.Client) TaskManager {
	return &taskPersistenceClient{
		persistence: persistence,
		m3Client:    m3Client,
	}
}

// NewHistoryPersistenceClient creates a HistoryManager client to manage workflow execution history
func NewHistoryPersistenceClient(persistence HistoryManager, m3Client metrics.Client) HistoryManager {
	return &historyPersistenceClient{
		persistence: persistence,
		m3Client:    m3Client,
	}
}

// NewHistoryPersistenceClient creates a HistoryManager client to manage workflow execution history
func NewMetadataPersistenceClient(persistence MetadataManager, m3Client metrics.Client) MetadataManager {
	return &metadataPersistenceClient{
		persistence: persistence,
		m3Client:    m3Client,
	}
}

func (p *shardPersistenceClient) CreateShard(request *CreateShardRequest) error {
	p.m3Client.IncCounter(metrics.CreateShardScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CreateShardScope, metrics.PersistenceLatency)
	err := p.persistence.CreateShard(request)
	sw.Stop()

	if err != nil {
		if _, ok := err.(*ShardAlreadyExistError); ok {
			p.m3Client.IncCounter(metrics.CreateShardScope, metrics.PersistenceErrShardExistsCounter)
		} else {
			p.m3Client.IncCounter(metrics.CreateShardScope, metrics.PersistenceFailures)
		}
	}

	return err
}

func (p *shardPersistenceClient) GetShard(
	request *GetShardRequest) (*GetShardResponse, error) {
	p.m3Client.IncCounter(metrics.GetShardScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetShardScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetShard(request)
	sw.Stop()

	if err != nil {
		switch err.(type) {
		case *workflow.EntityNotExistsError:
			p.m3Client.IncCounter(metrics.GetShardScope, metrics.CadenceErrEntityNotExistsCounter)
		default:
			p.m3Client.IncCounter(metrics.GetShardScope, metrics.PersistenceFailures)
		}
	}

	return response, err
}

func (p *shardPersistenceClient) UpdateShard(request *UpdateShardRequest) error {
	p.m3Client.IncCounter(metrics.UpdateShardScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.UpdateShardScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateShard(request)
	sw.Stop()

	if err != nil {
		if _, ok := err.(*ShardOwnershipLostError); ok {
			p.m3Client.IncCounter(metrics.UpdateShardScope, metrics.PersistenceErrShardOwnershipLostCounter)
		} else {
			p.m3Client.IncCounter(metrics.UpdateShardScope, metrics.PersistenceFailures)
		}
	}

	return err
}

func (p *workflowExecutionPersistenceClient) CreateWorkflowExecution(request *CreateWorkflowExecutionRequest) (*CreateWorkflowExecutionResponse, error) {
	p.m3Client.IncCounter(metrics.CreateWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CreateWorkflowExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateWorkflowExecution(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.CreateWorkflowExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetWorkflowExecution(request *GetWorkflowExecutionRequest) (*GetWorkflowExecutionResponse, error) {
	p.m3Client.IncCounter(metrics.GetWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetWorkflowExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetWorkflowExecution(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.GetWorkflowExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) UpdateWorkflowExecution(request *UpdateWorkflowExecutionRequest) error {
	p.m3Client.IncCounter(metrics.UpdateWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.UpdateWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateWorkflowExecution(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.UpdateWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) DeleteWorkflowExecution(request *DeleteWorkflowExecutionRequest) error {
	p.m3Client.IncCounter(metrics.DeleteWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.DeleteWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteWorkflowExecution(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.DeleteWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) GetCurrentExecution(request *GetCurrentExecutionRequest) (*GetCurrentExecutionResponse, error) {
	p.m3Client.IncCounter(metrics.GetCurrentExecutionScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetCurrentExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetCurrentExecution(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.GetCurrentExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetTransferTasks(request *GetTransferTasksRequest) (*GetTransferTasksResponse, error) {
	p.m3Client.IncCounter(metrics.GetTransferTasksScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetTransferTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetTransferTasks(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.GetTransferTasksScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) CompleteTransferTask(request *CompleteTransferTaskRequest) error {
	p.m3Client.IncCounter(metrics.CompleteTransferTaskScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CompleteTransferTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTransferTask(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.CompleteTransferTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) GetTimerIndexTasks(request *GetTimerIndexTasksRequest) (*GetTimerIndexTasksResponse, error) {
	p.m3Client.IncCounter(metrics.GetTimerIndexTasksScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetTimerIndexTasksScope, metrics.PersistenceLatency)
	resonse, err := p.persistence.GetTimerIndexTasks(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.GetTimerIndexTasksScope, err)
	}

	return resonse, err
}

func (p *workflowExecutionPersistenceClient) CompleteTimerTask(request *CompleteTimerTaskRequest) error {
	p.m3Client.IncCounter(metrics.CompleteTimerTaskScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CompleteTimerTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTimerTask(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.CompleteTimerTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *workflow.WorkflowExecutionAlreadyStartedError:
		p.m3Client.IncCounter(scope, metrics.CadenceErrExecutionAlreadyStartedCounter)
	case *ShardOwnershipLostError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrShardOwnershipLostCounter)
	case *ConditionFailedError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *taskPersistenceClient) CreateTasks(request *CreateTasksRequest) (*CreateTasksResponse, error) {
	p.m3Client.IncCounter(metrics.CreateTaskScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CreateTaskScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateTasks(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.CreateTaskScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) GetTasks(request *GetTasksRequest) (*GetTasksResponse, error) {
	p.m3Client.IncCounter(metrics.GetTasksScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.GetTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetTasks(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.GetTasksScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) CompleteTask(request *CompleteTaskRequest) error {
	p.m3Client.IncCounter(metrics.CompleteTaskScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.CompleteTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTask(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.CompleteTaskScope, err)
	}

	return err
}

func (p *taskPersistenceClient) LeaseTaskList(request *LeaseTaskListRequest) (*LeaseTaskListResponse, error) {
	p.m3Client.IncCounter(metrics.LeaseTaskListScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.LeaseTaskListScope, metrics.PersistenceLatency)
	response, err := p.persistence.LeaseTaskList(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.LeaseTaskListScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) UpdateTaskList(request *UpdateTaskListRequest) (*UpdateTaskListResponse, error) {
	p.m3Client.IncCounter(metrics.UpdateTaskListScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.UpdateTaskListScope, metrics.PersistenceLatency)
	response, err := p.persistence.UpdateTaskList(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.UpdateTaskListScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *ConditionFailedError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *historyPersistenceClient) AppendHistoryEvents(request *AppendHistoryEventsRequest) error {
	p.m3Client.IncCounter(metrics.AppendHistoryEventsScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.AppendHistoryEventsScope, metrics.PersistenceLatency)
	err := p.persistence.AppendHistoryEvents(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.AppendHistoryEventsScope, err)
	}

	return err
}

func (p *historyPersistenceClient) GetWorkflowExecutionHistory(
	request *GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error) {
	p.m3Client.IncCounter(metrics.PersistenceGetWorkflowExecutionHistoryScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceGetWorkflowExecutionHistoryScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetWorkflowExecutionHistory(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetWorkflowExecutionHistoryScope, err)
	}

	return response, err
}

func (p *historyPersistenceClient) DeleteWorkflowExecutionHistory(
	request *DeleteWorkflowExecutionHistoryRequest) error {
	p.m3Client.IncCounter(metrics.DeleteWorkflowExecutionHistoryScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.DeleteWorkflowExecutionHistoryScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteWorkflowExecutionHistory(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.DeleteWorkflowExecutionHistoryScope, err)
	}

	return err
}

func (p *historyPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *workflow.EntityNotExistsError:
		p.m3Client.IncCounter(scope, metrics.CadenceErrEntityNotExistsCounter)
	case *ConditionFailedError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.m3Client.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *metadataPersistenceClient) CreateDomain(request *CreateDomainRequest) (*CreateDomainResponse, error) {
	p.m3Client.IncCounter(metrics.PersistenceCreateDomainScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceCreateDomainScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateDomain(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCreateDomainScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) GetDomain(request *GetDomainRequest) (*GetDomainResponse, error) {
	p.m3Client.IncCounter(metrics.PersistenceGetDomainScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceGetDomainScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetDomain(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetDomainScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) UpdateDomain(request *UpdateDomainRequest) error {
	p.m3Client.IncCounter(metrics.PersistenceUpdateDomainScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceUpdateDomainScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateDomain(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpdateDomainScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) DeleteDomain(request *DeleteDomainRequest) error {
	p.m3Client.IncCounter(metrics.PersistenceDeleteDomainScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceDeleteDomainScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteDomain(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteDomainScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) DeleteDomainByName(request *DeleteDomainByNameRequest) error {
	p.m3Client.IncCounter(metrics.PersistenceDeleteDomainByNameScope, metrics.PersistenceRequests)

	sw := p.m3Client.StartTimer(metrics.PersistenceDeleteDomainByNameScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteDomainByName(request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteDomainByNameScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *workflow.DomainAlreadyExistsError:
		p.m3Client.IncCounter(scope, metrics.CadenceErrDomainAlreadyExistsCounter)
	case *workflow.EntityNotExistsError:
		p.m3Client.IncCounter(scope, metrics.CadenceErrEntityNotExistsCounter)
	case *workflow.BadRequestError:
		p.m3Client.IncCounter(scope, metrics.CadenceErrBadRequestCounter)
	default:
		p.m3Client.IncCounter(scope, metrics.PersistenceFailures)
	}
}