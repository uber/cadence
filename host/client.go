// Copyright (c) 2017 Uber Technologies, Inc.
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

package host

import (
	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/admin/adminserviceclient"
	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common/service"
)

// AdminClient is the interface exposed by admin service client
type AdminClient interface {
	admin.Client
}

// FrontendClient is the interface exposed by frontend service client
type FrontendClient interface {
	frontend.Client
}

// HistoryClient is the interface exposed by history service client
type HistoryClient interface {
	history.Client
}

// NewAdminClient creates a client to cadence admin client
func NewAdminClient(d *yarpc.Dispatcher) AdminClient {
	return admin.NewThriftClient(adminserviceclient.New(d.ClientConfig(testOutboundName(service.Frontend))))
}

// NewFrontendClient creates a client to cadence frontend client
func NewFrontendClient(d *yarpc.Dispatcher) FrontendClient {
	return frontend.NewThriftClient(workflowserviceclient.New(d.ClientConfig(testOutboundName(service.Frontend))))
}

// NewHistoryClient creates a client to cadence history service client
func NewHistoryClient(d *yarpc.Dispatcher) HistoryClient {
	return history.NewThriftClient(historyserviceclient.New(d.ClientConfig(testOutboundName(service.History))))
}
