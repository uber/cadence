// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package esql

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xwb1989/sqlparser"
)

// ProcessFunc ...
// esql use replace function to apply user colName or column value replacing policy
type ProcessFunc func(string) (string, error)

// FilterFunc ...
// esql use filter function decide whether the policy will be applied to the column
// only accept column names that filter(colName) == true
type FilterFunc func(string) bool

// ESql ...
// ESql is used to hold necessary information that required in parsing
type ESql struct {
	filterKey    FilterFunc  // select the column we want to replace name
	filterValue  FilterFunc  // select the column we want to process value
	processKey   ProcessFunc // if selected by filterReplace, change the column name
	processValue ProcessFunc // if selected by filterProcess, change the column value
	cadence      bool
	pageSize     int
	bucketNumber int
}

// NewESql ... return a new default ESql
func NewESql() *ESql {
	return &ESql{
		pageSize:     DefaultPageSize,
		bucketNumber: DefaultBucketNumber,
		cadence:      false,
		processKey:   nil,
		processValue: nil,
		filterKey:    nil,
		filterValue:  nil,
	}
}

// ProcessQueryKey ... set up user specified column name replacement policy
// should not be called if there is potential race condition
func (e *ESql) ProcessQueryKey(filterArg FilterFunc, replaceArg ProcessFunc) {
	e.filterKey = filterArg
	e.processKey = replaceArg
}

// ProcessQueryValue ... set up user specified column value processing policy
// should not be called if there is potential race condition
func (e *ESql) ProcessQueryValue(filterArg FilterFunc, processArg ProcessFunc) {
	e.filterValue = filterArg
	e.processValue = processArg
}

// SetPageSize ... set the number of documents returned in a non-aggregation query
// should not be called if there is potential race condition
func (e *ESql) SetPageSize(pageSizeArg int) {
	e.pageSize = pageSizeArg
}

// SetBucketNum ... set the number of bucket returned in an aggregation query
// should not be called if there is potential race condition
func (e *ESql) SetBucketNum(bucketNumArg int) {
	e.bucketNumber = bucketNumArg
}

// ConvertPretty ...
// Transform sql to elasticsearch dsl, and prettify the output json
//
// usage:
//  - dsl, sortField, err := e.ConvertPretty(sql, pageParam1, pageParam2, ...)
//
// arguments:
//  - sql: the sql query needs conversion in string format
//  - pagination: variadic arguments that indicates es search_after for pagination
//
// return values:
//  - dsl: the elasticsearch dsl json style string
//  - sortField: string array that contains all column names used for sorting. useful for pagination.
//  - err: contains err information
func (e *ESql) ConvertPretty(sql string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	dsl, sortField, err = e.Convert(sql, pagination...)
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

// Convert ...
// Transform sql to elasticsearch dsl string
//
// usage:
//  - dsl, sortField, err := e.Convert(sql, pageParam1, pageParam2, ...)
//
// arguments:
//  - sql: the sql query needs conversion in string format
//  - pagination: variadic arguments that indicates es search_after for
//
// return values:
//	- dsl: the elasticsearch dsl json style string
//	- sortField: string array that contains all column names used for sorting. useful for pagination.
//  - err: contains err information
func (e *ESql) Convert(sql string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", nil, err
	}

	//sql valid, start to handle
	switch stmt.(type) {
	case *sqlparser.Select:
		dsl, sortField, err = e.convertSelect(*(stmt.(*sqlparser.Select)), "", pagination...)
	default:
		err = fmt.Errorf(`esql: Queries other than select not supported`)
	}

	if err != nil {
		return "", nil, err
	}
	return dsl, sortField, nil
}
