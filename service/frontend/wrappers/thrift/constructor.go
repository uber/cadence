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

package thrift

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/admin/adminserviceserver"
	"github.com/uber/cadence/.gen/go/cadence/workflowserviceserver"
	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	"github.com/uber/cadence/common/types/mapper/thrift"
	"github.com/uber/cadence/service/frontend/admin"
	"github.com/uber/cadence/service/frontend/api"
)

type (
	AdminHandler struct {
		h admin.Handler
	}
	APIHandler struct {
		h api.Handler
	}
)

func NewAdminHandler(h admin.Handler) AdminHandler {
	return AdminHandler{h}
}

func NewAPIHandler(h api.Handler) APIHandler {
	return APIHandler{h}
}

func (t AdminHandler) Register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(adminserviceserver.New(t))
}

func (t APIHandler) Register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(workflowserviceserver.New(t))
	dispatcher.Register(metaserver.New(t))
}

func (t APIHandler) Health(ctx context.Context) (*health.HealthStatus, error) {
	response, err := t.h.Health(ctx)
	return thrift.FromHealthStatus(response), thrift.FromError(err)
}
