// Copyright (c) 2021 Uber Technologies, Inc.
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

package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
)

func TestDirectChooser(t *testing.T) {
	newReq := func(shardKey string) *transport.Request {
		return &transport.Request{
			Caller:   "caller",
			Service:  "service",
			ShardKey: shardKey,
		}
	}

	tests := []struct {
		desc           string
		retainConn     bool
		req            *transport.Request
		multipleChoose bool
		wantChooseErr  bool
	}{
		{
			desc:       "legacy chooser",
			retainConn: false,
			req:        newReq("key"),
		},
		{
			desc:          "legacy chooser - empty shard key",
			retainConn:    false,
			req:           newReq(""),
			wantChooseErr: true,
		},
		{
			desc:       "connection retain mode",
			retainConn: true,
			req:        newReq("key"),
		},
		{
			desc:          "connection retain mode - empty shard key",
			retainConn:    true,
			req:           newReq(""),
			wantChooseErr: true,
		},
		{
			desc:           "connection retain mode - multiple choose should return chooser from cache",
			retainConn:     true,
			req:            newReq("key"),
			multipleChoose: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			logger := testlogger.New(t)
			metricCl := metrics.NewNoopMetricsClient()
			serviceName := "service"
			directConnRetainFn := func(opts ...dynamicconfig.FilterOption) bool { return tc.retainConn }
			grpcTransport := grpc.NewTransport()

			chooser := newDirectChooser(serviceName, grpcTransport, logger, metricCl, directConnRetainFn)

			assert.False(t, chooser.IsRunning(), "expected IsRunning()=false before Start()")

			if err := chooser.Start(); err != nil {
				t.Fatalf("failed to start direct peer chooser: %v", err)
			}

			assert.NoError(t, chooser.Start(), "starting again should be no-op")

			assert.True(t, chooser.IsRunning())

			peer, onFinish, err := chooser.Choose(context.Background(), tc.req)
			if tc.wantChooseErr != (err != nil) {
				t.Fatalf("Choose() err = %v, wantChooseErr = %v", err, tc.wantChooseErr)
			}

			if err == nil {
				assert.NotNil(t, peer)
				assert.NotNil(t, onFinish)

				// call onFinish will release the peer for legacy chooser
				onFinish(nil)
			}

			if tc.multipleChoose {
				peer2, onFinish2, err2 := chooser.Choose(context.Background(), tc.req)
				assert.NoError(t, err2)
				assert.NotNil(t, onFinish2)
				assert.Equal(t, peer, peer2)
				onFinish2(nil)
			}

			if err := chooser.Stop(); err != nil {
				t.Fatalf("failed to stop direct peer chooser: %v", err)
			}

			assert.NoError(t, chooser.Stop(), "stopping again should be no-op")
		})
	}
}
