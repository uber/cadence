# Activity and Workflow Retries
\
Activities and workflows can fail due to various intermediate conditions. In those cases, we want
to retry the failed activity or child workflow or even parent workflow. This can be easily achieved
by supplying an optional retry policy. A retry policy looks like:

``` go
// RetryPolicy defines the retry policy
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
    // If not set or set to 0, it means unlimited, and rely on ExpirationInterval to stop.
    // Either MaximumAttempts or ExpirationInterval is required.
    MaximumAttempts int32

    // Non-Retriable errors. This is optional. Cadence server will stop retry if error reason matches this list.
    // Error reason for custom error is specified when your activity/workflow return cadence.NewCustomError(reason).
    // Error reason for panic error is "cadenceInternal:Panic".
    // Error reason for any other error is "cadenceInternal:Generic".
    // Error reason for timeouts is: "cadenceInternal:Timeout TIMEOUT_TYPE". TIMEOUT_TYPE could be START_TO_CLOSE or HEARTBEAT.
    // Note, cancellation is not a failure, so it won't be retried.
    NonRetriableErrorReasons []string
}
```

To enable retry, supply a custom retry policy to the ActivityOptions or ChildWorkflowOptions
when execute them.

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
    RetryPolicy:            retryPolicy, // enable retry
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
        // recover from finished progress
        var finishedIndex int
        if err := activity.GetHeartbeatDetails(ctx, &finishedIndex); err == nil {
            startIdx = finishedIndex + 1 // start from next one.
        }
    }

    // normal activity logic...
    for i:=startIdx; i<inputArg.EndIdx; i++ {
        // code for processing item i goes here...
        activity.RecordHeartbeat(ctx, i) // report progress
    }
}
```

Like retry for activity, you need to supply a retry policy for ChildWorkflowOptions to enable
retry for child workflow. To enable retry for a parent workflow, you supply a retry policy when
start workflow via StartWorkflowOptions.

There are some subtle changes to workflow's history events when RetryPolicy is used.
For activity with RetryPolicy:

* The ActivityTaskScheduledEvent will have extended ScheduleToStartTimeout and ScheduleToCloseTimeout. These 2 timeouts
  will be overwrite by server to be as long as the RetryPolicy's ExpirationInterval. If the ExpirationInterval is not
  specified it will be overwrite to workflow's timeout.
* The ActivityTaskStartedEvent will not show up in history until the activity is completed or failed with no more retry.
  This is to avoid recording the ActivityTaskStarted event but later it failed and retried. Using DescribeWorkflowExecution
  API will return the PendingActivityInfo and it would contain attemptCount if it is retrying.

For workflow with RetryPolicy:

* If workflow failed and need a retry, the workflow execution will be closed with ContinueAsNew event. The event will
  have ContinueAsNewInitiator set to RetryPolicy and the new RunID for next retry attempt.
* The new attempt will be created immediately. But the first decision task won't be scheduled until the backoff duration
  which is also recorded in the new run's WorkflowExecutionStartedEventAttributes event as firstDecisionTaskBackoffSeconds.
