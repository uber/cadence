package esql

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xwb1989/sqlparser"
)

// SetCadence ... specify whether do special handling for cadence visibility
// should not be called if there is potential race condition
// should not be called by non-cadence user
func (e *ESql) SetCadence(cadenceArg bool) {
	e.cadence = cadenceArg
}

// ConvertPrettyCadence ...
// convert sql to es dsl, for cadence usage
func (e *ESql) ConvertPrettyCadence(sql string, domainID string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	dsl, sortField, err = e.ConvertCadence(sql, domainID, pagination...)
	if err != nil {
		return "", nil, err
	}

	var prettifiedDSLBytes bytes.Buffer
	err = json.Indent(&prettifiedDSLBytes, []byte(dsl), "", "  ")
	if err != nil {
		return "", nil, err
	}
	return string(prettifiedDSLBytes.Bytes()), sortField, err
}

// ConvertCadence ...
// convert sql to es dsl, for cadence usage
func (e *ESql) ConvertCadence(sql string, domainID string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	if !e.cadence {
		err = fmt.Errorf(`esql: cadence option not turned on`)
		return "", nil, err
	}
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", nil, err
	}

	//sql valid, start to handle
	switch stmt.(type) {
	case *sqlparser.Select:
		dsl, sortField, err = e.convertSelect(*(stmt.(*sqlparser.Select)), domainID, pagination...)
	default:
		err = fmt.Errorf(`esql: Queries other than select not supported`)
	}

	if err != nil {
		return "", nil, err
	}
	return dsl, sortField, nil
}
