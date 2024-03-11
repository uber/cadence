package host

import (
	"errors"

	"github.com/pborman/uuid"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func (s *IntegrationSuite) TestWorkflowIDSpecificRateLimits() {
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

	busyErrCount := 0
	for i := 0; i < 10; i++ {
		ctx, cancel := createContext()
		defer cancel()

		_, err := s.engine.StartWorkflowExecution(ctx, request)

		var busyErr *types.ServiceBusyError
		if errors.As(err, &busyErr) {
			busyErrCount += 1
		} else {
			s.NoError(err)
		}
	}

	// The per workflow rate limit is 5, so we should see 5 busy errors
	s.Equal(5, busyErrCount)
}
