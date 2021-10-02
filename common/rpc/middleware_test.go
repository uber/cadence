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
	"go.uber.org/yarpc/api/transport"

	"github.com/uber/cadence/common"
)

func TestAuthOubboundMiddleware(t *testing.T) {
	m := authOutboundMiddleware{}
	_, err := m.Call(context.Background(), &transport.Request{}, &fakeOutbound{verify: func(request *transport.Request) {
		assert.Empty(t, request.Headers)
	}})
	assert.NoError(t, err)

	m = authOutboundMiddleware{fakeAuthProvider{err: assert.AnError}}
	_, err = m.Call(context.Background(), &transport.Request{}, &fakeOutbound{})
	assert.Error(t, err)

	m = authOutboundMiddleware{fakeAuthProvider{token: []byte("token")}}
	_, err = m.Call(context.Background(), &transport.Request{}, &fakeOutbound{verify: func(request *transport.Request) {
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

type fakeOutbound struct {
	verify   func(*transport.Request)
	response *transport.Response
}

func (o fakeOutbound) Call(ctx context.Context, request *transport.Request) (*transport.Response, error) {
	if o.verify != nil {
		o.verify(request)
	}
	return o.response, nil
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
