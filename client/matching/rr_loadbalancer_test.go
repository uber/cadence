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

package matching

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/types"
)

func TestNewRoundRobinLoadBalancer(t *testing.T) {
	domainIDToName := func(domainID string) (string, error) {
		return "testDomainName", nil
	}
	dc := dynamicconfig.NewCollection(dynamicconfig.NewNopClient(), testlogger.New(t))

	lb := NewRoundRobinLoadBalancer(domainIDToName, dc)
	assert.NotNil(t, lb)
	rb, ok := lb.(*roundRobinLoadBalancer)
	assert.NotNil(t, rb)
	assert.True(t, ok)

	assert.NotNil(t, rb.domainIDToName)
	assert.NotNil(t, rb.nReadPartitions)
	assert.NotNil(t, rb.nWritePartitions)
	assert.NotNil(t, rb.readCache)
	assert.NotNil(t, rb.writeCache)
}

func TestPickPartition(t *testing.T) {
	tests := []struct {
		name           string
		domainID       string
		taskList       types.TaskList
		taskListType   int
		forwardedFrom  string
		nPartitions    int
		setupCache     func(mockCache *cache.MockCache)
		expectedResult string
	}{
		{
			name:           "ForwardedFrom is not empty",
			domainID:       "testDomain",
			taskList:       types.TaskList{Name: "testTaskList", Kind: types.TaskListKindNormal.Ptr()},
			taskListType:   1,
			forwardedFrom:  "otherDomain",
			nPartitions:    3,
			setupCache:     nil,
			expectedResult: "testTaskList",
		},
		{
			name:           "Sticky task list",
			domainID:       "testDomain",
			taskList:       types.TaskList{Name: "testTaskList", Kind: types.TaskListKindSticky.Ptr()},
			taskListType:   1,
			forwardedFrom:  "",
			nPartitions:    3,
			setupCache:     nil,
			expectedResult: "testTaskList",
		},
		{
			name:           "Reserved task list prefix",
			domainID:       "testDomain",
			taskList:       types.TaskList{Name: fmt.Sprintf("%vTest", common.ReservedTaskListPrefix), Kind: types.TaskListKindNormal.Ptr()},
			taskListType:   1,
			forwardedFrom:  "",
			nPartitions:    3,
			setupCache:     nil,
			expectedResult: fmt.Sprintf("%vTest", common.ReservedTaskListPrefix),
		},
		{
			name:           "nPartitions <= 1",
			domainID:       "testDomain",
			taskList:       types.TaskList{Name: "testTaskList", Kind: types.TaskListKindNormal.Ptr()},
			taskListType:   1,
			forwardedFrom:  "",
			nPartitions:    1,
			setupCache:     nil,
			expectedResult: "testTaskList",
		},
		{
			name:          "Cache miss and partitioned task list",
			domainID:      "testDomain",
			taskList:      types.TaskList{Name: "testTaskList", Kind: types.TaskListKindNormal.Ptr()},
			taskListType:  1,
			forwardedFrom: "",
			nPartitions:   3,
			setupCache: func(mockCache *cache.MockCache) {
				mockCache.EXPECT().Get(key{
					domainID:     "testDomain",
					taskListName: "testTaskList",
					taskListType: 1,
				}).Return(nil)
				mockCache.EXPECT().PutIfNotExist(key{
					domainID:     "testDomain",
					taskListName: "testTaskList",
					taskListType: 1,
				}, gomock.Any()).DoAndReturn(func(key key, val interface{}) (interface{}, error) {
					if *val.(*int64) != -1 {
						panic("Expected value to be -1")
					}
					return val, nil
				})
			},
			expectedResult: "testTaskList",
		},
		{
			name:          "Cache error and partitioned task list",
			domainID:      "testDomain",
			taskList:      types.TaskList{Name: "testTaskList", Kind: types.TaskListKindNormal.Ptr()},
			taskListType:  1,
			forwardedFrom: "",
			nPartitions:   3,
			setupCache: func(mockCache *cache.MockCache) {
				mockCache.EXPECT().Get(key{
					domainID:     "testDomain",
					taskListName: "testTaskList",
					taskListType: 1,
				}).Return(nil)
				mockCache.EXPECT().PutIfNotExist(key{
					domainID:     "testDomain",
					taskListName: "testTaskList",
					taskListType: 1,
				}, gomock.Any()).Return(nil, fmt.Errorf("cache error"))
			},
			expectedResult: "testTaskList",
		},
		{
			name:          "Cache hit and partitioned task list",
			domainID:      "testDomain",
			taskList:      types.TaskList{Name: "testTaskList", Kind: types.TaskListKindNormal.Ptr()},
			taskListType:  1,
			forwardedFrom: "",
			nPartitions:   3,
			setupCache: func(mockCache *cache.MockCache) {
				mockCache.EXPECT().Get(key{
					domainID:     "testDomain",
					taskListName: "testTaskList",
					taskListType: 1,
				}).Return(new(int64))
			},
			expectedResult: "/__cadence_sys/testTaskList/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCache := cache.NewMockCache(ctrl)

			// If the test requires setting up cache behavior, call setupCache
			if tt.setupCache != nil {
				tt.setupCache(mockCache)
			}

			// Call the pickPartition function
			result := pickPartition(
				tt.domainID,
				tt.taskList,
				tt.taskListType,
				tt.forwardedFrom,
				tt.nPartitions,
				mockCache,
			)

			// Assert that the result matches the expected result
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestPickWritePartition(t *testing.T) {
	tests := []struct {
		name              string
		domainID          string
		taskList          types.TaskList
		taskListType      int
		forwardedFrom     string
		domainIDToName    func(string, *testing.T) (string, error)
		nReadPartitions   func(string, string, int, *testing.T) int
		nWritePartitions  func(string, string, int, *testing.T) int
		pickPartitionFn   func(string, types.TaskList, int, string, int, cache.Cache, *testing.T) string
		expectedPartition string
		expectError       bool
	}{
		{
			name:          "successful partition pick",
			domainID:      "testDomainID",
			taskList:      types.TaskList{Name: "testTaskList"},
			taskListType:  1,
			forwardedFrom: "",
			domainIDToName: func(domainID string, t *testing.T) (string, error) {
				assert.Equal(t, "testDomainID", domainID) // Assert parameter with t
				return "testDomainName", nil
			},
			nReadPartitions: func(domainName, taskListName string, taskListType int, t *testing.T) int {
				assert.Equal(t, "testDomainName", domainName) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskListName)
				assert.Equal(t, 1, taskListType)
				return 3
			},
			nWritePartitions: func(domainName, taskListName string, taskListType int, t *testing.T) int {
				assert.Equal(t, "testDomainName", domainName) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskListName)
				assert.Equal(t, 1, taskListType)
				return 4
			},
			pickPartitionFn: func(domainID string, taskList types.TaskList, taskListType int, forwardedFrom string, nPartitions int, partitionCache cache.Cache, t *testing.T) string {
				assert.Equal(t, "testDomainID", domainID) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskList.GetName())
				assert.Equal(t, 1, taskListType)
				assert.Equal(t, "", forwardedFrom)
				assert.Equal(t, 3, nPartitions)
				return "partition1"
			},
			expectedPartition: "partition1",
			expectError:       false,
		},
		{
			name:          "domainIDToName returns error",
			domainID:      "badDomainID",
			taskList:      types.TaskList{Name: "testTaskList"},
			taskListType:  1,
			forwardedFrom: "",
			domainIDToName: func(domainID string, t *testing.T) (string, error) {
				assert.Equal(t, "badDomainID", domainID) // Assert parameter with t
				return "", fmt.Errorf("domain not found")
			},
			expectedPartition: "testTaskList",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &roundRobinLoadBalancer{
				domainIDToName: func(domainID string) (string, error) {
					return tt.domainIDToName(domainID, t)
				},
				nReadPartitions: func(domainName, taskListName string, taskListType int) int {
					return tt.nReadPartitions(domainName, taskListName, taskListType, t)
				},
				nWritePartitions: func(domainName, taskListName string, taskListType int) int {
					return tt.nWritePartitions(domainName, taskListName, taskListType, t)
				},
				pickPartitionFn: func(domainName string, taskList types.TaskList, taskListType int, forwardedFrom string, nPartitions int, partitionCache cache.Cache) string {
					return tt.pickPartitionFn(domainName, taskList, taskListType, forwardedFrom, nPartitions, partitionCache, t)
				},
				writeCache: cache.New(&cache.Options{
					TTL:             0,
					InitialCapacity: 100,
					Pin:             false,
					MaxCount:        3000,
					ActivelyEvict:   false,
				}),
			}

			partition := lb.PickWritePartition(tt.domainID, tt.taskList, tt.taskListType, tt.forwardedFrom)
			assert.Equal(t, tt.expectedPartition, partition)
		})
	}
}

func TestPickReadPartition(t *testing.T) {
	tests := []struct {
		name              string
		domainID          string
		taskList          types.TaskList
		taskListType      int
		forwardedFrom     string
		domainIDToName    func(string, *testing.T) (string, error)
		nReadPartitions   func(string, string, int, *testing.T) int
		nWritePartitions  func(string, string, int, *testing.T) int
		pickPartitionFn   func(string, types.TaskList, int, string, int, cache.Cache, *testing.T) string
		expectedPartition string
		expectError       bool
	}{
		{
			name:          "successful partition pick",
			domainID:      "testDomainID",
			taskList:      types.TaskList{Name: "testTaskList"},
			taskListType:  1,
			forwardedFrom: "",
			domainIDToName: func(domainID string, t *testing.T) (string, error) {
				assert.Equal(t, "testDomainID", domainID) // Assert parameter with t
				return "testDomainName", nil
			},
			nReadPartitions: func(domainName, taskListName string, taskListType int, t *testing.T) int {
				assert.Equal(t, "testDomainName", domainName) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskListName)
				assert.Equal(t, 1, taskListType)
				return 3
			},
			nWritePartitions: func(domainName, taskListName string, taskListType int, t *testing.T) int {
				assert.Equal(t, "testDomainName", domainName) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskListName)
				assert.Equal(t, 1, taskListType)
				return 4
			},
			pickPartitionFn: func(domainID string, taskList types.TaskList, taskListType int, forwardedFrom string, nPartitions int, partitionCache cache.Cache, t *testing.T) string {
				assert.Equal(t, "testDomainID", domainID) // Assert parameters with t
				assert.Equal(t, "testTaskList", taskList.GetName())
				assert.Equal(t, 1, taskListType)
				assert.Equal(t, "", forwardedFrom)
				assert.Equal(t, 3, nPartitions)
				return "partition1"
			},
			expectedPartition: "partition1",
			expectError:       false,
		},
		{
			name:          "domainIDToName returns error",
			domainID:      "badDomainID",
			taskList:      types.TaskList{Name: "testTaskList"},
			taskListType:  1,
			forwardedFrom: "",
			domainIDToName: func(domainID string, t *testing.T) (string, error) {
				assert.Equal(t, "badDomainID", domainID) // Assert parameter with t
				return "", fmt.Errorf("domain not found")
			},
			expectedPartition: "testTaskList",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &roundRobinLoadBalancer{
				domainIDToName: func(domainID string) (string, error) {
					return tt.domainIDToName(domainID, t)
				},
				nReadPartitions: func(domainName, taskListName string, taskListType int) int {
					return tt.nReadPartitions(domainName, taskListName, taskListType, t)
				},
				nWritePartitions: func(domainName, taskListName string, taskListType int) int {
					return tt.nWritePartitions(domainName, taskListName, taskListType, t)
				},
				pickPartitionFn: func(domainName string, taskList types.TaskList, taskListType int, forwardedFrom string, nPartitions int, partitionCache cache.Cache) string {
					return tt.pickPartitionFn(domainName, taskList, taskListType, forwardedFrom, nPartitions, partitionCache, t)
				},
				writeCache: cache.New(&cache.Options{
					TTL:             0,
					InitialCapacity: 100,
					Pin:             false,
					MaxCount:        3000,
					ActivelyEvict:   false,
				}),
			}

			partition := lb.PickReadPartition(tt.domainID, tt.taskList, tt.taskListType, tt.forwardedFrom)
			assert.Equal(t, tt.expectedPartition, partition)
		})
	}
}
