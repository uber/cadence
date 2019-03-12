# Error handling

An activity, or child workflow, might fail and you could handle errors differently based on different
error cases. If the activity returns an error as `errors.New()` or `fmt.Errorf()`, those errors will
be converted to `workflow.GenericError`. If the activity returns an error as
`cadence.NewCustomError(“err-reason”, details)`, that error will be converted to `*cadence.CustomError`.
There are other types of errors like `workflow.TimeoutError`, `workflow.CanceledError` and
`workflow.PanicError`. Following is an example of what your error code could look like:

```go
err := workflow.ExecuteActivity(ctx, YourActivityFunc).Get(ctx, nil)
switch err := err.(type) {
case *cadence.CustomError:
        switch err.Reason() {
        case "err-reason-a":
                // handle error-reason-a
                var details YourErrorDetailsType
                err.Details(&details)
                // deal with details
        case "err-reason-b":
                // handle error-reason-b
        default:
                // handle all other error reasons
        }
case *workflow.GenericError:
        switch err.Error() {
        case "err-msg-1":
                // handle error with message "err-msg-1"
        case "err-msg-2":
                // handle error with message "err-msg-2"
        default:
                // handle all other generic errors
        }
case *workflow.TimeoutError:
        switch err.TimeoutType() {
        case shared.TimeoutTypeScheduleToStart:
                // handle ScheduleToStart timeout
        case shared.TimeoutTypeStartToClose:
                // handle StartToClose timeout
        case shared.TimeoutTypeHeartbeat:
                // handle heartbeat timeout
        default:
        }
case *workflow.PanicError:
         // handle panic error
case *cadence.CanceledError:
        // handle canceled error
default:
        // all other cases (ideally, this should not happen)
}
```
