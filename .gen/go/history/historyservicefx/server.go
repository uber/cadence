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

// Code generated by thriftrw-plugin-yarpc
// @generated

package historyservicefx

import (
	historyserviceserver "github.com/uber/cadence/.gen/go/history/historyserviceserver"
	fx "go.uber.org/fx"
	transport "go.uber.org/yarpc/api/transport"
	thrift "go.uber.org/yarpc/encoding/thrift"
)

// ServerParams defines the dependencies for the HistoryService server.
type ServerParams struct {
	fx.In

	Handler historyserviceserver.Interface
}

// ServerResult defines the output of HistoryService server module. It provides the
// procedures of a HistoryService handler to an Fx application.
//
// The procedures are provided to the "yarpcfx" value group. Dig 1.2 or newer
// must be used for this feature to work.
type ServerResult struct {
	fx.Out

	Procedures []transport.Procedure `group:"yarpcfx"`
}

// Server provides procedures for HistoryService to an Fx application. It expects a
// historyservicefx.Interface to be present in the container.
//
// 	fx.Provide(
// 		func(h *MyHistoryServiceHandler) historyserviceserver.Interface {
// 			return h
// 		},
// 		historyservicefx.Server(),
// 	)
func Server(opts ...thrift.RegisterOption) interface{} {
	return func(p ServerParams) ServerResult {
		procedures := historyserviceserver.New(p.Handler, opts...)
		return ServerResult{Procedures: procedures}
	}
}
