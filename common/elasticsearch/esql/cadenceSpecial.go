package esql

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (e *ESql) addCadenceSort(orderBySlice []string, sortFields []string) ([]string, []string, error) {
	switch len(orderBySlice) {
	case 0: // if unsorted, use default sorting
		cadenceOrderStartTime := fmt.Sprintf(`{"%v": "%v"}`, StartTime, StartTimeOrder)
		orderBySlice = append(orderBySlice, cadenceOrderStartTime)
		sortFields = append(sortFields, StartTime)
	case 1: // user should not use tieBreaker to sort
		if sortFields[0] == TieBreaker {
			err := fmt.Errorf("esql: Cadence does not allow user sort by RunID")
			return nil, nil, err
		}
	default:
		err := fmt.Errorf("esql: Cadence only allow 1 custom sort field")
		return nil, nil, err
	}

	// add tie breaker
	cadenceOrderTieBreaker := fmt.Sprintf(`{"%v": "%v"}`, TieBreaker, TieBreakerOrder)
	orderBySlice = append(orderBySlice, cadenceOrderTieBreaker)
	sortFields = append(sortFields, TieBreaker)
	return orderBySlice, sortFields, nil
}

func (e *ESql) addCadenceDomainTimeQuery(sel sqlparser.Select, domainID string, dslMap map[string]interface{}) {
	var domainIDQuery string
	if domainID != "" {
		domainIDQuery = fmt.Sprintf(`{"term": {"%v": "%v"}}`, DomainID, domainID)
	}
	if sel.Where == nil {
		if domainID != "" {
			dslMap["query"] = domainIDQuery
		}
	} else {
		if domainID != "" {
			domainIDQuery = domainIDQuery + ","
		}
		if strings.Contains(fmt.Sprintf("%v", dslMap["query"]), ExecutionTime) {
			executionTimeBound := fmt.Sprintf(`{"range": {"%v": {"gte": "0"}}}`, ExecutionTime)
			dslMap["query"] = fmt.Sprintf(`{"bool": {"filter": [%v %v, %v]}}`, domainIDQuery, executionTimeBound, dslMap["query"])
		} else {
			dslMap["query"] = fmt.Sprintf(`{"bool": {"filter": [%v %v]}}`, domainIDQuery, dslMap["query"])
		}
	}
}
