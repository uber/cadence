# Sessions

The session framework provides a straightforward interface for scheduling multiple activities on a single worker without requiring you to manually specify the task list name. It also includes features like **concurrent session limitation** and **worker failure detection**.

## Use Cases

- **File Processing**: You may want to implement a workflow that can download a file, process it and then upload the modified version. If these three steps are implemented as three different activities, all of them should be executed by the same worker.

- **Machine Learning Model Training**: Training a machine learning model typically involves three stages: download the data set, optimize the model, and upload the trained parameter. Since the models may consume a large amount of resources (GPU memory for example), the number of models processed on a host needs to be limited.

## Basic Usage

Before using the session framework to write your workflow code, you need to configure your worker to process sessions. To do that, set the `EnableSessionWorker` field of `worker.Options` to `true` when starting your worker.

The most important APIs provided by the session framework are `workflow.CreateSession()` and `workflow.CompleteSession()`. The basic idea is that all the activities executed within a session will be processed by the same worker and these two APIs allow you to create new sessions and close them after all activities finish executing.

Here's a more detailed description of these two APIs:
```go
type SessionOptions struct {
  // ExecutionTimeout: required, no default
  //     Specifies the maximum amount of time the session can run
  ExecutionTimeout time.Duration

  // CreationTimeout: required, no default
  //     Specifies how long session creation can take before returning an error
  CreationTimeout  time.Duration
}

func CreateSession(ctx Context, sessionOptions *SessionOptions) (Context, error)
```

`CreateSession()` takes in workflow.Context, sessionOptions and returns a new context which contains metadata information of the created session (referred to as the **session context** below). When it's called, it will check the task list name specified in the `ActivityOptions` (or in the `StartWorkflowOptions` if the task list name is not specified in the `ActivityOptions`), and create the session on one of the workers which is polling that task list. 

The returned session context should be used to execute all activities belonging to the session. The context will be cancelled if the worker executing this session dies or `CompleteSession()` is called. When using the returned session context to execute activities, a `workflow.ErrSessionFailed` error may be returned if the session framework detects that the worker executing this session has died. The failure of your activities won't affect the state of the session, so you still need to handle the errors returned from your activites and call `CompleteSession()` if necessary.

`CreateSession()` will return an error if the context passed in already contains an open session. If all the workers are currently busy and unable to handle new sessions, the framework will keep retrying until the `CreationTimeout` you specified in the `SessionOptions` has passed before returning an error (check the **Concurrent Session Limitation** section for more details).

```go
func CompleteSession(ctx Context)
```

`CompleteSession()` releases the resources reserved on the worker, so it's important to call it as soon as you no longer need the session. It will cancel the session context and therefore all the activities using that session context. Note that it's safe to call `CompleteSession()` on a failed session, meaning that you can call it from a `defer` function after the session is successfully created.

### Sample Code

```go
func FileProcessingWorkflow(ctx workflow.Context, fileID string) (err error) {
  ao := workflow.ActivityOptions{
    ScheduleToStartTimeout: time.Second * 5,
    StartToCloseTimeout:    time.Minute,
  }
  ctx = workflow.WithActivityOptions(ctx, ao)

  so := &workflow.SessionOptions{
    CreationTimeout:  time.Minute,
    ExecutionTimeout: time.Minute,
  }
  sessionCtx, err := workflow.CreateSession(ctx, so)
  if err != nil {
    return err
  }
  defer workflow.CompleteSession(sessionCtx)

  var fInfo *fileInfo
  err = workflow.ExecuteActivity(sessionCtx, downloadFileActivityName, fileID).Get(sessionCtx, &fInfo)
  if err != nil {
    return err
  }

  var fInfoProcessed *fileInfo
  err = workflow.ExecuteActivity(sessionCtx, processFileActivityName, *fInfo).Get(sessionCtx, &fInfoProcessed)
  if err != nil {
    return err
  }

  return workflow.ExecuteActivity(sessionCtx, uploadFileActivityName, *fInfoProcessed).Get(sessionCtx, nil)
}
```

## Session Metadata

```go
type SessionInfo struct {
  // A unique ID for the session
  SessionID         string

  // The hostname of the worker that is executing the session
  HostName          string

  // ... other unexported fields
}

func GetSessionInfo(ctx Context) *SessionInfo
```

The session context also stores some session metadata, which can be retrieved by the `GetSessionInfo()` API. If the context passed in doesn't contain any session metadata, this API will return a `nil` pointer. 

## Concurrent Session Limitation

To limit the number of concurrent sessions running on a worker, set the `MaxConcurrentSessionExecutionSize` field of `worker.Options` to the desired value. By default this field is set to a very large value, so there's no need to manually set it if no limitation is needed.

If a worker hits this limitation, it won't accept any new `CreateSession()` requests until one of the existing sessions is completed. `CreateSession()` will return an error if the session can't be created within `CreationTimeout`.

## Recreate Session

For long-running sessions, you may want to use the ContinueAsNew feature to split the workflow into multiple runs when all activities need to be executed by the same worker. The `RecreateSession()`  API is designed for such a use case.

```go
func RecreateSession(ctx Context, recreateToken []byte, sessionOptions *SessionOptions) (Context, error)
```

Its usage is the same as `CreateSession()` except that it also takes in a `recreateToken`, which is needed to create a new session on the same worker as the previous one. You can get the token by calling the `GetRecreateToken()` method of the `SessionInfo` object.

```go
token := workflow.GetSessionInfo(sessionCtx).GetRecreateToken()
```

## Q&A

### Is there a complete example?
Yes, the [file processing example](https://github.com/samarabbas/cadence-samples/blob/master/cmd/samples/fileprocessing/workflow.go) in the cadence-sample repo has been updated to use the session framework.

### What happens to my activity if the worker dies?
If your activity has already been scheduled, it will be cancelled. If not, you will get a `workflow.ErrSessionFailed` error when you call `workflow.ExecuteActivity()`.

### Is the concurrent session limitation per process or per host?
It's per worker process, so make sure there's only one worker process running on the host if you plan to use that feature.


## Future Work

* **[Support automatic session re-establishing](https://github.com/uber-go/cadence-client/issues/775)**   
Right now a session is considered failed if the worker process dies. However, for some use cases, you may only care whether worker host is alive or not. For these uses cases, the session should be automatically re-established if the worker process is restarted.

* **[Support fine-grained concurrent session limitation](https://github.com/uber-go/cadence-client/issues/776)**   
The current implementation assumes that all sessions are consuming the same type of resource and there's only one global limitation. Our plan is to allow you to specify what type of resource your session will consume and enforce different limitations on different types of resources.