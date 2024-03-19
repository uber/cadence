package host

import (
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
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

	ctx, cancel := createContext()
	defer cancel()

	// The ratelimit is 5 per second, so we should be able to start 5 workflows without any error
	for i := 0; i < 5; i++ {
		_, err := s.engine.StartWorkflowExecution(ctx, request)
		assert.NoError(s.T(), err)
	}

	// Now we should get a rate limit error
	for i := 0; i < 5; i++ {
		_, err := s.engine.StartWorkflowExecution(ctx, request)
		var busyErr *types.ServiceBusyError
		assert.ErrorAs(s.T(), err, &busyErr)
		assert.Equal(s.T(), common.WorkflowIDRateLimitReason, busyErr.Reason)
	}
}
