package logging

import "github.com/uber-common/bark"

// LogPersistantStoreErrorEvent is used to log errors from persistence layer.
func LogPersistantStoreErrorEvent(logger bark.Logger, operation string, err error, details string) {
	logger.WithFields(bark.Fields{
		TagWorkflowEventID: PersistentStoreErrorEventID,
		TagStoreOperation:  operation,
		TagWorkflowErr:     err,
	}).Errorf("Persistent store operation failure. Operation Details: %v", details)
}

// LogOperationFailedEvent is used to log generic operation failures.
func LogOperationFailedEvent(logger bark.Logger, msg string, err error) {
	logger.WithFields(bark.Fields{
		TagWorkflowEventID: OperationFailed,
	}).Warnf("%v.  Error: %v", msg, err)
}
