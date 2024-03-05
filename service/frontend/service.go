// Copyright (c) 2019 Uber Technologies, Inc.
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

package frontend

import (
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/client"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/service/frontend/admin"
	"github.com/uber/cadence/service/frontend/api"
	"github.com/uber/cadence/service/frontend/config"
	"github.com/uber/cadence/service/frontend/wrappers/accesscontrolled"
	"github.com/uber/cadence/service/frontend/wrappers/clusterredirection"
	"github.com/uber/cadence/service/frontend/wrappers/grpc"
	"github.com/uber/cadence/service/frontend/wrappers/metered"
	"github.com/uber/cadence/service/frontend/wrappers/ratelimited"
	"github.com/uber/cadence/service/frontend/wrappers/thrift"
)

// Service represents the cadence-frontend service
type Service struct {
	resource.Resource

	status       int32
	handler      *api.WorkflowHandler
	adminHandler admin.Handler
	stopC        chan struct{}
	config       *config.Config
	params       *resource.Params
}

// NewService builds a new cadence-frontend service
func NewService(
	params *resource.Params,
) (resource.Resource, error) {

	isAdvancedVisExistInConfig := len(params.PersistenceConfig.AdvancedVisibilityStore) != 0
	serviceConfig := config.NewConfig(
		dynamicconfig.NewCollection(
			params.DynamicConfig,
			params.Logger,
			dynamicconfig.ClusterNameFilter(params.ClusterMetadata.GetCurrentClusterName()),
		),
		params.PersistenceConfig.NumHistoryShards,
		isAdvancedVisExistInConfig,
		params.HostName,
	)
	params.PersistenceConfig.HistoryMaxConns = serviceConfig.HistoryMgrNumConns()

	serviceResource, err := resource.New(
		params,
		service.Frontend,
		&service.Config{
			PersistenceMaxQPS:       serviceConfig.PersistenceMaxQPS,
			PersistenceGlobalMaxQPS: serviceConfig.PersistenceGlobalMaxQPS,
			ThrottledLoggerMaxRPS:   serviceConfig.ThrottledLogRPS,

			EnableReadVisibilityFromES:      serviceConfig.EnableReadVisibilityFromES,
			AdvancedVisibilityWritingMode:   nil, // frontend service never write
			EnableReadVisibilityFromPinot:   serviceConfig.EnableReadVisibilityFromPinot,
			EnableLogCustomerQueryParameter: serviceConfig.EnableLogCustomerQueryParameter,

			EnableDBVisibilitySampling:                  serviceConfig.EnableVisibilitySampling,
			EnableReadDBVisibilityFromClosedExecutionV2: serviceConfig.EnableReadFromClosedExecutionV2,
			DBVisibilityListMaxQPS:                      serviceConfig.VisibilityListMaxQPS,
			WriteDBVisibilityOpenMaxQPS:                 nil, // frontend service never write
			WriteDBVisibilityClosedMaxQPS:               nil, // frontend service never write

			ESVisibilityListMaxQPS: serviceConfig.ESVisibilityListMaxQPS,
			ESIndexMaxResultWindow: serviceConfig.ESIndexMaxResultWindow,
			ValidSearchAttributes:  serviceConfig.ValidSearchAttributes,
		},
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		Resource: serviceResource,
		status:   common.DaemonStatusInitialized,
		config:   serviceConfig,
		stopC:    make(chan struct{}),
		params:   params,
	}, nil
}

// Start starts the service
func (s *Service) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	logger := s.GetLogger()
	logger.Info("frontend starting")

	// domain handler's shared between admin and workflow handler, so instantiate it centrally and share it
	dh := domain.NewHandler(
		s.config.DomainConfig,
		s.GetLogger(),
		s.GetDomainManager(),
		s.GetClusterMetadata(),
		domain.NewDomainReplicator(s.GetDomainReplicationQueue(), s.GetLogger()),
		s.GetArchivalMetadata(),
		s.GetArchiverProvider(),
		s.GetTimeSource(),
	)

	// Base handler
	s.handler = api.NewWorkflowHandler(s, s.config, client.NewVersionChecker(), dh)

	userRateLimiter := quotas.NewMultiStageRateLimiter(
		quotas.NewDynamicRateLimiter(s.config.UserRPS.AsFloat64()),
		quotas.NewCollection(quotas.NewPerMemberDynamicRateLimiterFactory(
			service.Frontend,
			s.config.GlobalDomainUserRPS,
			s.config.MaxDomainUserRPSPerInstance,
			s.GetMembershipResolver(),
		)),
	)
	workerRateLimiter := quotas.NewMultiStageRateLimiter(
		quotas.NewDynamicRateLimiter(s.config.WorkerRPS.AsFloat64()),
		quotas.NewCollection(quotas.NewPerMemberDynamicRateLimiterFactory(
			service.Frontend,
			s.config.GlobalDomainWorkerRPS,
			s.config.MaxDomainWorkerRPSPerInstance,
			s.GetMembershipResolver(),
		)),
	)
	visibilityRateLimiter := quotas.NewMultiStageRateLimiter(
		quotas.NewDynamicRateLimiter(s.config.VisibilityRPS.AsFloat64()),
		quotas.NewCollection(quotas.NewPerMemberDynamicRateLimiterFactory(
			service.Frontend,
			s.config.GlobalDomainVisibilityRPS,
			s.config.MaxDomainVisibilityRPSPerInstance,
			s.GetMembershipResolver(),
		)),
	)
	asyncRateLimiter := quotas.NewMultiStageRateLimiter(
		quotas.NewDynamicRateLimiter(s.config.AsyncRPS.AsFloat64()),
		quotas.NewCollection(quotas.NewPerMemberDynamicRateLimiterFactory(
			service.Frontend,
			s.config.GlobalDomainAsyncRPS,
			s.config.MaxDomainAsyncRPSPerInstance,
			s.GetMembershipResolver(),
		)),
	)
	// Additional decorations
	var handler api.Handler = s.handler
	handler = ratelimited.NewAPIHandler(handler, s.GetDomainCache(), userRateLimiter, workerRateLimiter, visibilityRateLimiter, asyncRateLimiter)
	handler = metered.NewAPIHandler(handler, s.GetLogger(), s.GetMetricsClient(), s.GetDomainCache(), s.config)
	if s.params.ClusterRedirectionPolicy != nil {
		handler = clusterredirection.NewAPIHandler(handler, s, s.config, *s.params.ClusterRedirectionPolicy)
	}
	handler = accesscontrolled.NewAPIHandler(handler, s, s.params.Authorizer, s.params.AuthorizationConfig)

	// Register the latest (most decorated) handler
	thriftHandler := thrift.NewAPIHandler(handler)
	thriftHandler.Register(s.GetDispatcher())

	grpcHandler := grpc.NewAPIHandler(handler)
	grpcHandler.Register(s.GetDispatcher())

	s.adminHandler = admin.NewHandler(s, s.params, s.config, dh)
	s.adminHandler = accesscontrolled.NewAdminHandler(s.adminHandler, s, s.params.Authorizer, s.params.AuthorizationConfig)

	adminThriftHandler := thrift.NewAdminHandler(s.adminHandler)
	adminThriftHandler.Register(s.GetDispatcher())

	adminGRPCHandler := grpc.NewAdminHandler(s.adminHandler)
	adminGRPCHandler.Register(s.GetDispatcher())

	// must start resource first
	s.Resource.Start()
	s.handler.Start()
	s.adminHandler.Start()

	// base (service is not started in frontend or admin handler) in case of race condition in yarpc registration function

	logger.Info("frontend started")

	<-s.stopC
}

// Stop stops the service
func (s *Service) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	// initiate graceful shutdown:
	// 1. Fail rpc health check, this will cause client side load balancer to stop forwarding requests to this node
	// 2. wait for failure detection time
	// 3. stop taking new requests by returning InternalServiceError
	// 4. Wait for a second
	// 5. Stop everything forcefully and return

	requestDrainTime := common.MinDuration(time.Second, s.config.ShutdownDrainDuration())
	failureDetectionTime := common.MaxDuration(0, s.config.ShutdownDrainDuration()-requestDrainTime)

	s.GetLogger().Info("ShutdownHandler: Updating rpc health status to ShuttingDown")
	s.handler.UpdateHealthStatus(api.HealthStatusShuttingDown)

	s.GetLogger().Info("ShutdownHandler: Waiting for others to discover I am unhealthy")
	time.Sleep(failureDetectionTime)

	s.handler.Stop()
	s.adminHandler.Stop()

	s.GetLogger().Info("ShutdownHandler: Draining traffic")
	time.Sleep(requestDrainTime)

	close(s.stopC)
	s.Resource.Stop()
	s.params.Logger.Info("frontend stopped")
}
