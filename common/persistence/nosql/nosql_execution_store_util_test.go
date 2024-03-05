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

package nosql

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

func TestNosqlExecutionStoreUtils(t *testing.T) {
	testCases := []struct {
		name       string
		setupStore func(*nosqlExecutionStore) (*nosqlplugin.WorkflowExecutionRequest, error)
		input      *persistence.InternalWorkflowSnapshot
		validate   func(*testing.T, *nosqlplugin.WorkflowExecutionRequest, error)
	}{
		{
			name: "PrepareCreateWorkflowExecutionRequestWithMaps - Success",
			setupStore: func(store *nosqlExecutionStore) (*nosqlplugin.WorkflowExecutionRequest, error) {
				workflowSnapshot := &persistence.InternalWorkflowSnapshot{
					ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					VersionHistories: &persistence.DataBlob{
						Encoding: common.EncodingTypeJSON,
						Data:     []byte(`[{"Branches":[{"BranchID":"test-branch-id","BeginNodeID":1,"EndNodeID":2}]}]`),
					},
				}
				return store.prepareCreateWorkflowExecutionRequestWithMaps(workflowSnapshot)
			},
			input: &persistence.InternalWorkflowSnapshot{},
			validate: func(t *testing.T, req *nosqlplugin.WorkflowExecutionRequest, err error) {
				assert.NoError(t, err)
				if err == nil {
					assert.NotNil(t, req)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockDB := nosqlplugin.NewMockDB(mockCtrl)
			store := newTestNosqlExecutionStore(mockDB, log.NewNoop())

			req, err := tc.setupStore(store)
			tc.validate(t, req, err)
		})
	}
}
