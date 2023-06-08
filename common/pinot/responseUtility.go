package pinot

import (
	"encoding/json"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"time"
)

func buildMap(hit []interface{}, columnNames []string, isSystemKey func(key string) (bool, string)) (map[string]interface{}, map[string]interface{}) {
	systemKeyMap := make(map[string]interface{})
	customKeyMap := make(map[string]interface{})

	for i := 0; i < len(columnNames); i++ {
		key := columnNames[i]
		// checks if it is system key, if yes, put it into the system map; otherwise put it into custom map
		ok, _ := isSystemKey(key)
		if ok {
			systemKeyMap[key] = hit[i]
		} else {
			customKeyMap[key] = hit[i]
		}
	}

	return systemKeyMap, customKeyMap
}

// VisibilityRecord is a struct of doc for deserialization
type VisibilityRecord struct {
	WorkflowID    string
	RunID         string
	WorkflowType  string
	DomainID      string
	StartTime     int64
	ExecutionTime int64
	CloseTime     int64
	CloseStatus   int
	HistoryLength int64
	Encoding      string
	TaskList      string
	IsCron        bool
	NumClusters   int16
	UpdateTime    int64
	Attr          string
}

func ConvertSearchResultToVisibilityRecord(hit []interface{}, columnNames []string, logger log.Logger, isSystemKey func(key string) (bool, string)) *p.InternalVisibilityWorkflowExecutionInfo {
	if len(hit) != len(columnNames) {
		return nil
	}

	systemKeyMap, customKeyMap := buildMap(hit, columnNames, isSystemKey)
	jsonSystemKeyMap, err := json.Marshal(systemKeyMap)
	if err != nil { // log and skip error
		logger.Error("unable to marshal systemKeyMap",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	var source *VisibilityRecord
	err = json.Unmarshal(jsonSystemKeyMap, &source)
	if err != nil { // log and skip error
		logger.Error("unable to Unmarshal systemKeyMap",
			tag.Error(err), //tag.ESDocID(fmt.Sprintf(columnNameToValue["DocID"]))
		)
		return nil
	}

	record := &p.InternalVisibilityWorkflowExecutionInfo{
		DomainID:         source.DomainID,
		WorkflowType:     source.WorkflowType,
		WorkflowID:       source.WorkflowID,
		RunID:            source.RunID,
		TypeName:         source.WorkflowType,
		StartTime:        time.UnixMilli(source.StartTime), // be careful: source.StartTime is in milisecond
		ExecutionTime:    time.UnixMilli(source.ExecutionTime),
		TaskList:         source.TaskList,
		IsCron:           source.IsCron,
		NumClusters:      source.NumClusters,
		SearchAttributes: customKeyMap,
	}
	if source.UpdateTime != 0 {
		record.UpdateTime = time.UnixMilli(source.UpdateTime)
	}
	if source.CloseTime != 0 {
		record.CloseTime = time.UnixMilli(source.CloseTime)
		record.Status = toWorkflowExecutionCloseStatus(source.CloseStatus)
		record.HistoryLength = source.HistoryLength
	}

	return record
}

func toWorkflowExecutionCloseStatus(status int) *types.WorkflowExecutionCloseStatus {
	if status < 0 {
		return nil
	}
	closeStatus := types.WorkflowExecutionCloseStatus(status)
	return &closeStatus
}
