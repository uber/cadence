// Copyright (c) 2020 Uber Technologies, Inc.
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

package thrift

import (
	"github.com/uber/cadence/.gen/go/admin/adminserviceclient"
	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/.gen/go/matching/matchingserviceclient"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
)

type (
	adminClient struct {
		c adminserviceclient.Interface
	}
	frontendClient struct {
		c workflowserviceclient.Interface
	}
	historyClient struct {
		c historyserviceclient.Interface
	}
	matchingClient struct {
		c matchingserviceclient.Interface
	}
)

func NewAdminClient(c adminserviceclient.Interface) admin.Client {
	return adminClient{c}
}

func NewFrontendClient(c workflowserviceclient.Interface) frontend.Client {
	return frontendClient{c}
}

func NewHistoryClient(c historyserviceclient.Interface) history.Client {
	return historyClient{c}
}

func NewMatchingClient(c matchingserviceclient.Interface) matching.Client {
	return matchingClient{c}
}
