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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/yarpctest"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
)

func TestAuthOubboundMiddleware(t *testing.T) {
	m := authOutboundMiddleware{}
	_, err := m.Call(context.Background(), &transport.Request{}, &fakeOutbound{verify: func(_ context.Context, request *transport.Request) {
		assert.Empty(t, request.Headers)
	}})
	assert.NoError(t, err)

	m = authOutboundMiddleware{fakeAuthProvider{err: assert.AnError}}
	_, err = m.Call(context.Background(), &transport.Request{}, &fakeOutbound{})
	assert.Error(t, err)

	m = authOutboundMiddleware{fakeAuthProvider{token: []byte("token")}}
	_, err = m.Call(context.Background(), &transport.Request{}, &fakeOutbound{verify: func(_ context.Context, request *transport.Request) {
		assert.Equal(t, "token", request.Headers.Items()[common.AuthorizationTokenHeaderName])
	}})
	assert.NoError(t, err)
}

func TestResponseInfoMiddleware(t *testing.T) {
	m := responseInfoMiddleware{}
	ctx, responseInfo := ContextWithResponseInfo(context.Background())
	_, err := m.Call(ctx, &transport.Request{}, &fakeOutbound{response: &transport.Response{BodySize: 12345}})
	assert.NoError(t, err)
	assert.Equal(t, 12345, responseInfo.Size)
}

func TestResponseInfoMiddleware_Error(t *testing.T) {
	m := responseInfoMiddleware{}
	ctx, responseInfo := ContextWithResponseInfo(context.Background())
	_, err := m.Call(ctx, &transport.Request{}, &fakeOutbound{err: fmt.Errorf("test")})
	assert.Error(t, err)
	assert.Equal(t, 0, responseInfo.Size)
}

func TestInboundMetricsMiddleware(t *testing.T) {
	m := inboundMetricsMiddleware{}
	h := &fakeHandler{}
	err := m.Handle(context.Background(), &transport.Request{Transport: "grpc", Caller: "x-caller"}, nil, h)
	assert.NoError(t, err)
	assert.ElementsMatch(t, metrics.GetContextTags(h.ctx), []metrics.Tag{
		metrics.TransportTag("grpc"),
		metrics.CallerTag("x-caller"),
	})
}

func TestOverrideCallerMiddleware(t *testing.T) {
	m := overrideCallerMiddleware{"x-caller"}
	_, err := m.Call(context.Background(), &transport.Request{Caller: "service"}, &fakeOutbound{verify: func(_ context.Context, r *transport.Request) {
		assert.Equal(t, "x-caller", r.Caller)
	}})
	assert.NoError(t, err)
}

func TestHeaderForwardingMiddleware(t *testing.T) {
	inboundHeaders := map[string]string{
		"key-a": "inbound-value-a",
		"key-b": "inbound-value-b",
	}
	outboundHeaders := map[string]string{
		"key-b": "outbound-value-b",
		"key-c": "outbound-value-c",
	}
	combinedHeaders := map[string]string{
		"key-a": "inbound-value-a",
		"key-b": "outbound-value-b",
		"key-c": "outbound-value-c",
	}

	ctx := yarpctest.ContextWithCall(context.Background(), &yarpctest.Call{Headers: inboundHeaders})
	request := transport.Request{Headers: transport.HeadersFromMap(outboundHeaders)}

	m := HeaderForwardingMiddleware{}
	// No ongoing inbound call -> keep existing outbound headers
	_, err := m.Call(context.Background(), &request, &fakeOutbound{verify: func(_ context.Context, r *transport.Request) {
		assert.Equal(t, outboundHeaders, r.Headers.Items())
	}})
	assert.NoError(t, err)

	// With ongoing inbound call -> forward inbound headers not present in the request
	_, err = m.Call(ctx, &request, &fakeOutbound{verify: func(_ context.Context, r *transport.Request) {
		assert.Equal(t, combinedHeaders, r.Headers.Items())
	}})
	assert.NoError(t, err)
}

func TestOutboundTimeoutMiddleware(t *testing.T) {
	ctx := context.Background()

	m := &outboundTimeoutMiddleware{}
	m.Call(ctx, &transport.Request{}, &fakeOutbound{verify: func(newCtx context.Context, _ *transport.Request) {
		assert.Equal(t, ctx, newCtx)
	}})

	m = &outboundTimeoutMiddleware{timeoutFn: func(method string) *time.Duration {
		switch method {
		case "A":
			duration := time.Minute
			return &duration
		case "B":
			return nil
		default:
			return nil
		}
	}}

	m.Call(ctx, &transport.Request{Procedure: "service::A"}, &fakeOutbound{verify: func(newCtx context.Context, _ *transport.Request) {
		deadline, ok := newCtx.Deadline()
		assert.True(t, ok)
		assert.InDelta(t, time.Now().Add(time.Minute).Unix(), deadline.Unix(), float64(time.Second))
	}})

	m.Call(ctx, &transport.Request{Procedure: "service::B"}, &fakeOutbound{verify: func(newCtx context.Context, _ *transport.Request) {
		assert.Equal(t, ctx, newCtx)
	}})

	m.Call(ctx, &transport.Request{Procedure: "service::C"}, &fakeOutbound{verify: func(newCtx context.Context, _ *transport.Request) {
		assert.Equal(t, ctx, newCtx)
	}})
}

type fakeHandler struct {
	ctx context.Context
}

func (h *fakeHandler) Handle(ctx context.Context, req *transport.Request, resw transport.ResponseWriter) error {
	h.ctx = ctx
	return nil
}

type fakeOutbound struct {
	verify   func(context.Context, *transport.Request)
	response *transport.Response
	err      error
}

func (o fakeOutbound) Call(ctx context.Context, request *transport.Request) (*transport.Response, error) {
	if o.verify != nil {
		o.verify(ctx, request)
	}
	return o.response, o.err
}
func (o fakeOutbound) Start() error                      { return nil }
func (o fakeOutbound) Stop() error                       { return nil }
func (o fakeOutbound) IsRunning() bool                   { return true }
func (o fakeOutbound) Transports() []transport.Transport { return nil }

type fakeAuthProvider struct {
	token []byte
	err   error
}

func (p fakeAuthProvider) GetAuthToken() ([]byte, error) {
	return p.token, p.err
}
