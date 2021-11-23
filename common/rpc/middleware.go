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

	"go.uber.org/yarpc/api/transport"
)

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
