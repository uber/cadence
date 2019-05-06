# Starting workflow executions

Given a workflow interface executing a workflow requires initializing a `WorkflowClient` instance, creating
a client side stub to the workflow, and then calling a method annotated with @WorkflowMethod.

```java
WorkflowClient workflowClient = WorkflowClient.newClient(cadenceServiceHost, cadenceServicePort, domain);
// Create a workflow stub.
FileProcessingWorkflow workflow = workflowClient.newWorkflowStub(FileProcessingWorkflow.class);
```

There are two ways to start workflow execution: synchronously and asynchronously. Synchronous invocation starts a workflow
and then waits for its completion. If the process that started the workflow crashes or stops the waiting, the workflow continues executing.
Because workflows are potentially long running, and crashes of clients happen, it is not very commonly found in production use.
Asynchronous start initiates workflow execution and immediately returns to the caller. This is the most common way to start
workflows in production code.

Synchronous start:
```java
// Start a workflow and the wait for a result.
// Note that if the waiting process is killed, the workflow will continue execution.
String result = workflow.processFile(workflowArgs);
```

Asynchronous:
```java
// Returns as soon as the workflow starts.
WorkflowExecution workflowExecution = WorkflowClient.start(workflow::processFile, workflowArgs);

System.out.println("Started process file workflow with workflowId=\"" + workflowExecution.getWorkflowId()
                    + "\" and runId=\"" + workflowExecution.getRunId() + "\"");
```

If you need to wait for a workflow completion after an asynchronous start, the simplest way
is to call the blocking version again. If `WorkflowOptions.WorkflowIdReusePolicy` is not `AllowDuplicate` then instead
of throwing `DuplicateWorkflowException`, it reconnects to an existing workflow and waits for its completion.
The following example shows how to do this from a different process than the one that started the workflow. All this process
needs is a `WorkflowID`.

```java
WorkflowExecution execution = new WorkflowExecution().setWorkflowId(workflowId);
FileProcessingWorkflow workflow = workflowClient.newWorkflowStub(execution);
// Returns result potentially waiting for workflow to complete.
String result = workflow.processFile(workflowArgs);
```
