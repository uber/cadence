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
	"io"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"

	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
)

type authOutboundMiddleware struct {
	authProvider worker.AuthorizationProvider
}

func (m *authOutboundMiddleware) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	if m.authProvider == nil {
		return out.Call(ctx, request)
	}

	token, err := m.authProvider.GetAuthToken()
	if err != nil {
		return nil, err
	}
	request.Headers = request.Headers.
		With(common.AuthorizationTokenHeaderName, string(token))

	return out.Call(ctx, request)
}

const _responseInfoContextKey = "response-info"

// ContextWithResponseInfo will create a child context that has ResponseInfo set as value.
// This value will get filled after the call is made and can be used later to retrieve some info of interest.
func ContextWithResponseInfo(parent context.Context) (context.Context, *ResponseInfo) {
	responseInfo := &ResponseInfo{}
	return context.WithValue(parent, _responseInfoContextKey, responseInfo), responseInfo
}

// ResponseInfo structure is filled with data after the RPC call.
// It can be obtained with rpc.ContextWithResponseInfo function.
type ResponseInfo struct {
	Size int
}

type countingReadCloser struct {
	reader    io.ReadCloser
	bytesRead *int
}

func (r *countingReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	*r.bytesRead += n
	return n, err
}

func (r *countingReadCloser) Close() (err error) {
	return r.reader.Close()
}

// ResponseInfoMiddleware populates context with ResponseInfo structure which contains info about response that was received.
// In particular, it counts the size of the response in bytes. Such information can be useful down the line, where payload are deserialized and no longer have their size.
type ResponseInfoMiddleware struct{}

func (m *ResponseInfoMiddleware) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	response, err := out.Call(ctx, request)

	if value := ctx.Value(_responseInfoContextKey); value != nil {
		if responseInfo, ok := value.(*ResponseInfo); ok && response != nil {
			// We can not use response.BodySize here, because it is not set on all transports.
			// Instead wrap body reader with counter, that increments responseInfo.Size as it is read.
			response.Body = &countingReadCloser{reader: response.Body, bytesRead: &responseInfo.Size}
		}
	}

	return response, err
}

// InboundMetricsMiddleware tags context with additional metric tags from incoming request.
type InboundMetricsMiddleware struct{}

func (m *InboundMetricsMiddleware) Handle(ctx context.Context, req *transport.Request, resw transport.ResponseWriter, h transport.UnaryHandler) error {
	ctx = metrics.TagContext(ctx,
		metrics.CallerTag(req.Caller),
		metrics.TransportTag(req.Transport),
	)
	return h.Handle(ctx, req, resw)
}

type overrideCallerMiddleware struct {
	caller string
}

func (m *overrideCallerMiddleware) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	request.Caller = m.caller
	return out.Call(ctx, request)
}

// HeaderForwardingMiddleware forwards headers from current inbound RPC call that is being handled to new outbound calls being made.
// If new value for the same header key is provided in the outbound request, keep it instead.
type HeaderForwardingMiddleware struct{}

func (m *HeaderForwardingMiddleware) Call(ctx context.Context, request *transport.Request, out transport.UnaryOutbound) (*transport.Response, error) {
	if inboundCall := yarpc.CallFromContext(ctx); inboundCall != nil && request != nil {
		outboundHeaders := request.Headers
		for _, key := range inboundCall.HeaderNames() {
			if _, exists := outboundHeaders.Get(key); !exists {
				value := inboundCall.Header(key)
				outboundHeaders = outboundHeaders.With(key, value)
			}
		}
		request.Headers = outboundHeaders
	}

	return out.Call(ctx, request)
}
