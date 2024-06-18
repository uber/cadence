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

package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/yarpc"
	"golang.org/x/exp/maps"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas/global/algorithm"
	"github.com/uber/cadence/common/types"
)

func TestClient(t *testing.T) {
	setup := func(t *testing.T) (c *client, hc *history.MockClient, pr *history.MockPeerResolver) {
		ctrl := gomock.NewController(t)
		hc = history.NewMockClient(ctrl)
		pr = history.NewMockPeerResolver(ctrl)
		impl := New(hc, pr, testlogger.New(t), metrics.NewNoopMetricsClient())
		return impl.(*client), hc, pr
	}

	t.Run("valid request", func(t *testing.T) {
		c, hc, pr := setup(t)
		data := map[string]Calls{
			"a": {10, 20},
			"b": {3, 5},
			"c": {78, 9},
		}
		response := map[string]float64{
			"a": 0.1,
			"b": 0.2,
			"c": 0.3,
		}
		encode := func(w map[algorithm.Limit]algorithm.HostWeight) (*types.RatelimitUpdateResponse, error) {
			a, err := AggregatorWeightsToAny(w)
			assert.NoError(t, err, "should be impossible: could not encode aggregator response to any")
			return &types.RatelimitUpdateResponse{
				Any: a,
			}, nil
		}
		pr.EXPECT().GlobalRatelimitPeers(gomock.InAnyOrder(maps.Keys(data))).Return(map[string][]string{
			"agg-1": {"a", "c"},
			"agg-2": {"b"},
		}, nil)
		hc.EXPECT().
			RatelimitUpdate(gomock.Any(), matchrequest{t, []string{"a", "c"}}, matchyarpc{t, yarpc.WithShardKey("agg-1")}).
			Return(encode(map[algorithm.Limit]algorithm.HostWeight{"a": 0.1, "c": 0.3}))
		hc.EXPECT().
			RatelimitUpdate(gomock.Any(), matchrequest{t, []string{"b"}}, matchyarpc{t, yarpc.WithShardKey("agg-2")}).
			Return(encode(map[algorithm.Limit]algorithm.HostWeight{"b": 0.2}))

		result := c.Update(context.Background(), time.Second, data)
		assert.NoError(t, result.Err)
		assert.Equal(t, response, result.Weights)
	})
	t.Run("bad peers", func(t *testing.T) {
		c, _, pr := setup(t)
		err := errors.New("peer issue")
		pr.EXPECT().GlobalRatelimitPeers(gomock.Any()).Return(nil, err)
		res := c.Update(context.Background(), time.Second, nil)
		assert.ErrorContains(t, res.Err, "unable to shard")
		assert.ErrorIs(t, res.Err, err) // should contain the cause
	})
}

func stringy[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = string(in[i])
	}
	return out
}

// matches a single yarpc option
type matchyarpc struct {
	t   *testing.T
	opt yarpc.CallOption
}

var _ gomock.Matcher = matchyarpc{}

func (m matchyarpc) Matches(x interface{}) bool {
	opts, ok := x.([]yarpc.CallOption)
	if !ok {
		m.t.Logf("not a []yarpc.CallOption: %T", x)
		return false
	}
	m.t.Logf("opts: %#v, should be: %#v", opts, m.opt)
	return len(opts) == 1 && reflect.DeepEqual(opts[0], m.opt)
}

func (m matchyarpc) String() string {
	return fmt.Sprintf("%#v", m.opt)
}

// matches load-keys in an update request
type matchrequest struct {
	t        *testing.T
	loadKeys []string
}

var _ gomock.Matcher = matchrequest{}

func (a matchrequest) Matches(x interface{}) bool {
	v, ok := x.(*types.RatelimitUpdateRequest)
	if !ok {
		a.t.Logf("not a *types.RatelimitUpdateRequest: %T", x)
		return false
	}
	up, err := AnyToAggregatorUpdate(v.Any)
	if err != nil {
		a.t.Logf("failed to decode to agg update: %v", err)
		return false
	}
	keys := map[string]struct{}{}
	for _, k := range a.loadKeys {
		keys[k] = struct{}{}
	}
	gotKeys := map[string]struct{}{}
	for _, k := range stringy(maps.Keys(up.Load)) {
		gotKeys[k] = struct{}{}
	}
	a.t.Logf("want keys: %v, got keys: %v", maps.Keys(keys), maps.Keys(gotKeys))
	return reflect.DeepEqual(keys, gotKeys)
}

func (a matchrequest) String() string {
	return fmt.Sprintf("request with load keys: %v", a.loadKeys)
}
