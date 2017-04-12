package history

import (
	"sync"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common"
)

const (
	historyCacheInitialSize               = 256
	historyCacheMaxSize                   = 1 * 1024
	historyCacheTTL         time.Duration = time.Hour
)

type (
	executionLock struct {
		sync.Mutex
		waitingCount int
	}

	releaseWorkflowExecutionFunc func()

	historyCache struct {
		cache.Cache
		shard            ShardContext
		executionManager persistence.ExecutionManager
		disabled         bool
		logger           bark.Logger

		sync.Mutex
		lockTable      map[string]*executionLock
		pinnedContexts map[string]*workflowExecutionContext
	}
)

var (
	// ErrTryLock is a temporary error that is thrown by the API
	// when it loses the race to create workflow execution context
	ErrTryLock = &workflow.InternalServiceError{Message: "Failed to acquire lock, backoff and retry"}
)

func newHistoryCache(shard ShardContext, logger bark.Logger) *historyCache {
	opts := &cache.Options{}
	opts.InitialCapacity = historyCacheInitialSize
	opts.TTL = historyCacheTTL

	return &historyCache{
		Cache:            cache.New(historyCacheMaxSize, opts),
		shard:            shard,
		executionManager: shard.GetExecutionManager(),
		lockTable:        make(map[string]*executionLock),
		pinnedContexts:   make(map[string]*workflowExecutionContext),
		logger: logger.WithFields(bark.Fields{
			tagWorkflowComponent: tagValueHistoryCacheComponent,
		}),
	}
}

func (c *historyCache) getOrCreateWorkflowExecution(domainID string,
	execution workflow.WorkflowExecution) (*workflowExecutionContext, releaseWorkflowExecutionFunc, error) {
	if execution.GetWorkflowId() == "" {
		return nil, nil, &workflow.InternalServiceError{Message: "Can't load workflow execution.  WorkflowId not set."}
	}

	// RunID is not provided, lets try to retrieve the RunID for current active execution
	if execution.GetRunId() == "" {
		response, err := c.getCurrentExecutionWithRetry(&persistence.GetCurrentExecutionRequest{
			DomainID:   domainID,
			WorkflowID: execution.GetWorkflowId(),
		})

		if err != nil {
			return nil, nil, err
		}

		execution.RunId = common.StringPtr(response.RunID)
	}

	executionLock := c.acquireExecutionLock(execution.GetRunId())

	releaseFunc := func() {
		c.releaseExecutionLock(execution.GetRunId(), executionLock)
	}

	// Test hook for disabling the cache
	if c.disabled {
		return newWorkflowExecutionContext(domainID, execution, c.shard, c.executionManager, c.logger), releaseFunc, nil
	}

	key := execution.GetRunId()
	context, cacheHit := c.Get(key).(*workflowExecutionContext)
	if !cacheHit {
		// First check if there is a pinned context for this execution
		context = c.getPinnedContext(execution.GetRunId())
		if context == nil {
			// Let's create the workflow execution context
			context = newWorkflowExecutionContext(domainID, execution, c.shard, c.executionManager, c.logger)
			context = c.PutIfNotExist(key, context).(*workflowExecutionContext)
		}
	}

	c.pinExecutionContext(execution.GetRunId(), context)
	return context, releaseFunc, nil
}

func (c *historyCache) getCurrentExecutionWithRetry(
	request *persistence.GetCurrentExecutionRequest) (*persistence.GetCurrentExecutionResponse, error) {
	var response *persistence.GetCurrentExecutionResponse
	op := func() error {
		var err error
		response, err = c.executionManager.GetCurrentExecution(request)

		return err
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *historyCache) acquireExecutionLock(runID string) *executionLock {
	var lock *executionLock
	var found bool

	c.Lock()
	if lock, found = c.lockTable[runID]; !found {
		lock = &executionLock{}
		c.lockTable[runID] = lock
	}
	lock.waitingCount++
	c.Unlock()

	lock.Lock()
	return lock
}

func (c *historyCache) releaseExecutionLock(runID string, lock *executionLock) {
	c.Lock()
	lock.waitingCount--
	if lock.waitingCount == 0 {
		delete(c.lockTable, runID)
		delete(c.pinnedContexts, runID)
	}
	c.Unlock()
	lock.Unlock()
}

func (c *historyCache) pinExecutionContext(runID string, context *workflowExecutionContext) {
	c.Lock()
	defer c.Unlock()
	if pinned, found := c.pinnedContexts[runID]; found {
		// sanity check, the context objects have to be the same
		if pinned != context {
			c.logger.Panicf("Found more than one workflow execution contexts for the same execution!")
		}
		return
	}
	c.pinnedContexts[runID] = context
}

func (c *historyCache) getPinnedContext(runID string) *workflowExecutionContext {
	c.Lock()
	defer c.Unlock()
	return c.pinnedContexts[runID]
}
