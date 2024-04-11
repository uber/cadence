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

package clusterredirection

import (
	"time"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

type (
	domainIDGetter interface {
		GetDomainID() string
	}
)

func (handler *clusterRedirectionHandler) beforeCall(
	scope int,
) (metrics.Scope, time.Time) {
	return handler.GetMetricsClient().Scope(scope), handler.GetTimeSource().Now()
}

func (handler *clusterRedirectionHandler) afterCall(
	recovered interface{},
	scope metrics.Scope,
	startTime time.Time,
	domainName string,
	domainID string,
	cluster string,
	retError *error,
) {
	var extraTags []tag.Tag
	if domainName != "" {
		extraTags = append(extraTags, tag.WorkflowDomainName(domainName))
	}
	if domainID != "" {
		extraTags = append(extraTags, tag.WorkflowDomainID(domainID))
	}
	log.CapturePanic(recovered, handler.GetLogger().WithTags(extraTags...), retError)

	scope = scope.Tagged(metrics.TargetClusterTag(cluster))
	scope.IncCounter(metrics.CadenceDcRedirectionClientRequests)
	scope.RecordTimer(metrics.CadenceDcRedirectionClientLatency, handler.GetTimeSource().Now().Sub(startTime))
	if *retError != nil {
		scope.IncCounter(metrics.CadenceDcRedirectionClientFailures)
	}
}

// noopdomainIDGetter is a domainIDGetter that always returns empty string.
// it is used for extraction of domainID from domainIDGetter in case of token extraction failure.
type noopdomainIDGetter struct{}

func (noopdomainIDGetter) GetDomainID() string {
	return ""
}
