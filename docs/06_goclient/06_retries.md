# Activity and Workflow Retries

Activities and workflows can fail due to various intermediate conditions. In those cases, we want
to retry the failed activity or child workflow or even the parent workflow. This can be achieved
by supplying an optional retry policy. A retry policy looks like the following:

``` go
// RetryPolicy defines the retry policy.
RetryPolicy struct {
    // Backoff interval for the first retry. If coefficient is 1.0 then it is used for all retries.
    // Required, no default value.
    InitialInterval time.Duration

    // Coefficient used to calculate the next retry backoff interval.
    // The next retry interval is previous interval multiplied by this coefficient.
    // Must be 1 or larger. Default is 2.0.
    BackoffCoefficient float64

    // Maximum backoff interval between retries. Exponential backoff leads to interval increase.
    // This value is the cap of the interval. Default is 100x of initial interval.
    MaximumInterval time.Duration

    // Maximum time to retry. Either ExpirationInterval or MaximumAttempts is required.
    // When exceeded the retries stop even if maximum retries is not reached yet.
    ExpirationInterval time.Duration

    // Maximum number of attempts. When exceeded the retries stop even if not expired yet.
    // If not set or set to 0, it means unlimited, and relies on ExpirationInterval to stop.
    // Either MaximumAttempts or ExpirationInterval is required.
    MaximumAttempts int32

    // Non-Retriable errors. This is optional. Cadence server will stop retry if error reason matches this list.
    // Error reason for custom error is specified when your activity/workflow returns cadence.NewCustomError(reason).
    // Error reason for panic error is "cadenceInternal:Panic".
    // Error reason for any other error is "cadenceInternal:Generic".
    // Error reason for timeouts is: "cadenceInternal:Timeout TIMEOUT_TYPE". TIMEOUT_TYPE could be START_TO_CLOSE or HEARTBEAT.
    // Note that cancellation is not a failure, so it won't be retried.
    NonRetriableErrorReasons []string
}
```

To enable retry, supply a custom retry policy to `ActivityOptions` or `ChildWorkflowOptions`
when you execute them.

``` go
expiration := time.Minute * 10
retryPolicy := &cadence.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2,
    MaximumInterval:    expiration,
    ExpirationInterval: time.Minute * 10,
    MaximumAttempts:    5,
}
ao := workflow.ActivityOptions{
    ScheduleToStartTimeout: expiration,
    StartToCloseTimeout:    expiration,
    HeartbeatTimeout:       time.Second * 30,
    RetryPolicy:            retryPolicy, // Enable retry.
}
ctx = workflow.WithActivityOptions(ctx, ao)
activityFuture := workflow.ExecuteActivity(ctx, SampleActivity, params)
```

If activity heartbeat its progress before it failed, the retry attempt will contain the progress
so activity implementation could resume from failed progress like:

``` go
func SampleActivity(ctx context.Context, inputArg InputParams) error {
    startIdx := inputArg.StartIndex
    if activity.HasHeartbeatDetails(ctx) {
        // Recover from finished progress.
        var finishedIndex int
        if err := activity.GetHeartbeatDetails(ctx, &finishedIndex); err == nil {
            startIdx = finishedIndex + 1 // Start from next one.
        }
    }

    // Normal activity logic...
    for i:=startIdx; i<inputArg.EndIdx; i++ {
        // Code for processing item i goes here...
        activity.RecordHeartbeat(ctx, i) // Report progress.
    }
}
```

Like retry for an activity, you need to supply a retry policy for `ChildWorkflowOptions` to enable
retry for a child workflow. To enable retry for a parent workflow, supply a retry policy when
you start a workflow via `StartWorkflowOptions`.

There are some subtle changes to workflow's history events when `RetryPolicy` is used.
For an activity with `RetryPolicy`:

* The `ActivityTaskScheduledEvent` will have extended `ScheduleToStartTimeout` and `ScheduleToCloseTimeout`. These two timeouts
  will be overwritten by the server to be as long as the retry policy's `ExpirationInterval`. If the `ExpirationInterval`
  is not specified, it will be overwritten to the workflow's timeout.
* The `ActivityTaskStartedEvent` will not show up in history until the activity is completed or failed with no more retry.
  This is to avoid recording the `ActivityTaskStarted` event but later it failed and retried. Using the `DescribeWorkflowExecution`
  API will return the `PendingActivityInfo` and it will contain `attemptCount` if it is retrying.

For a workflow with `RetryPolicy`:

* If a workflow failed and needs to retry, the workflow execution will be closed with a `ContinueAsNew` event. The event
  will have the `ContinueAsNewInitiator` set to `RetryPolicy` and the new `RunID` for the next retry attempt.
* The new attempt will be created immediately. But the first decision task won't be scheduled until the backoff duration
  which is also recorded in the new run's `WorkflowExecutionStartedEventAttributes` event as `firstDecisionTaskBackoffSeconds`.
