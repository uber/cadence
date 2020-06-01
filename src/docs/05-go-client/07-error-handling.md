---
layout: default
title: Error handling
permalink: /docs/go-client/error-handling
---

# Error handling

An :activity:, or child :workflow:, might fail and you could handle errors differently based on different
error cases. If the :activity: returns an error as `errors.New()` or `fmt.Errorf()`, those errors will
be converted to `workflow.GenericError`. If the :activity: returns an error as
`cadence.NewCustomError(“err-reason”, details)`, that error will be converted to `*cadence.CustomError`.
There are other types of errors such as `workflow.TimeoutError`, `workflow.CanceledError` and
`workflow.PanicError`. Following is an example of what your error code might look like:

```go
err := workflow.ExecuteActivity(ctx, YourActivityFunc).Get(ctx, nil)
switch err := err.(type) {
    case *cadence.CustomError:
        switch err.Reason() {
            case "err-reason-a":
                // Handle error-reason-a.
                var details YourErrorDetailsType
                err.Details(&details)
                // Deal with details.
            case "err-reason-b":
                // Handle error-reason-b.
            default:
                // Handle all other error reasons.
        }
    case *workflow.GenericError:
        switch err.Error() {
            case "err-msg-1":
                // Handle error with message "err-msg-1".
            case "err-msg-2":
                // Handle error with message "err-msg-2".
            default:
                // Handle all other generic errors.
        }
    case *workflow.TimeoutError:
        switch err.TimeoutType() {
            case shared.TimeoutTypeScheduleToStart:
                // Handle ScheduleToStart timeout.
            case shared.TimeoutTypeStartToClose:
                // Handle StartToClose timeout.
            case shared.TimeoutTypeHeartbeat:
                // Handle heartbeat timeout.
            default:
        }
    case *workflow.PanicError:
        // Handle panic error.
    case *cadence.CanceledError:
        // Handle canceled error.
    default:
        // All other cases (ideally, this should not happen).
}
```
