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

package resource

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/isolationgroup"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

func TestEnsureIsolationGroupImpl(t *testing.T) {

	ctrl := gomock.NewController(t)

	tests := map[string]struct {
		params       *Params
		cfg          *service.Config
		allIGs       []string
		dcAffordance func(client *dynamicconfig.MockClient)
		expectedErr  error
	}{
		"default config - all values set": {
			params: &Params{
				Name:                "some-service",
				Logger:              loggerimpl.NewNopLogger(),
				IsolationGroupState: nil,
				Partitioner:         nil,
			},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(dynamicconfig.AllIsolationGroups, gomock.Any()).Return([]interface{}{"zone-1", "zone-2", "zone-3"}, nil)
			},
		},
		"empty values - the cluster hasn't been configured for this feature": {
			params: &Params{
				Name:                "some-service",
				Logger:              loggerimpl.NewNopLogger(),
				IsolationGroupState: nil,
				Partitioner:         nil,
			},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(dynamicconfig.AllIsolationGroups, gomock.Any()).Return(nil, nil)
			},
		},
		"passed in overrides": {
			params: &Params{
				Name:                "some-service",
				Logger:              loggerimpl.NewNopLogger(),
				IsolationGroupState: isolationgroup.NewMockState(ctrl),
				Partitioner:         nil,
			},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(dynamicconfig.AllIsolationGroups, gomock.Any()).Return(nil, nil)
			},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			gomock := gomock.NewController(t)
			dcMock := dynamicconfig.NewMockClient(gomock)
			domainMgr := persistence.NewMockDomainManager(gomock)
			domainCache := cache.NewDomainCache(domainMgr, cluster.Metadata{}, metrics.NewNoopMetricsClient(), loggerimpl.NewNopLogger())
			coll := dynamicconfig.NewCollection(dcMock, loggerimpl.NewNopLogger())
			td.dcAffordance(dcMock)
			cfgStore := persistence.NewMockConfigStoreManager(gomock)
			stopChan := make(chan struct{})
			state := ensureIsolationGroupStateHandlerOrDefault(td.params, cfgStore, coll, domainCache, stopChan)
			assert.NotNil(t, state)
		})
	}
}
