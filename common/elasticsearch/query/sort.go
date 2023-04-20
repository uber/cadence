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

package query

type Sorter interface {
	Source() (interface{}, error)
}

// FieldSort sorts by a given field.
type FieldSort struct {
	Sorter
	fieldName string
	ascending bool
}

// NewFieldSort creates a new FieldSort.
func NewFieldSort(fieldName string) *FieldSort {
	return &FieldSort{
		fieldName: fieldName,
		ascending: true,
	}
}

// FieldName specifies the name of the field to be used for sorting.
func (s *FieldSort) FieldName(fieldName string) *FieldSort {
	s.fieldName = fieldName
	return s
}

// Order defines whether sorting ascending (default) or descending.
func (s *FieldSort) Order(ascending bool) *FieldSort {
	s.ascending = ascending
	return s
}

// Asc sets ascending sort order.
func (s *FieldSort) Asc() *FieldSort {
	s.ascending = true
	return s
}

// Desc sets descending sort order.
func (s *FieldSort) Desc() *FieldSort {
	s.ascending = false
	return s
}

// Source returns the JSON-serializable data.
func (s *FieldSort) Source() (interface{}, error) {
	source := make(map[string]interface{})
	x := make(map[string]interface{})
	source[s.fieldName] = x
	if s.ascending {
		x["order"] = "asc"
	} else {
		x["order"] = "desc"
	}

	return source, nil
}
