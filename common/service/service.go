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

package service

import (
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/authorization"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	es "github.com/uber/cadence/common/elasticsearch"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/rpc"
)

type (
	// BootstrapParams holds the set of parameters
	// needed to bootstrap a service
	BootstrapParams struct {
		Name            string
		InstanceID      string
		Logger          log.Logger
		ThrottledLogger log.Logger

		MetricScope         tally.Scope
		MembershipFactory   MembershipMonitorFactory
		RPCFactory          common.RPCFactory
		PProfInitializer    common.PProfInitializer
		PersistenceConfig   config.Persistence
		ClusterMetadata     cluster.Metadata
		ReplicatorConfig    config.Replicator
		MetricsClient       metrics.Client
		MessagingClient     messaging.Client
		BlobstoreClient     blobstore.Client
		ESClient            es.GenericClient
		ESConfig            *config.ElasticSearchConfig
		DynamicConfig       dynamicconfig.Client
		DispatcherProvider  rpc.DispatcherProvider
		DCRedirectionPolicy config.DCRedirectionPolicy
		PublicClient        workflowserviceclient.Interface
		ArchivalMetadata    archiver.ArchivalMetadata
		ArchiverProvider    provider.ArchiverProvider
		Authorizer          authorization.Authorizer // NOTE: this can be nil. If nil, AccessControlledHandlerImpl will initiate one with config.Authorization
		AuthorizationConfig config.Authorization     // NOTE: empty(default) struct will get a authorization.NoopAuthorizer
	}

	// MembershipMonitorFactory provides a bootstrapped membership monitor
	MembershipMonitorFactory interface {
		// GetMembershipMonitor return a membership monitor
		GetMembershipMonitor() (membership.Monitor, error)
	}
)

// UpdateLoggerWithServiceName tag logging with service name from the top level
func (params *BootstrapParams) UpdateLoggerWithServiceName(name string) {
	if params.Logger != nil {
		params.Logger = params.Logger.WithTags(tag.Service(name))
	}
	if params.ThrottledLogger != nil {
		params.ThrottledLogger = params.ThrottledLogger.WithTags(tag.Service(name))
	}
}

// GetMetricsServiceIdx returns the metrics name
func GetMetricsServiceIdx(serviceName string, logger log.Logger) metrics.ServiceIdx {
	switch serviceName {
	case common.FrontendServiceName:
		return metrics.Frontend
	case common.HistoryServiceName:
		return metrics.History
	case common.MatchingServiceName:
		return metrics.Matching
	case common.WorkerServiceName:
		return metrics.Worker
	default:
		logger.Fatal("Unknown service name for metrics!")
	}

	// this should never happen!
	return metrics.NumServices
}
