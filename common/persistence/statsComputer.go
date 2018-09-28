package persistence

type (
	// statsComputer is to computing struct sizes after serialization
	statsComputer struct{}
)

func (sc *statsComputer) computeMutableStateStats(req *PersistenceUpdateWorkflowExecutionRequest) *MutableStateStats {
	executionInfoSize := computeExecutionInfoSize(req.ExecutionInfo)

	activityInfoCount := 0
	activityInfoSize := 0
	for _, ai := range req.UpsertActivityInfos {
		activityInfoCount++
		activityInfoSize += computeActivityInfoSize(ai)
	}

	timerInfoCount := 0
	timerInfoSize := 0
	for _, ti := range req.UpserTimerInfos {
		timerInfoCount++
		timerInfoSize += computeTimerInfoSize(ti)
	}

	childExecutionInfoCount := 0
	childExecutionInfoSize := 0
	for _, ci := range req.UpsertChildExecutionInfos {
		childExecutionInfoCount++
		childExecutionInfoSize += computeChildInfoSize(ci)
	}

	signalInfoCount := 0
	signalInfoSize := 0
	for _, si := range req.UpsertSignalInfos {
		signalInfoCount++
		signalInfoSize += computeSignalInfoSize(si)
	}

	bufferedEventsCount := 0
	bufferedEventsSize := 0

	bufferedEventsCount++
	bufferedEventsSize += len(req.NewBufferedEvents.Data)

	bufferedReplicationTasksCount := 0
	bufferedReplicationTasksSize := 0
	bufferedReplicationTasksCount++
	bufferedReplicationTasksSize += computeBufferedReplicationTasksSize(req.NewBufferedReplicationTask)

	requestCancelInfoCount := len(req.UpsertRequestCancelInfos)

	totalSize := executionInfoSize
	totalSize += activityInfoSize
	totalSize += timerInfoSize
	totalSize += childExecutionInfoSize
	totalSize += signalInfoSize
	totalSize += bufferedEventsSize
	totalSize += bufferedReplicationTasksSize

	return &MutableStateStats{
		MutableStateSize:              totalSize,
		ExecutionInfoSize:             executionInfoSize,
		ActivityInfoSize:              activityInfoSize,
		TimerInfoSize:                 timerInfoSize,
		ChildInfoSize:                 childExecutionInfoSize,
		SignalInfoSize:                signalInfoSize,
		BufferedEventsSize:            bufferedEventsSize,
		BufferedReplicationTasksSize:  bufferedReplicationTasksSize,
		ActivityInfoCount:             activityInfoCount,
		TimerInfoCount:                timerInfoCount,
		ChildInfoCount:                childExecutionInfoCount,
		SignalInfoCount:               signalInfoCount,
		BufferedEventsCount:           bufferedEventsCount,
		BufferedReplicationTasksCount: bufferedReplicationTasksCount,
		RequestCancelInfoCount:        requestCancelInfoCount,
	}
}

func (sc *statsComputer) computeMutableStateUpdateStats(req *PersistenceUpdateWorkflowExecutionRequest) *MutableStateUpdateSessionStats {
	executionInfoSize := computeExecutionInfoSize(req.ExecutionInfo)

	activityInfoCount := 0
	activityInfoSize := 0
	for _, ai := range req.UpsertActivityInfos {
		activityInfoCount++
		activityInfoSize += computeActivityInfoSize(ai)
	}

	timerInfoCount := 0
	timerInfoSize := 0
	for _, ti := range req.UpserTimerInfos {
		timerInfoCount++
		timerInfoSize += computeTimerInfoSize(ti)
	}

	childExecutionInfoCount := 0
	childExecutionInfoSize := 0
	for _, ci := range req.UpsertChildExecutionInfos {
		childExecutionInfoCount++
		childExecutionInfoSize += computeChildInfoSize(ci)
	}

	signalInfoCount := 0
	signalInfoSize := 0
	for _, si := range req.UpsertSignalInfos {
		signalInfoCount++
		signalInfoSize += computeSignalInfoSize(si)
	}

	bufferedEventsSize := len(req.NewBufferedEvents.Data)

	bufferedReplicationTasksSize := computeBufferedReplicationTasksSize(req.NewBufferedReplicationTask)

	requestCancelInfoCount := len(req.UpsertRequestCancelInfos)

	deleteActivityInfoCount := len(req.DeleteActivityInfos)

	deleteTimerInfoCount := len(req.DeleteTimerInfos)

	deleteChildInfoCount := 0
	if req.DeleteChildExecutionInfo != nil {
		deleteChildInfoCount = 1
	}

	deleteSignalInfoCount := 0
	if req.DeleteSignalInfo != nil {
		deleteSignalInfoCount = 1
	}

	deleteRequestCancelInfoCount := 0
	if req.DeleteRequestCancelInfo != nil {
		deleteRequestCancelInfoCount = 1
	}

	totalSize := executionInfoSize
	totalSize += activityInfoSize
	totalSize += timerInfoSize
	totalSize += childExecutionInfoSize
	totalSize += signalInfoSize
	totalSize += bufferedEventsSize
	totalSize += bufferedReplicationTasksSize

	return &MutableStateUpdateSessionStats{
		MutableStateSize:             totalSize,
		ExecutionInfoSize:            executionInfoSize,
		ActivityInfoSize:             activityInfoSize,
		TimerInfoSize:                timerInfoSize,
		ChildInfoSize:                childExecutionInfoSize,
		SignalInfoSize:               signalInfoSize,
		BufferedEventsSize:           bufferedEventsSize,
		BufferedReplicationTasksSize: bufferedReplicationTasksSize,
		ActivityInfoCount:            activityInfoCount,
		TimerInfoCount:               timerInfoCount,
		ChildInfoCount:               childExecutionInfoCount,
		SignalInfoCount:              signalInfoCount,
		RequestCancelInfoCount:       requestCancelInfoCount,
		DeleteActivityInfoCount:      deleteActivityInfoCount,
		DeleteTimerInfoCount:         deleteTimerInfoCount,
		DeleteChildInfoCount:         deleteChildInfoCount,
		DeleteSignalInfoCount:        deleteSignalInfoCount,
		DeleteRequestCancelInfoCount: deleteRequestCancelInfoCount,
	}
}

func computeExecutionInfoSize(executionInfo *PersistenceWorkflowExecutionInfo) int {
	size := len(executionInfo.WorkflowID)
	size += len(executionInfo.TaskList)
	size += len(executionInfo.WorkflowTypeName)
	size += len(executionInfo.ParentWorkflowID)

	return size
}

func computeActivityInfoSize(ai *PersistenceActivityInfo) int {
	size := len(ai.ActivityID)
	size += len(ai.ScheduledEvent.Data)
	size += len(ai.StartedEvent.Data)
	size += len(ai.Details)

	return size
}

func computeTimerInfoSize(ti *TimerInfo) int {
	size := len(ti.TimerID)

	return size
}

func computeChildInfoSize(ci *PersistenceChildExecutionInfo) int {
	size := len(ci.InitiatedEvent.Data)
	size += len(ci.StartedEvent.Data)

	return size
}

func computeSignalInfoSize(si *SignalInfo) int {
	size := len(si.SignalName)
	size += len(si.Input)
	size += len(si.Control)

	return size
}

func computeBufferedReplicationTasksSize(task *PersistenceBufferedReplicationTask) int {
	size := 0

	if task != nil {
		if task.History != nil {
			size += len(task.History.Data)
		}

		if task.NewRunHistory != nil {
			size += len(task.NewRunHistory.Data)
		}
	}

	return size
}
