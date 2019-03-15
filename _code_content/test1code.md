Here's some code.
```go
package simple

import (
    "context"

    "go.uber.org/cadence/activity"
    "go.uber.org/zap"
)

func init() {
    activity.Register(SimpleActivity)
}

// SimpleActivity is a sample Cadence activity function that takes one parameter and
// returns a string containing the parameter value.
func SimpleActivity(ctx context.Context, value string) (string, error) {
    activity.GetLogger(ctx).Info("SimpleActivity called.", zap.String("Value", value))
    return "Processed: " + value, nil
}
```
