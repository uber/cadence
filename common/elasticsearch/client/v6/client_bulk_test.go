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

package v6

import (
	"errors"
	"testing"

	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/elasticsearch/bulk"
)

func TestConvertV6ErrorToGenericError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected *bulk.GenericError
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
		{
			name: "non-elasticsearch error",
			err:  errors.New("generic error"),
			expected: &bulk.GenericError{
				Status:  bulk.UnknownStatusCode,
				Details: errors.New("generic error"),
			},
		},
		{
			name: "elasticsearch error",
			err: &elastic.Error{
				Status: 404,
				Details: &elastic.ErrorDetails{
					Type:   "index_not_found_exception",
					Reason: "no such index",
				},
			},
			expected: &bulk.GenericError{
				Status: 404,
				Details: &elastic.Error{
					Status: 404,
					Details: &elastic.ErrorDetails{
						Type:   "index_not_found_exception",
						Reason: "no such index",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertV6ErrorToGenericError(tt.err)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Status, result.Status)

				if expectedDetails, ok := tt.expected.Details.(*elastic.Error); ok {
					resultDetails, _ := result.Details.(*elastic.Error)
					assert.NotNil(t, resultDetails)
					assert.Equal(t, expectedDetails.Status, resultDetails.Status)
					assert.Equal(t, expectedDetails.Details.Type, resultDetails.Details.Type)
					assert.Equal(t, expectedDetails.Details.Reason, resultDetails.Details.Reason)
				} else {
					assert.Equal(t, tt.expected.Details.Error(), result.Details.Error())
				}
			}
		})
	}
}

func TestFromV6toGenericBulkResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       *elastic.BulkResponse
		expectNil      bool
		expectedTook   int
		expectedErrors bool
	}{
		{
			name:           "nil response",
			response:       nil,
			expectNil:      false,
			expectedTook:   0,
			expectedErrors: false,
		},
		{
			name: "non-nil response",
			response: &elastic.BulkResponse{
				Took:   100,
				Errors: true,
				Items: []map[string]*elastic.BulkResponseItem{
					{
						"index": &elastic.BulkResponseItem{
							Status: 200,
							Error:  nil,
						},
					},
				},
			},
			expectNil:      false,
			expectedTook:   100,
			expectedErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromV6ToGenericBulkResponse(tt.response)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTook, result.Took)
				assert.Equal(t, tt.expectedErrors, result.Errors)
			}
		})
	}
}

func TestFromV6ToGenericBulkResponseItemMaps(t *testing.T) {
	tests := []struct {
		name     string
		v6items  []map[string]*elastic.BulkResponseItem
		expected []map[string]*bulk.GenericBulkResponseItem
	}{
		{
			name: "normal case",
			v6items: []map[string]*elastic.BulkResponseItem{
				{
					"index": &elastic.BulkResponseItem{
						Status: 200,
					},
				},
				{
					"update": &elastic.BulkResponseItem{
						Status: 404,
						Error: &elastic.ErrorDetails{
							Type:   "index_not_found_exception",
							Reason: "no such index",
						},
					},
				},
			},
			expected: []map[string]*bulk.GenericBulkResponseItem{
				{
					"index": {
						Status: 200,
					},
				},
				{
					"update": {
						Status: 404,
					},
				},
			},
		},
		{
			name:     "nil case",
			v6items:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genericItems := fromV6ToGenericBulkResponseItemMaps(tt.v6items)
			assert.Len(t, genericItems, len(tt.expected), "The lengths of actual and expected slices should match.")
			for i, expectedMap := range tt.expected {
				actualMap := genericItems[i]

				for key, expectedItem := range expectedMap {
					actualItem, exists := actualMap[key]
					assert.True(t, exists, "Key should exist in actual map: "+key)

					if expectedItem != nil && actualItem != nil {
						assert.Equal(t, expectedItem.Status, actualItem.Status, "Status should match for key: "+key)
					} else {
						assert.Equal(t, expectedItem, actualItem, "Both expected and actual items should be nil for key: "+key)
					}
				}
			}
		})
	}
}

func TestFromV6ToGenericBulkResponseItemMap(t *testing.T) {
	tests := []struct {
		name     string
		v6item   map[string]*elastic.BulkResponseItem
		expected map[string]*bulk.GenericBulkResponseItem
	}{
		{
			name:     "nil case",
			v6item:   nil,
			expected: nil,
		},
		{
			name: "normal case",
			v6item: map[string]*elastic.BulkResponseItem{
				"index": {
					Index:   "test_index",
					Type:    "test_type",
					Id:      "1",
					Version: 1,
					Result:  "created",
					Status:  201,
				},
			},
			expected: map[string]*bulk.GenericBulkResponseItem{
				"index": {
					Index:   "test_index",
					Type:    "test_type",
					ID:      "1",
					Version: 1,
					Result:  "created",
					Status:  201,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genericItems := fromV6ToGenericBulkResponseItemMap(tt.v6item)
			assert.Equal(t, tt.expected, genericItems)
		})
	}
}

func TestFromV6ToGenericBulkResponseItem(t *testing.T) {
	elasticItem := &elastic.BulkResponseItem{
		Index:   "test_index",
		Type:    "test_type",
		Id:      "1",
		Version: 1,
		Result:  "created",
		Status:  201,
	}
	expectedGenericItem := &bulk.GenericBulkResponseItem{
		Index:   "test_index",
		Type:    "test_type",
		ID:      "1",
		Version: 1,
		Result:  "created",
		Status:  201,
	}

	result := fromV6ToGenericBulkResponseItem(elasticItem)
	assert.Equal(t, expectedGenericItem, result)
}

func TestFromV6ToGenericBulkableRequests(t *testing.T) {
	mockRequests := []elastic.BulkableRequest{
		elastic.NewBulkIndexRequest().Index("index1").Type("_doc").Id("1").Doc(map[string]interface{}{"field": "value1"}),
		elastic.NewBulkIndexRequest().Index("index2").Type("_doc").Id("2").Doc(map[string]interface{}{"field": "value2"}),
	}

	genericRequests := fromV6ToGenericBulkableRequests(mockRequests)
	assert.Len(t, genericRequests, len(mockRequests))
}
