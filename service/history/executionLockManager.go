package history

import "sync"

type (
	executionLock struct {
		sync.Mutex
		waitCount int
	}

	executionLockManager struct {
		sync.Mutex
		lockTable map[string]*executionLock
	}
)

func newExecutionLockManager() *executionLockManager {
	return &executionLockManager{
		lockTable: make(map[string]*executionLock),
	}
}

func (e *executionLockManager) lockExecution(runID string) {
	var lock *executionLock
	var found bool

	e.Lock()
	if lock, found = e.lockTable[runID]; !found {
		lock = &executionLock{}
		e.lockTable[runID] = lock
	}
	lock.waitCount++
	e.Unlock()

	lock.Lock()
}

func (e *executionLockManager) unlockExecution(runID string) {
	var lock *executionLock

	e.Lock()
	lock = e.lockTable[runID]
	lock.waitCount--
	if lock.waitCount == 0 {
		delete(e.lockTable, runID)
	}
	e.Unlock()

	lock.Unlock()
}
