// Copyright (c) 2020 Uber Technologies, Inc.
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

package cassandra

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/testdata"
)

func TestInsertVisibility(t *testing.T) {
	tests := []struct {
		desc          string
		row           *nosqlplugin.VisibilityRowForInsert
		ttlSeconds    int64
		queryMockFunc func(*gocql.MockQuery)
		wantErr       bool
	}{
		{
			desc:       "Query with ttl less than maxCassandraTTL",
			row:        testdata.NewVisibilityRowForInsert(),
			ttlSeconds: int64(1000),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().WithTimestamp(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(nil)
			},
			wantErr: false,
		},
		{
			desc:       "Query With ttl greater than maxCassandraTTL",
			row:        testdata.NewVisibilityRowForInsert(),
			ttlSeconds: maxCassandraTTL + 1,
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().WithTimestamp(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(nil)
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			query := gocql.NewMockQuery(ctrl)
			if test.queryMockFunc != nil {
				test.queryMockFunc(query)
			}
			session := &fakeSession{
				query: query,
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}
			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))

			err := db.InsertVisibility(context.Background(), test.ttlSeconds, test.row)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateVisibility(t *testing.T) {
	tests := []struct {
		desc       string
		row        *nosqlplugin.VisibilityRowForUpdate
		ttlSeconds int64
		wantErr    bool
		wantPanic  bool
	}{
		{
			desc:       "Query with ttl less than maxCassandraTTL",
			row:        testdata.NewVisibilityRowForUpdate(false, true),
			ttlSeconds: int64(100),
			wantErr:    false,
			wantPanic:  false,
		},
		{
			desc:       "Query with ttl greater than maxCassandraTTL",
			row:        testdata.NewVisibilityRowForUpdate(false, true),
			ttlSeconds: maxCassandraTTL + 1,
			wantErr:    false,
			wantPanic:  false,
		},
		{
			desc:       "panic if updateCloseToOpen is set",
			row:        testdata.NewVisibilityRowForUpdate(true, true),
			ttlSeconds: int64(100),
			wantErr:    false,
			wantPanic:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.wantPanic {
					t.Errorf("test: %s, panicWanted: %v, panicOccured: %v", test.desc, test.wantPanic, r != nil)
				}
			}()
			ctrl := gomock.NewController(t)
			session := &fakeSession{
				query: gocql.NewMockQuery(ctrl),
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}
			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))
			err := db.UpdateVisibility(context.Background(), test.ttlSeconds, test.row)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSelectOneClosedWorkflow(t *testing.T) {
	tests := []struct {
		desc        string
		domainID    string
		workflowID  string
		runID       string
		itrMockFunc func(*gocql.MockIter)
		wantError   bool
		wantResult  bool
	}{
		{
			desc:       "return error when query's iter function returns nil",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			wantError:  true,
			wantResult: false,
		},
		{
			desc:       "return nil if reading closed_workflow_execution_record returns false",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
			},
			wantError:  false,
			wantResult: false,
		},
		{
			desc:       "return error if closing iterator fails",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(true)
				itr.EXPECT().Close().Return(errors.New("close error"))
			},
			wantError:  true,
			wantResult: false,
		},
		{
			desc:       "success",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(true)
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			query := gocql.NewMockQuery(ctrl)
			query.EXPECT().WithContext(gomock.Any()).Return(query)
			if test.itrMockFunc != nil {
				itr := gocql.NewMockIter(ctrl)
				test.itrMockFunc(itr)
				query.EXPECT().Iter().Return(itr)
			} else {
				query.EXPECT().Iter().Return(nil)
			}
			session := &fakeSession{
				query: query,
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}
			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))
			result, err := db.SelectOneClosedWorkflow(context.Background(), test.domainID, test.workflowID, test.runID)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if test.wantResult {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestDeleteVisibility(t *testing.T) {
	tests := []struct {
		desc           string
		domainID       string
		workflowID     string
		runID          string
		mockItr        bool
		itrMockFunc    func(*gocql.MockIter)
		queryMockFunc  func(*gocql.MockQuery)
		context        context.Context
		dc             *persistence.DynamicConfiguration
		clientMockFunc func(*gocql.MockClient)
		wantError      bool
	}{
		{
			desc:      "return nil if visibility_admin_key not present in context",
			context:   context.Background(),
			mockItr:   false,
			wantError: false,
		},
		{
			desc:       "return error when query's iter function returns nil",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query)
			},
			mockItr:   true,
			wantError: true,
		},
		{
			desc:       "return nil if reading open_workflow_execution_records returns false",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(false)
			},
			wantError: false,
		},
		{
			desc:       "return error if closing iterator fails",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Close().Return(errors.New("close error"))
			},
			wantError: true,
		},
		{
			desc:       "success with db dynamic_configuration as nil",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(2)
				query.EXPECT().Exec().Return(nil)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Close().Return(nil)
			},
			wantError: false,
		},
		{
			desc:       "success with enable_cassandra_all_consistency_level_delete db dynamic_configuration",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(2)
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(nil)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Close().Return(nil)
			},
			dc: &persistence.DynamicConfiguration{
				EnableCassandraAllConsistencyLevelDelete: func(opts ...dynamicconfig.FilterOption) bool { return true },
			},
			wantError: false,
		},
		{
			desc:       "succes with cassandra_default_consistency_level when cassandra_all_consistency_level fails",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(2)
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(errors.New("all consistency level fail"))
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(nil)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Close().Return(nil)
			},
			dc: &persistence.DynamicConfiguration{
				EnableCassandraAllConsistencyLevelDelete: func(opts ...dynamicconfig.FilterOption) bool { return true },
			},
			clientMockFunc: func(client *gocql.MockClient) {
				client.EXPECT().IsCassandraConsistencyError(gomock.Any()).Return(true)
			},
			wantError: false,
		},
		{
			desc:       "return error cassandra_all_consistency_level fails and error is not cassandra consistency error",
			domainID:   testdata.DomainID,
			workflowID: testdata.WorkflowID,
			runID:      testdata.RunID,
			context:    context.WithValue(context.Background(), persistence.VisibilityAdminDeletionKey("visibilityAdminDelete"), true),
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().WithContext(gomock.Any()).Return(query).Times(2)
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().Exec().Return(errors.New("all consistency level fail"))
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Close().Return(nil)
			},
			dc: &persistence.DynamicConfiguration{
				EnableCassandraAllConsistencyLevelDelete: func(opts ...dynamicconfig.FilterOption) bool { return true },
			},
			clientMockFunc: func(client *gocql.MockClient) {
				client.EXPECT().IsCassandraConsistencyError(gomock.Any()).Return(false)
			},
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			query := gocql.NewMockQuery(ctrl)
			if test.queryMockFunc != nil {
				test.queryMockFunc(query)
			}
			if test.mockItr && test.itrMockFunc != nil {
				itr := gocql.NewMockIter(ctrl)
				test.itrMockFunc(itr)
				query.EXPECT().Iter().Return(itr)
			} else if test.mockItr {
				query.EXPECT().Iter().Return(nil)
			}
			session := &fakeSession{
				query: query,
			}
			client := gocql.NewMockClient(ctrl)
			if test.clientMockFunc != nil {
				test.clientMockFunc(client)
			}
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			db := newCassandraDBFromSession(cfg, session, logger, test.dc, dbWithClient(client))
			err := db.DeleteVisibility(test.context, test.domainID, test.workflowID, test.runID)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSelectVisibility(t *testing.T) {
	tests := []struct {
		desc          string
		filter        *nosqlplugin.VisibilityFilter
		queryMockFunc func(query *gocql.MockQuery)
		mockItr       bool
		itrMockFunc   func(itr *gocql.MockIter)
		wantError     bool
		wantResult    bool
		wantPanic     bool
	}{
		{
			desc:      "panic for invalid filter type",
			filter:    &nosqlplugin.VisibilityFilter{FilterType: nosqlplugin.VisibilityFilterType(10)},
			wantPanic: true,
		},
		{
			desc:   "all-open type filter return error if iterator is nil",
			filter: &nosqlplugin.VisibilityFilter{FilterType: nosqlplugin.AllOpen},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr:    true,
			wantError:  true,
			wantResult: false,
		},
		{
			desc:   "all-open type filter return error if closing iterator fails",
			filter: &nosqlplugin.VisibilityFilter{FilterType: nosqlplugin.AllOpen},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Scan(generateMockParams(12)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(errors.New("close error"))
			},
			wantError:  true,
			wantResult: false,
		},
		{
			desc: "all-open type filter success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.AllOpen,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Scan(generateMockParams(12)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "all-closed type filter sort by start-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.AllClosed,
				SortType:   nosqlplugin.SortByStartTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "all-closed type filter sort by closed-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.AllClosed,
				SortType:   nosqlplugin.SortByClosedTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "panic on all-closed type filter invalid sort attribute",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.AllClosed,
				SortType:   nosqlplugin.VisibilitySortType(100),
			},
			wantPanic: true,
		},
		{
			desc: "open-by-workflow-type filter success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.OpenByWorkflowType,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Scan(generateMockParams(12)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "closed-by-workflow-type filter sort by start-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowType,
				SortType:   nosqlplugin.SortByStartTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "closed-by-workflow-type filter sort by closed-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowType,
				SortType:   nosqlplugin.SortByClosedTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "panic on closed-by-workflow-type filter invalid sort attribute",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowType,
				SortType:   nosqlplugin.VisibilitySortType(100),
			},
			wantPanic: true,
		},
		{
			desc: "open-by-workflow-id filter success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.OpenByWorkflowID,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(12)...).Return(true)
				itr.EXPECT().Scan(generateMockParams(12)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "closed-by-workflow-id filter sort by start-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowID,
				SortType:   nosqlplugin.SortByStartTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "closed-by-workflow-id filter sort by closed-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowID,
				SortType:   nosqlplugin.SortByClosedTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "panic on closed-by-workflow-id filter invalid sort attribute",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByWorkflowID,
				SortType:   nosqlplugin.VisibilitySortType(100),
			},
			wantPanic: true,
		},
		{
			desc: "closed-by-closed-status filter sort by start-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByClosedStatus,
				SortType:   nosqlplugin.SortByStartTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "closed-by-closed-status filter sort by closed-time success",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByClosedStatus,
				SortType:   nosqlplugin.SortByClosedTime,
			},
			queryMockFunc: func(query *gocql.MockQuery) {
				query.EXPECT().Consistency(gomock.Any()).Return(query)
				query.EXPECT().WithContext(gomock.Any()).Return(query)
				query.EXPECT().PageSize(gomock.Any()).Return(query)
				query.EXPECT().PageState(gomock.Any()).Return(query)
			},
			mockItr: true,
			itrMockFunc: func(itr *gocql.MockIter) {
				itr.EXPECT().Scan(generateMockParams(15)...).Return(false)
				itr.EXPECT().PageState().Return([]byte("test"))
				itr.EXPECT().Close().Return(nil)
			},
			wantError:  false,
			wantResult: true,
		},
		{
			desc: "panic on closed-by-closed-status filter invalid sort attribute",
			filter: &nosqlplugin.VisibilityFilter{
				FilterType: nosqlplugin.ClosedByClosedStatus,
				SortType:   nosqlplugin.VisibilitySortType(100),
			},
			wantPanic: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.wantPanic {
					t.Errorf("test: %s, panicWanted: %v, panicOccured: %v", test.desc, test.wantPanic, r != nil)
				}
			}()

			ctrl := gomock.NewController(t)
			query := gocql.NewMockQuery(ctrl)
			if test.queryMockFunc != nil {
				test.queryMockFunc(query)
			}
			if test.mockItr && test.itrMockFunc != nil {
				itr := gocql.NewMockIter(ctrl)
				test.itrMockFunc(itr)
				query.EXPECT().Iter().Return(itr)
			} else if test.mockItr {
				query.EXPECT().Iter().Return(nil)
			}
			session := &fakeSession{
				query: query,
			}
			client := gocql.NewMockClient(ctrl)
			cfg := &config.NoSQL{}
			logger := testlogger.New(t)
			dc := &persistence.DynamicConfiguration{}
			db := newCassandraDBFromSession(cfg, session, logger, dc, dbWithClient(client))
			result, err := db.SelectVisibility(context.Background(), test.filter)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if test.wantResult {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func generateMockParams(count int) []interface{} {
	params := []interface{}{}
	for i := 0; i < count; i++ {
		params = append(params, gomock.Any())
	}
	return params
}
