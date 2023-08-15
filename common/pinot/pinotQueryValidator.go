// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pinot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/uber/cadence/common/log"

	"github.com/xwb1989/sqlparser"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/types"
)

// VisibilityQueryValidator for sql query validation
type VisibilityQueryValidator struct {
	validSearchAttributes map[string]interface{}
}

// NewPinotQueryValidator create VisibilityQueryValidator
func NewPinotQueryValidator(validSearchAttributes map[string]interface{}) *VisibilityQueryValidator {
	return &VisibilityQueryValidator{
		validSearchAttributes: validSearchAttributes,
	}
}

// ValidateQuery validates that search attributes in the query and returns modified query.
func (qv *VisibilityQueryValidator) ValidateQuery(whereClause string) (string, error) {
	if len(whereClause) != 0 {
		// Build a placeholder query that allows us to easily parse the contents of the where clause.
		// IMPORTANT: This query is never executed, it is just used to parse and validate whereClause
		var placeholderQuery string
		whereClause := strings.TrimSpace(whereClause)
		if common.IsJustOrderByClause(whereClause) { // just order by
			placeholderQuery = fmt.Sprintf("SELECT * FROM dummy %s", whereClause)
		} else {
			placeholderQuery = fmt.Sprintf("SELECT * FROM dummy WHERE %s", whereClause)
		}

		stmt, err := sqlparser.Parse(placeholderQuery)
		if err != nil {
			return "", &types.BadRequestError{Message: "Invalid query."}
		}

		sel, ok := stmt.(*sqlparser.Select)
		if !ok {
			return "", &types.BadRequestError{Message: "Invalid select query."}
		}
		buf := sqlparser.NewTrackedBuffer(nil)
		// validate where expr
		if sel.Where != nil {
			err = qv.validateWhereExpr(sel.Where.Expr)
			if err != nil {
				return "", &types.BadRequestError{Message: err.Error()}
			}
			sel.Where.Expr.Format(buf)
		}

		sel.OrderBy.Format(buf)

		return buf.String(), nil
	}
	return whereClause, nil
}

func (qv *VisibilityQueryValidator) validateWhereExpr(expr sqlparser.Expr) error {
	if expr == nil {
		return nil
	}

	switch expr := expr.(type) {
	case *sqlparser.AndExpr, *sqlparser.OrExpr:
		return qv.validateAndOrExpr(expr)
	case *sqlparser.ComparisonExpr:
		return qv.validateComparisonExpr(expr)
	case *sqlparser.RangeCond:
		return nil
		//return qv.validateRangeExpr(expr)
	case *sqlparser.ParenExpr:
		return qv.validateWhereExpr(expr.Expr)
	default:
		return errors.New("invalid where clause")
	}

}

func (qv *VisibilityQueryValidator) validateAndOrExpr(expr sqlparser.Expr) error {
	var leftExpr sqlparser.Expr
	var rightExpr sqlparser.Expr

	switch expr := expr.(type) {
	case *sqlparser.AndExpr:
		leftExpr = expr.Left
		rightExpr = expr.Right
	case *sqlparser.OrExpr:
		leftExpr = expr.Left
		rightExpr = expr.Right
	}

	if err := qv.validateWhereExpr(leftExpr); err != nil {
		return err
	}
	return qv.validateWhereExpr(rightExpr)
}

func (qv *VisibilityQueryValidator) validateComparisonExpr(expr sqlparser.Expr) error {
	comparisonExpr := expr.(*sqlparser.ComparisonExpr)
	colName, ok := comparisonExpr.Left.(*sqlparser.ColName)
	if !ok {
		return errors.New("invalid comparison expression, left")
	}

	colNameStr := colName.Name.String()

	if !qv.isValidSearchAttributes(colNameStr) {
		return fmt.Errorf("invalid search attribute %q", colNameStr)
	}

	if definition.IsSystemIndexedKey(colNameStr) {
		if comparisonExpr.Operator != sqlparser.EqualStr {
			return nil
		}
		// need to deal with missing value
		// Question: why is the right side of a system attribute a colname, custom attribute is a SQLVal?
		colVal, ok := comparisonExpr.Right.(*sqlparser.ColName)
		if !ok {
			return nil
		}
		colValStr := colVal.Name.String()
		if colValStr == "missing" {
			var newColVal string
			if strings.ToLower(colNameStr) == "historylength" {
				newColVal = "0"
			} else {
				newColVal = "-1" // -1 is the default value for all Closed workflows related fields
			}
			comparisonExpr.Right = &sqlparser.ColName{
				Metadata:  colName.Metadata,
				Name:      sqlparser.NewColIdent(newColVal),
				Qualifier: colName.Qualifier,
			}
		}

	} else { // check type: if is IndexedValueTypeString, change to like statement for partial match
		if valType, ok := qv.validSearchAttributes[colNameStr]; ok {
			indexValType := common.ConvertIndexedValueTypeToInternalType(valType, log.NewNoop())

			if indexValType == types.IndexedValueTypeString {
				colVal, ok := comparisonExpr.Right.(*sqlparser.SQLVal)
				if !ok {
					return errors.New("invalid comparison expression, right")
				}
				colValStr := string(colVal.Val)

				// change to like statement for partial match
				comparisonExpr.Operator = sqlparser.LikeStr
				comparisonExpr.Right = &sqlparser.SQLVal{
					Type: sqlparser.StrVal,
					Val:  []byte("%" + colValStr + "%"),
				}
			}
		}
	}

	return nil
}

// isValidSearchAttributes return true if key is registered
func (qv *VisibilityQueryValidator) isValidSearchAttributes(key string) bool {
	validAttr := qv.validSearchAttributes
	_, isValidKey := validAttr[key]
	return isValidKey
}
