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
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (e *ESql) convertSelect(sel sqlparser.Select, domainID string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	if sel.Distinct != "" {
		err := fmt.Errorf(`esql: SELECT DISTINCT not supported. use GROUP BY instead`)
		return "", nil, err
	}

	var rootParent sqlparser.Expr
	// a map that contains the main components of a query
	dslMap := make(map[string]interface{})

	// handle WHERE keyword
	if sel.Where != nil {
		dslQuery, err := e.convertWhereExpr(sel.Where.Expr, rootParent)
		if err != nil {
			return "", nil, err
		}
		dslMap["query"] = dslQuery
	}
	// cadence special handling: add domain ID query
	if e.cadence {
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

	// handle FROM keyword, currently only support 1 target table
	if len(sel.From) != 1 {
		if len(sel.From) == 0 {
			err = fmt.Errorf("esql: invalid from expressino: no from expression specified")
		} else {
			err = fmt.Errorf("esql: join not supported")
		}
		return "", nil, err
	}

	// handle SELECT keyword
	_, selectedColNameSlice, aggNameSlice, err := e.extractSelectedExpr(sel.SelectExprs)
	if err != nil {
		return "", nil, err
	}
	if len(selectedColNameSlice) > 0 {
		colNames := `"` + strings.Join(selectedColNameSlice, `", "`) + `"`
		dslMap["_source"] = fmt.Sprintf(`{"includes": [%v]}`, colNames)
	}

	// handle all aggregations, including GROUP BY, SELECT <agg function>, ORDER BY <agg function>, HAVING
	dslAgg, err := e.convertAgg(sel)
	if err != nil {
		return "", nil, err
	}
	if dslAgg != "" || len(aggNameSlice) > 0 {
		if dslAgg != "" {
			dslMap["aggs"] = dslAgg
		}
		// do not return document contents if this is an aggregation query
		// dslMap["_source"] = "false"
		// dslMap["stored_fields"] = `"_none_"`
		dslMap["size"] = 0
	} else {
		// handle LIMIT and OFFSET keyword, these 2 keywords only works in non-aggregation query
		dslMap["size"] = e.pageSize
		if sel.Limit != nil {
			if sel.Limit.Offset != nil {
				dslMap["from"] = sqlparser.String(sel.Limit.Offset)
			}
			dslMap["size"] = sqlparser.String(sel.Limit.Rowcount)
		}
		// handle pagination
		var searchAfterSlice []string
		for _, v := range pagination {
			switch v.(type) {
			case int:
				searchAfterSlice = append(searchAfterSlice, fmt.Sprintf(`%v`, v))
			default:
				searchAfterSlice = append(searchAfterSlice, fmt.Sprintf(`"%v"`, v))
			}
		}
		if len(searchAfterSlice) > 0 {
			searchAfterStr := strings.Join(searchAfterSlice, ",")
			dslMap["search_after"] = fmt.Sprintf(`[%v]`, searchAfterStr)
		}
	}

	// handle ORDER BY <column name>
	// if it is an aggregate query, no point to order
	if _, exist := dslMap["aggs"]; !exist && len(aggNameSlice) == 0 {
		var orderBySlice []string
		for _, orderExpr := range sel.OrderBy {
			var colNameStr string
			if colName, ok := orderExpr.Expr.(*sqlparser.ColName); ok {
				colNameStr, err = e.convertColName(colName)
				if err != nil {
					return "", nil, err
				}
			} else {
				err := fmt.Errorf(`esql: mix order by aggregations and column names`)
				return "", nil, err
			}
			colNameStr = strings.Trim(colNameStr, "`")
			orderByStr := fmt.Sprintf(`{"%v": "%v"}`, colNameStr, orderExpr.Direction)
			orderBySlice = append(orderBySlice, orderByStr)
			sortField = append(sortField, colNameStr)
		}
		// cadence special handling: add runID as sorting tie breaker
		if e.cadence {
			switch len(orderBySlice) {
			case 0: // if unsorted, use default sorting
				cadenceOrderStartTime := fmt.Sprintf(`{"%v": "%v"}`, StartTime, StartTimeOrder)
				orderBySlice = append(orderBySlice, cadenceOrderStartTime)
				sortField = append(sortField, StartTime)
			case 1: // user should not use tieBreaker to sort
				if sortField[0] == TieBreaker {
					err = fmt.Errorf("esql: Cadence does not allow user sort by RunID")
					return "", nil, err
				}
			default:
				err = fmt.Errorf("esql: Cadence only allow 1 custom sort field")
				return "", nil, err
			}

			// add tie breaker
			cadenceOrderTieBreaker := fmt.Sprintf(`{"%v": "%v"}`, TieBreaker, TieBreakerOrder)
			orderBySlice = append(orderBySlice, cadenceOrderTieBreaker)
			sortField = append(sortField, TieBreaker)
		}
		if len(orderBySlice) > 0 {
			dslMap["sort"] = fmt.Sprintf("[%v]", strings.Join(orderBySlice, ","))
		}
	}

	// generate the final json query
	var dslQuerySlice []string
	for tag, content := range dslMap {
		dslQuerySlice = append(dslQuerySlice, fmt.Sprintf(`"%v": %v`, tag, content))
	}
	dsl = "{" + strings.Join(dslQuerySlice, ",") + "}"
	return dsl, sortField, nil
}

func (e *ESql) convertWhereExpr(expr sqlparser.Expr, parent sqlparser.Expr) (string, error) {
	var err error
	if expr == nil {
		err = fmt.Errorf("esql: invalid where expression, where expression should not be nil")
	}

	switch expr.(type) {
	case *sqlparser.ComparisonExpr:
		return e.convertComparisionExpr(expr, parent, false)
	case *sqlparser.AndExpr:
		return e.convertAndExpr(expr, parent)
	case *sqlparser.OrExpr:
		return e.convertOrExpr(expr, parent)
	case *sqlparser.ParenExpr:
		return e.convertParenExpr(expr, parent)
	case *sqlparser.NotExpr:
		return e.convertNotExpr(expr, parent)
	case *sqlparser.RangeCond:
		return e.convertBetweenExpr(expr, parent, true, true, false)
	case *sqlparser.IsExpr:
		return e.convertIsExpr(expr, parent, false)
	default:
		err = fmt.Errorf(`esql: %T expression not supported in WHERE clause`, expr)
		return "", err
	}
}

func (e *ESql) convertBetweenExpr(expr sqlparser.Expr, parent sqlparser.Expr, fromInclusive bool, toInclusive bool, not bool) (string, error) {
	rangeCond := expr.(*sqlparser.RangeCond)
	lhs, ok := rangeCond.Left.(*sqlparser.ColName)
	if !ok {
		err := fmt.Errorf("esql: invalid range column name")
		return "", err
	}
	lhsStr, err := e.convertColName(lhs)
	if err != nil {
		return "", err
	}

	fromStr := strings.Trim(sqlparser.String(rangeCond.From), `'`)
	toStr := strings.Trim(sqlparser.String(rangeCond.To), `'`)
	op := rangeCond.Operator
	if not {
		op = oppositeOperator[op]
	}

	gt := "gte"
	lt := "lte"
	if !fromInclusive {
		gt = "gt"
	}
	if !toInclusive {
		lt = "lt"
	}

	dsl := fmt.Sprintf(`{"range": {"%v": {"%v": "%v", "%v": "%v"}}}`, lhsStr, gt, fromStr, lt, toStr)
	if op == sqlparser.NotBetweenStr {
		dsl = fmt.Sprintf(`{"bool": {"must_not": [%v]}}`, dsl)
	}
	return dsl, nil
}

func (e *ESql) convertParenExpr(expr sqlparser.Expr, parent sqlparser.Expr) (string, error) {
	exprInside := expr.(*sqlparser.ParenExpr).Expr
	return e.convertWhereExpr(exprInside, expr)
}

// * dsl must_not is not an equivalent to sql NOT, should convert the inside expression accordingly
func (e *ESql) convertNotExpr(expr sqlparser.Expr, parent sqlparser.Expr) (string, error) {
	notExpr := expr.(*sqlparser.NotExpr)
	exprInside := notExpr.Expr
	switch (exprInside).(type) {
	case *sqlparser.NotExpr:
		expr1 := exprInside.(*sqlparser.NotExpr)
		expr2 := expr1.Expr
		return e.convertWhereExpr(expr2, parent)
	case *sqlparser.AndExpr:
		expr1 := exprInside.(*sqlparser.AndExpr)
		var exprLeft sqlparser.Expr = &sqlparser.NotExpr{Expr: expr1.Left}
		var exprRight sqlparser.Expr = &sqlparser.NotExpr{Expr: expr1.Right}
		var expr2 sqlparser.Expr = &sqlparser.OrExpr{Left: exprLeft, Right: exprRight}
		return e.convertOrExpr(expr2, parent)
	case *sqlparser.OrExpr:
		expr1 := exprInside.(*sqlparser.OrExpr)
		var exprLeft sqlparser.Expr = &sqlparser.NotExpr{Expr: expr1.Left}
		var exprRight sqlparser.Expr = &sqlparser.NotExpr{Expr: expr1.Right}
		var expr2 sqlparser.Expr = &sqlparser.AndExpr{Left: exprLeft, Right: exprRight}
		return e.convertAndExpr(expr2, parent)
	case *sqlparser.ParenExpr:
		expr1 := exprInside.(*sqlparser.ParenExpr)
		exprBody := expr1.Expr
		var expr2 sqlparser.Expr = &sqlparser.NotExpr{Expr: exprBody}
		return e.convertNotExpr(expr2, parent)
	case *sqlparser.ComparisonExpr:
		return e.convertComparisionExpr(exprInside, parent, true)
	case *sqlparser.IsExpr:
		return e.convertIsExpr(exprInside, parent, true)
	case *sqlparser.RangeCond:
		return e.convertBetweenExpr(exprInside, parent, true, true, true)
	default:
		err := fmt.Errorf("esql: %T expression not supported", exprInside)
		return "", err
	}
}

func (e *ESql) convertAndExpr(expr sqlparser.Expr, parent sqlparser.Expr) (string, error) {
	andExpr := expr.(*sqlparser.AndExpr)
	lhsExpr := andExpr.Left
	rhsExpr := andExpr.Right

	lhsStr, err := e.convertWhereExpr(lhsExpr, expr)
	if err != nil {
		return "", err
	}
	rhsStr, err := e.convertWhereExpr(rhsExpr, expr)
	if err != nil {
		return "", err
	}
	var dsl string
	if lhsStr == "" || rhsStr == "" {
		dsl = lhsStr + rhsStr
	} else {
		dsl = lhsStr + `,` + rhsStr
	}

	// merge chained AND expression
	if _, ok := parent.(*sqlparser.AndExpr); ok {
		return dsl, nil
	}
	return fmt.Sprintf(`{"bool": {"filter": [%v]}}`, dsl), nil
}

func (e *ESql) convertOrExpr(expr sqlparser.Expr, parent sqlparser.Expr) (string, error) {
	orExpr := expr.(*sqlparser.OrExpr)
	lhsExpr := orExpr.Left
	rhsExpr := orExpr.Right

	lhsStr, err := e.convertWhereExpr(lhsExpr, expr)
	if err != nil {
		return "", err
	}
	rhsStr, err := e.convertWhereExpr(rhsExpr, expr)
	if err != nil {
		return "", err
	}
	var dsl string
	if lhsStr == "" || rhsStr == "" {
		dsl = lhsStr + rhsStr
	} else {
		dsl = lhsStr + `,` + rhsStr
	}

	// merge chained OR expression
	if _, ok := parent.(*sqlparser.OrExpr); ok {
		return dsl, nil
	}
	return fmt.Sprintf(`{"bool": {"should": [%v]}}`, dsl), nil
}

func (e *ESql) convertIsExpr(expr sqlparser.Expr, parent sqlparser.Expr, not bool) (string, error) {
	isExpr := expr.(*sqlparser.IsExpr)
	lhs, ok := isExpr.Expr.(*sqlparser.ColName)
	if !ok {
		return "", fmt.Errorf("esql: is expression only support colname missing check")
	}
	lhsStr, err := e.convertColName(lhs)
	if err != nil {
		return "", err
	}

	dsl := ""
	op := isExpr.Operator
	if not {
		if _, exist := oppositeOperator[op]; !exist {
			err := fmt.Errorf("esql: is expression only support is null and is not null")
			return "", err
		}
		op = oppositeOperator[op]
	}
	switch op {
	case sqlparser.IsNullStr:
		dsl = fmt.Sprintf(`{"bool": {"must_not": {"exists": {"field": "%v"}}}}`, lhsStr)
	case sqlparser.IsNotNullStr:
		dsl = fmt.Sprintf(`{"exists": {"field": "%v"}}`, lhsStr)
	default:
		return "", fmt.Errorf("esql: is expression only support is null and is not null")
	}
	return dsl, nil
}

func (e *ESql) convertComparisionExpr(expr sqlparser.Expr, parent sqlparser.Expr, not bool) (string, error) {
	// extract lhs, and check lhs is a colName
	comparisonExpr := expr.(*sqlparser.ComparisonExpr)
	lhsExpr := comparisonExpr.Left
	lhs, ok := lhsExpr.(*sqlparser.ColName)
	if !ok {
		return "", fmt.Errorf("esql: invalid comparison expression, lhs must be a column name")
	}

	lhsStr, err := e.convertColName(lhs)
	if err != nil {
		return "", err
	}

	// extract rhs
	rhsExpr := comparisonExpr.Right
	rhsStr, err := e.convertValExpr(rhsExpr)
	if err != nil {
		return "", err
	}
	rhsStr, err = e.filterAndProcess(lhsStr, rhsStr)
	if err != nil {
		return "", err
	}

	op := comparisonExpr.Operator
	if not {
		if _, exist := oppositeOperator[op]; !exist {
			err := fmt.Errorf(`esql: %s operator not supported in comparison clause`, comparisonExpr.Operator)
			return "", err
		}
		op = oppositeOperator[op]
	}

	// generate dsl according to operator
	var dsl string
	switch op {
	case "=":
		dsl = fmt.Sprintf(`{"term": {"%v": "%v"}}`, lhsStr, rhsStr)
	case "<":
		dsl = fmt.Sprintf(`{"range": {"%v": {"lt": "%v"}}}`, lhsStr, rhsStr)
	case "<=":
		dsl = fmt.Sprintf(`{"range": {"%v": {"lte": "%v"}}}`, lhsStr, rhsStr)
	case ">":
		dsl = fmt.Sprintf(`{"range": {"%v": {"gt": "%v"}}}`, lhsStr, rhsStr)
	case ">=":
		dsl = fmt.Sprintf(`{"range": {"%v": {"gte": "%v"}}}`, lhsStr, rhsStr)
	case "<>", "!=":
		dsl = fmt.Sprintf(`{"bool": {"must_not": {"term": {"%v": "%v"}}}}`, lhsStr, rhsStr)
	case "in":
		rhsStr = strings.Replace(rhsStr, `'`, `"`, -1)
		rhsStr = strings.Trim(rhsStr, "(")
		rhsStr = strings.Trim(rhsStr, ")")
		dsl = fmt.Sprintf(`{"terms": {"%v": [%v]}}`, lhsStr, rhsStr)
	case "not in":
		rhsStr = strings.Replace(rhsStr, `'`, `"`, -1)
		rhsStr = strings.Trim(rhsStr, "(")
		rhsStr = strings.Trim(rhsStr, ")")
		dsl = fmt.Sprintf(`{"bool": {"must_not": {"terms": {"%v": [%v]}}}}`, lhsStr, rhsStr)
	case "like":
		rhsStr = strings.Replace(rhsStr, `_`, `?`, -1)
		rhsStr = strings.Replace(rhsStr, `%`, `*`, -1)
		dsl = fmt.Sprintf(`{"wildcard": {"%v": {"wildcard": "%v"}}}`, lhsStr, rhsStr)
	case "not like":
		rhsStr = strings.Replace(rhsStr, `_`, `?`, -1)
		rhsStr = strings.Replace(rhsStr, `%`, `*`, -1)
		dsl = fmt.Sprintf(`{"bool": {"must_not": {"wildcard": {"%v": {"wildcard": "%v"}}}}}`, lhsStr, rhsStr)
	case "regexp":
		dsl = fmt.Sprintf(`{"regexp": {"%v": "%v"}}`, lhsStr, rhsStr)
	case "not regexp":
		dsl = fmt.Sprintf(`{"bool": {"must_not": {"regexp": {"%v": "%v"}}}}`, lhsStr, rhsStr)
	default:
		err := fmt.Errorf(`esql: %s operator not supported in comparison clause`, comparisonExpr.Operator)
		return "", err
	}
	return dsl, nil
}

func (e *ESql) convertValExpr(expr sqlparser.Expr) (dsl string, err error) {
	switch expr.(type) {
	case *sqlparser.SQLVal:
		dsl = sqlparser.String(expr)
		dsl = strings.Trim(dsl, `'`)
	// ValTuple is not a pointer from sqlparser
	case sqlparser.ValTuple:
		dsl = sqlparser.String(expr)
	default:
		err = fmt.Errorf("esql: not supported rhs expression %T", expr)
		return "", err
	}
	return dsl, nil
}

func (e *ESql) convertColName(colName *sqlparser.ColName) (string, error) {
	// here we garuantee colName is of type *ColName
	colNameStr := sqlparser.String(colName)
	replacedColNameStr, err := e.filterAndReplace(colNameStr)
	if err != nil {
		return "", err
	}
	replacedColNameStr = strings.Replace(replacedColNameStr, "`", "", -1)
	return replacedColNameStr, nil
}

func (e *ESql) filterAndReplace(target string) (string, error) {
	if e.filterReplace != nil && e.filterReplace(target) && e.replace != nil {
		target, err := e.replace(target)
		if err != nil {
			return "", err
		}
		return target, nil
	}
	return target, nil
}

func (e *ESql) filterAndProcess(colName string, value string) (string, error) {
	if e.filterProcess != nil && e.filterProcess(colName) && e.process != nil {
		value, err := e.process(value)
		if err != nil {
			return "", err
		}
		return value, nil
	}
	return value, nil
}
