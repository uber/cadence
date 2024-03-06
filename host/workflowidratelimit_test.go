package host

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pborman/uuid"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func (s *IntegrationSuite) TestWorkflowIDSpecificRateLimits() {
	time.Sleep(1 * time.Minute) // wait for the rate limit to be updated
	const (
		testWorkflowID   = "integration-workflow-specific-rate-limit-test"
		testWorkflowType = "integration-workflow-specific-rate-limit-test-type"
		testTaskListName = "integration-workflow-specific-rate-limit-test-taskList"
		testIdentity     = "worker1"
	)

	request := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              s.domainName,
		WorkflowID:                          testWorkflowID,
		WorkflowType:                        &types.WorkflowType{Name: testWorkflowType},
		TaskList:                            &types.TaskList{Name: testTaskListName},
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            testIdentity,

		WorkflowIDReusePolicy: types.WorkflowIDReusePolicyTerminateIfRunning.Ptr(),
	}

	wg := sync.WaitGroup{}
	var busyErrCount atomic.Int32
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := createContext()
			defer cancel()
			start := time.Now()
			_, err := s.engine.StartWorkflowExecution(ctx, request)
			end := time.Now()
			fmt.Println("StartWorkflowExecution time: ", end.Sub(start), start, end)

			var busyErr *types.ServiceBusyError
			if errors.As(err, &busyErr) {
				busyErrCount.Add(1)
			}
		}()
	}

	wg.Wait()

	s.GreaterOrEqual(busyErrCount.Load(), int32(50))
}
