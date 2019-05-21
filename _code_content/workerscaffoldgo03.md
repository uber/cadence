---
name: workerscaffoldgo03
---

**sample_helper.go**
```go
package common

import (
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
)

type (
	// SampleHelper class for workflow sample helper.
	SampleHelper struct {
		Service workflowserviceclient.Interface
		Scope   tally.Scope
		Logger  *zap.Logger
	}
)

// SetupServiceConfig configures worker for local Cadence server
func (h *SampleHelper) SetupServiceConfig() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	h.Logger = logger
	h.Scope = tally.NoopScope
	err = h.buildServiceClient()
	if err != nil {
		panic(err)
	}
}
```
