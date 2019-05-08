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

// Code generated by thriftrw v1.18.0. DO NOT EDIT.
// @generated

package cadence

import (
	shared "github.com/uber/cadence/.gen/go/shared"
	thriftreflect "go.uber.org/thriftrw/thriftreflect"
)

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "cadence",
	Package:  "github.com/uber/cadence/.gen/go/cadence",
	FilePath: "cadence.thrift",
	SHA1:     "66642ff367b46d7cc0465fb141282d7e5cbb65e0",
	Includes: []*thriftreflect.ThriftModule{
		shared.ThriftModule,
	},
	Raw: rawIDL,
}

const rawIDL = "// Copyright (c) 2017 Uber Technologies, Inc.\n//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\ninclude \"shared.thrift\"\n\nnamespace java com.uber.cadence\n\n/**\n* WorkflowService API is exposed to provide support for long running applications.  Application is expected to call\n* StartWorkflowExecution to create an instance for each instance of long running workflow.  Such applications are expected\n* to have a worker which regularly polls for DecisionTask and ActivityTask from the WorkflowService.  For each\n* DecisionTask, application is expected to process the history of events for that session and respond back with next\n* decisions.  For each ActivityTask, application is expected to execute the actual logic for that task and respond back\n* with completion or failure.  Worker is expected to regularly heartbeat while activity task is running.\n**/\nservice WorkflowService {\n  /**\n  * RegisterDomain creates a new domain which can be used as a container for all resources.  Domain is a top level\n  * entity within Cadence, used as a container for all resources like workflow executions, tasklists, etc.  Domain\n  * acts as a sandbox and provides isolation for all resources within the domain.  All resources belongs to exactly one\n  * domain.\n  **/\n  void RegisterDomain(1: shared.RegisterDomainRequest registerRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.DomainAlreadyExistsError domainExistsError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * DescribeDomain returns the information and configuration for a registered domain.\n  **/\n  shared.DescribeDomainResponse DescribeDomain(1: shared.DescribeDomainRequest describeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n    * ListDomains returns the information and configuration for all domains.\n    **/\n    shared.ListDomainsResponse ListDomains(1: shared.ListDomainsRequest listRequest)\n      throws (\n        1: shared.BadRequestError badRequestError,\n        2: shared.InternalServiceError internalServiceError,\n        3: shared.EntityNotExistsError entityNotExistError,\n        4: shared.ServiceBusyError serviceBusyError,\n        5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n      )\n\n  /**\n  * UpdateDomain is used to update the information and configuration for a registered domain.\n  **/\n  shared.UpdateDomainResponse UpdateDomain(1: shared.UpdateDomainRequest updateRequest)\n      throws (\n        1: shared.BadRequestError badRequestError,\n        2: shared.InternalServiceError internalServiceError,\n        3: shared.EntityNotExistsError entityNotExistError,\n        4: shared.ServiceBusyError serviceBusyError,\n        5: shared.DomainNotActiveError domainNotActiveError,\n        6: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n      )\n\n  /**\n  * DeprecateDomain us used to update status of a registered domain to DEPRECATED.  Once the domain is deprecated\n  * it cannot be used to start new workflow executions.  Existing workflow executions will continue to run on\n  * deprecated domains.\n  **/\n  void DeprecateDomain(1: shared.DeprecateDomainRequest deprecateRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * StartWorkflowExecution starts a new long running workflow instance.  It will create the instance with\n  * 'WorkflowExecutionStarted' event in history and also schedule the first DecisionTask for the worker to make the\n  * first decision for this instance.  It will return 'WorkflowExecutionAlreadyStartedError', if an instance already\n  * exists with same workflowId.\n  **/\n  shared.StartWorkflowExecutionResponse StartWorkflowExecution(1: shared.StartWorkflowExecutionRequest startRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.WorkflowExecutionAlreadyStartedError sessionAlreadyExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.LimitExceededError limitExceededError,\n      7: shared.EntityNotExistsError entityNotExistError,\n      8: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * Returns the history of specified workflow execution.  It fails with 'EntityNotExistError' if speficied workflow\n  * execution in unknown to the service.\n  **/\n  shared.GetWorkflowExecutionHistoryResponse GetWorkflowExecutionHistory(1: shared.GetWorkflowExecutionHistoryRequest getRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * PollForDecisionTask is called by application worker to process DecisionTask from a specific taskList.  A\n  * DecisionTask is dispatched to callers for active workflow executions, with pending decisions.\n  * Application is then expected to call 'RespondDecisionTaskCompleted' API when it is done processing the DecisionTask.\n  * It will also create a 'DecisionTaskStarted' event in the history for that session before handing off DecisionTask to\n  * application worker.\n  **/\n  shared.PollForDecisionTaskResponse PollForDecisionTask(1: shared.PollForDecisionTaskRequest pollRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.ServiceBusyError serviceBusyError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.EntityNotExistsError entityNotExistError,\n      6: shared.DomainNotActiveError domainNotActiveError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondDecisionTaskCompleted is called by application worker to complete a DecisionTask handed as a result of\n  * 'PollForDecisionTask' API call.  Completing a DecisionTask will result in new events for the workflow execution and\n  * potentially new ActivityTask being created for corresponding decisions.  It will also create a DecisionTaskCompleted\n  * event in the history for that session.  Use the 'taskToken' provided as response of PollForDecisionTask API call\n  * for completing the DecisionTask.\n  * The response could contain a new decision task if there is one or if the request asking for one.\n  **/\n  shared.RespondDecisionTaskCompletedResponse RespondDecisionTaskCompleted(1: shared.RespondDecisionTaskCompletedRequest completeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondDecisionTaskFailed is called by application worker to indicate failure.  This results in\n  * DecisionTaskFailedEvent written to the history and a new DecisionTask created.  This API can be used by client to\n  * either clear sticky tasklist or report any panics during DecisionTask processing.  Cadence will only append first\n  * DecisionTaskFailed event to the history of workflow execution for consecutive failures.\n  **/\n  void RespondDecisionTaskFailed(1: shared.RespondDecisionTaskFailedRequest failedRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * PollForActivityTask is called by application worker to process ActivityTask from a specific taskList.  ActivityTask\n  * is dispatched to callers whenever a ScheduleTask decision is made for a workflow execution.\n  * Application is expected to call 'RespondActivityTaskCompleted' or 'RespondActivityTaskFailed' once it is done\n  * processing the task.\n  * Application also needs to call 'RecordActivityTaskHeartbeat' API within 'heartbeatTimeoutSeconds' interval to\n  * prevent the task from getting timed out.  An event 'ActivityTaskStarted' event is also written to workflow execution\n  * history before the ActivityTask is dispatched to application worker.\n  **/\n  shared.PollForActivityTaskResponse PollForActivityTask(1: shared.PollForActivityTaskRequest pollRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.ServiceBusyError serviceBusyError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.EntityNotExistsError entityNotExistError,\n      6: shared.DomainNotActiveError domainNotActiveError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RecordActivityTaskHeartbeat is called by application worker while it is processing an ActivityTask.  If worker fails\n  * to heartbeat within 'heartbeatTimeoutSeconds' interval for the ActivityTask, then it will be marked as timedout and\n  * 'ActivityTaskTimedOut' event will be written to the workflow history.  Calling 'RecordActivityTaskHeartbeat' will\n  * fail with 'EntityNotExistsError' in such situations.  Use the 'taskToken' provided as response of\n  * PollForActivityTask API call for heartbeating.\n  **/\n  shared.RecordActivityTaskHeartbeatResponse RecordActivityTaskHeartbeat(1: shared.RecordActivityTaskHeartbeatRequest heartbeatRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RecordActivityTaskHeartbeatByID is called by application worker while it is processing an ActivityTask.  If worker fails\n  * to heartbeat within 'heartbeatTimeoutSeconds' interval for the ActivityTask, then it will be marked as timedout and\n  * 'ActivityTaskTimedOut' event will be written to the workflow history.  Calling 'RecordActivityTaskHeartbeatByID' will\n  * fail with 'EntityNotExistsError' in such situations.  Instead of using 'taskToken' like in RecordActivityTaskHeartbeat,\n  * use Domain, WorkflowID and ActivityID\n  **/\n  shared.RecordActivityTaskHeartbeatResponse RecordActivityTaskHeartbeatByID(1: shared.RecordActivityTaskHeartbeatByIDRequest heartbeatRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskCompleted is called by application worker when it is done processing an ActivityTask.  It will\n  * result in a new 'ActivityTaskCompleted' event being written to the workflow history and a new DecisionTask\n  * created for the workflow so new decisions could be made.  Use the 'taskToken' provided as response of\n  * PollForActivityTask API call for completion. It fails with 'EntityNotExistsError' if the taskToken is not valid\n  * anymore due to activity timeout.\n  **/\n  void  RespondActivityTaskCompleted(1: shared.RespondActivityTaskCompletedRequest completeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskCompletedByID is called by application worker when it is done processing an ActivityTask.\n  * It will result in a new 'ActivityTaskCompleted' event being written to the workflow history and a new DecisionTask\n  * created for the workflow so new decisions could be made.  Similar to RespondActivityTaskCompleted but use Domain,\n  * WorkflowID and ActivityID instead of 'taskToken' for completion. It fails with 'EntityNotExistsError'\n  * if the these IDs are not valid anymore due to activity timeout.\n  **/\n  void  RespondActivityTaskCompletedByID(1: shared.RespondActivityTaskCompletedByIDRequest completeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskFailed is called by application worker when it is done processing an ActivityTask.  It will\n  * result in a new 'ActivityTaskFailed' event being written to the workflow history and a new DecisionTask\n  * created for the workflow instance so new decisions could be made.  Use the 'taskToken' provided as response of\n  * PollForActivityTask API call for completion. It fails with 'EntityNotExistsError' if the taskToken is not valid\n  * anymore due to activity timeout.\n  **/\n  void  RespondActivityTaskFailed(1: shared.RespondActivityTaskFailedRequest failRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskFailedByID is called by application worker when it is done processing an ActivityTask.\n  * It will result in a new 'ActivityTaskFailed' event being written to the workflow history and a new DecisionTask\n  * created for the workflow instance so new decisions could be made.  Similar to RespondActivityTaskFailed but use\n  * Domain, WorkflowID and ActivityID instead of 'taskToken' for completion. It fails with 'EntityNotExistsError'\n  * if the these IDs are not valid anymore due to activity timeout.\n  **/\n  void  RespondActivityTaskFailedByID(1: shared.RespondActivityTaskFailedByIDRequest failRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskCanceled is called by application worker when it is successfully canceled an ActivityTask.  It will\n  * result in a new 'ActivityTaskCanceled' event being written to the workflow history and a new DecisionTask\n  * created for the workflow instance so new decisions could be made.  Use the 'taskToken' provided as response of\n  * PollForActivityTask API call for completion. It fails with 'EntityNotExistsError' if the taskToken is not valid\n  * anymore due to activity timeout.\n  **/\n  void RespondActivityTaskCanceled(1: shared.RespondActivityTaskCanceledRequest canceledRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondActivityTaskCanceledByID is called by application worker when it is successfully canceled an ActivityTask.\n  * It will result in a new 'ActivityTaskCanceled' event being written to the workflow history and a new DecisionTask\n  * created for the workflow instance so new decisions could be made.  Similar to RespondActivityTaskCanceled but use\n  * Domain, WorkflowID and ActivityID instead of 'taskToken' for completion. It fails with 'EntityNotExistsError'\n  * if the these IDs are not valid anymore due to activity timeout.\n  **/\n  void RespondActivityTaskCanceledByID(1: shared.RespondActivityTaskCanceledByIDRequest canceledRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.DomainNotActiveError domainNotActiveError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RequestCancelWorkflowExecution is called by application worker when it wants to request cancellation of a workflow instance.\n  * It will result in a new 'WorkflowExecutionCancelRequested' event being written to the workflow history and a new DecisionTask\n  * created for the workflow instance so new decisions could be made. It fails with 'EntityNotExistsError' if the workflow is not valid\n  * anymore due to completion or doesn't exist.\n  **/\n  void RequestCancelWorkflowExecution(1: shared.RequestCancelWorkflowExecutionRequest cancelRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.CancellationAlreadyRequestedError cancellationAlreadyRequestedError,\n      5: shared.ServiceBusyError serviceBusyError,\n      6: shared.DomainNotActiveError domainNotActiveError,\n      7: shared.LimitExceededError limitExceededError,\n      8: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * SignalWorkflowExecution is used to send a signal event to running workflow execution.  This results in\n  * WorkflowExecutionSignaled event recorded in the history and a decision task being created for the execution.\n  **/\n  void SignalWorkflowExecution(1: shared.SignalWorkflowExecutionRequest signalRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.LimitExceededError limitExceededError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * SignalWithStartWorkflowExecution is used to ensure sending signal to a workflow.\n  * If the workflow is running, this results in WorkflowExecutionSignaled event being recorded in the history\n  * and a decision task being created for the execution.\n  * If the workflow is not running or not found, this results in WorkflowExecutionStarted and WorkflowExecutionSignaled\n  * events being recorded in history, and a decision task being created for the execution\n  **/\n  shared.StartWorkflowExecutionResponse SignalWithStartWorkflowExecution(1: shared.SignalWithStartWorkflowExecutionRequest signalWithStartRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.LimitExceededError limitExceededError,\n      7: shared.WorkflowExecutionAlreadyStartedError workflowAlreadyStartedError,\n      8: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n    * ResetWorkflowExecution reset an existing workflow execution to DecisionTaskCompleted event(exclusive).\n    * And it will immediately terminating the current execution instance.\n    **/\n  shared.ResetWorkflowExecutionResponse ResetWorkflowExecution(1: shared.ResetWorkflowExecutionRequest resetRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.LimitExceededError limitExceededError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n    \n  /**\n  * TerminateWorkflowExecution terminates an existing workflow execution by recording WorkflowExecutionTerminated event\n  * in the history and immediately terminating the execution instance.\n  **/\n  void TerminateWorkflowExecution(1: shared.TerminateWorkflowExecutionRequest terminateRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.DomainNotActiveError domainNotActiveError,\n      6: shared.LimitExceededError limitExceededError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * ListOpenWorkflowExecutions is a visibility API to list the open executions in a specific domain.\n  **/\n  shared.ListOpenWorkflowExecutionsResponse ListOpenWorkflowExecutions(1: shared.ListOpenWorkflowExecutionsRequest listRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.LimitExceededError limitExceededError,\n      6: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * ListClosedWorkflowExecutions is a visibility API to list the closed executions in a specific domain.\n  **/\n  shared.ListClosedWorkflowExecutionsResponse ListClosedWorkflowExecutions(1: shared.ListClosedWorkflowExecutionsRequest listRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * ListWorkflowExecutions is a visibility API to list workflow executions in a specific domain.\n  **/\n  shared.ListWorkflowExecutionsResponse ListWorkflowExecutions(1: shared.ListWorkflowExecutionsRequest listRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * ScanWorkflowExecutions is a visibility API to list large amount of workflow executions in a specific domain without order.\n  **/\n  shared.ListWorkflowExecutionsResponse ScanWorkflowExecutions(1: shared.ListWorkflowExecutionsRequest listRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * CountWorkflowExecutions is a visibility API to count of workflow executions in a specific domain.\n  **/\n  shared.CountWorkflowExecutionsResponse CountWorkflowExecutions(1: shared.CountWorkflowExecutionsRequest countRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n      5: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * RespondQueryTaskCompleted is called by application worker to complete a QueryTask (which is a DecisionTask for query)\n  * as a result of 'PollForDecisionTask' API call. Completing a QueryTask will unblock the client call to 'QueryWorkflow'\n  * API and return the query result to client as a response to 'QueryWorkflow' API call.\n  **/\n  void RespondQueryTaskCompleted(1: shared.RespondQueryTaskCompletedRequest completeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.ServiceBusyError serviceBusyError,\n      6: shared.DomainNotActiveError domainNotActiveError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * Reset the sticky tasklist related information in mutable state of a given workflow.\n  * Things cleared are:\n  * 1. StickyTaskList\n  * 2. StickyScheduleToStartTimeout\n  * 3. ClientLibraryVersion\n  * 4. ClientFeatureVersion\n  * 5. ClientImpl\n  **/\n  shared.ResetStickyTaskListResponse ResetStickyTaskList(1: shared.ResetStickyTaskListRequest resetRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.ServiceBusyError serviceBusyError,\n      6: shared.DomainNotActiveError domainNotActiveError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * QueryWorkflow returns query result for a specified workflow execution\n  **/\n  shared.QueryWorkflowResponse QueryWorkflow(1: shared.QueryWorkflowRequest queryRequest)\n\tthrows (\n\t  1: shared.BadRequestError badRequestError,\n\t  2: shared.InternalServiceError internalServiceError,\n\t  3: shared.EntityNotExistsError entityNotExistError,\n\t  4: shared.QueryFailedError queryFailedError,\n\t  5: shared.LimitExceededError limitExceededError,\n      6: shared.ServiceBusyError serviceBusyError,\n      7: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n\t)\n\n  /**\n  * DescribeWorkflowExecution returns information about the specified workflow execution.\n  **/\n  shared.DescribeWorkflowExecutionResponse DescribeWorkflowExecution(1: shared.DescribeWorkflowExecutionRequest describeRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.ServiceBusyError serviceBusyError,\n      6: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n  /**\n  * DescribeTaskList returns information about the target tasklist, right now this API returns the\n  * pollers which polled this tasklist in last few minutes.\n  **/\n  shared.DescribeTaskListResponse DescribeTaskList(1: shared.DescribeTaskListRequest request)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.ServiceBusyError serviceBusyError,\n      6: shared.ClientVersionNotSupportedError clientVersionNotSupportedError,\n    )\n\n}\n"
