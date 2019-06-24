# Sessions

The session framework provides a straight-forward interface for scheduling multiple activities on a single worker without having users manually specify the task list name. It also includes features like **concurrent session limitation** and **worker failure detection**.

## Use Cases

- **File Processing**: User may want to implement a workflow that can download a file, process it and then upload the modified version. If these three steps are implemented as three different activities, all of them should be executed by the same worker. This is a perfect use case of the session framework.

- **Machine Learning Model Training**: Similar to file processing, training a machine learning model also involves three stages: download the data set, optimize the model and upload the trained parameter. However, as each model may consume a large amount of GPU memory, the number of concurrent models being trained should be limited. If the training of each model is implemented as a session, this requirement can easily be achieved by using the concurrent session limitation feature.

## Basic Usage

Before using the session framework to write your workflow code, your worker needs to be configured to process sessions. To do that, set the `EnableSessionWorker` field of `worker.Options` to `true` when starting your worker.

The most important APIs provided by the session framework are `workflow.CreateSession()` and `workflow.CompleteSession()`. The basic idea is that all the activities executed within a session will be processed by the same worker and these two APIs allow users to create new sessions and close them after all activities finish execution.

Heres a more detailed description of these two APIs:
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

`CreateSession()` takes in workflow.Context, sessionOptions and returns a new session context. When it's called, it will check the task list name specified in the ActivityOptions (or the one specified in StartWorkflowOptions) and create the session on one of the workers which is polling that task list. `CreateSession()` will return an error if the context passed in already contains an open session or all the workers are currently busy and unable to handle new sessions (check the **Concurrent Session Limitation** section for more details).

The returned session context should be used to execute all activities belonging to the session. The context will be cancelled if the worker executing this session dies or `CompleteSession()` is called. When using the returned session context to execute activities, a `workflow.ErrSessionFailed` error may be returned if the session framework detects that the worker executing this session has died. The failure of user activities won't affect the state of the session, so user still needs to handler the errors returned from their activites and call `CompleteSession()` if necessary.

```go
func CompleteSession(ctx Context)
```

`CompleteSession()` takes in a session context and closes it. When it's called, the session context will be canceled (means all user activities using that session context will also be canceled) and the resources reserved at the worker will be released, so make sure `CompleteSession()` is called when you no longer need the session. As it's safe to call `CompleteSession()` on a failed session, typical usage is running it as a `defer` function after a session is successfully created.

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
  err = workflow.ExecuteActivity(sessionCtx, processFileActivityName, *fInfo).Get(ctx, &fInfoProcessed)
  if err != nil {
    return err
  }

  return workflow.ExecuteActivity(sessionCtx, uploadFileActivityName, *fInfoProcessed).Get(ctx, nil)
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

The session context also stores some metadata of the session, which can be retrieved by the `GetSessionInfo()` API. For now, there are only two exported fields in the returned `workflow.SessionInfo`. We may consider exposing more information in the future.

## Concurrent Session Limitation

It's very easy to limit the number of concurrent sessions running on a worker, just set the `MaxConcurrentSessionExecutionSize` field of `worker.Options` to the desired value. By default this field is set to a very large value, so there's no need to manually set it if no limitation is needed.

If a worker hits this limitation, it won't accept any new `CreateSession()` request until one of the existing sessions is completed and `CreateSession()` will return an error if the session can't be created within `CreationTimeout`.

## Recreate Session

For long-running sessions, user may want to use the ContinueAsNew feature to split the workflow into multiple runs, but still needs all the activities to be executed by the same worker. The `RecreateSession()`  API is designed for such use case.

```go
func RecreateSession(ctx Context, recreateToken []byte, sessionOptions *SessionOptions) (Context, error)
```

Its usage is the same as `CreateSession()` except that it also takes in a `recrateToken`, which is needed to create a new session on the same worker as the previous one. User can get the token by calling the `GetRecreateToken()` method of the `SessionInfo` object and pass it to the next run.

```go
token := workflow.GetSessionInfo(sessionCtx).GetRecrateToken()
```

## Q&A

### Is there a complete sample code?
Yes, the fileprocessing example in the cadence-sample repo has been updated to use the session framework.

### What happens to my activity if worker dies?
If your activity has already been scheduled, it will be cancelled. If not, you will get an `workflow.ErrSessionFailed` error when you call `workflow.ExecuteActivity()`.

### Is the concurrent session limitation per process or per host?
It's per worker process, so make sure there's only one worker process running on the host if you plan to use that feature.


## Future Work

* **Support automatic session re-establishing**:
  Right now a session is considered failed if the worker process dies. However, for some use cases, user only cares if the worker host is alive or not. For these uses cases, we should automatically re-establish the session if the worker process is restarted.

* **Support fine-grained concurrent session limitation**:
  The current implementation assumes that all sessions are consuming the same type of resource and there's only one global limitation. In the future, we may allow users to specify what type of resource their session will consume and enforce different limitations on different type of resources.