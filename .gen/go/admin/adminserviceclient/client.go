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

package adminserviceclient

import (
	context "context"
	reflect "reflect"

	admin "github.com/uber/cadence/.gen/go/admin"
	replicator "github.com/uber/cadence/.gen/go/replicator"
	shared "github.com/uber/cadence/.gen/go/shared"
	wire "go.uber.org/thriftrw/wire"
	yarpc "go.uber.org/yarpc"
	transport "go.uber.org/yarpc/api/transport"
	thrift "go.uber.org/yarpc/encoding/thrift"
)

// Interface is a client for the AdminService service.
type Interface interface {
	AddSearchAttribute(
		ctx context.Context,
		Request *admin.AddSearchAttributeRequest,
		opts ...yarpc.CallOption,
	) error

	CloseShard(
		ctx context.Context,
		Request *shared.CloseShardRequest,
		opts ...yarpc.CallOption,
	) error

	DeleteWorkflow(
		ctx context.Context,
		Request *admin.AdminDeleteWorkflowRequest,
		opts ...yarpc.CallOption,
	) (*admin.AdminDeleteWorkflowResponse, error)

	DescribeCluster(
		ctx context.Context,
		opts ...yarpc.CallOption,
	) (*admin.DescribeClusterResponse, error)

	DescribeHistoryHost(
		ctx context.Context,
		Request *shared.DescribeHistoryHostRequest,
		opts ...yarpc.CallOption,
	) (*shared.DescribeHistoryHostResponse, error)

	DescribeQueue(
		ctx context.Context,
		Request *shared.DescribeQueueRequest,
		opts ...yarpc.CallOption,
	) (*shared.DescribeQueueResponse, error)

	DescribeShardDistribution(
		ctx context.Context,
		Request *shared.DescribeShardDistributionRequest,
		opts ...yarpc.CallOption,
	) (*shared.DescribeShardDistributionResponse, error)

	DescribeWorkflowExecution(
		ctx context.Context,
		Request *admin.DescribeWorkflowExecutionRequest,
		opts ...yarpc.CallOption,
	) (*admin.DescribeWorkflowExecutionResponse, error)

	GetCrossClusterTasks(
		ctx context.Context,
		Request *shared.GetCrossClusterTasksRequest,
		opts ...yarpc.CallOption,
	) (*shared.GetCrossClusterTasksResponse, error)

	GetDLQReplicationMessages(
		ctx context.Context,
		Request *replicator.GetDLQReplicationMessagesRequest,
		opts ...yarpc.CallOption,
	) (*replicator.GetDLQReplicationMessagesResponse, error)

	GetDomainAsyncWorkflowConfiguraton(
		ctx context.Context,
		Request *admin.GetDomainAsyncWorkflowConfiguratonRequest,
		opts ...yarpc.CallOption,
	) (*admin.GetDomainAsyncWorkflowConfiguratonResponse, error)

	GetDomainIsolationGroups(
		ctx context.Context,
		Request *admin.GetDomainIsolationGroupsRequest,
		opts ...yarpc.CallOption,
	) (*admin.GetDomainIsolationGroupsResponse, error)

	GetDomainReplicationMessages(
		ctx context.Context,
		Request *replicator.GetDomainReplicationMessagesRequest,
		opts ...yarpc.CallOption,
	) (*replicator.GetDomainReplicationMessagesResponse, error)

	GetDynamicConfig(
		ctx context.Context,
		Request *admin.GetDynamicConfigRequest,
		opts ...yarpc.CallOption,
	) (*admin.GetDynamicConfigResponse, error)

	GetGlobalIsolationGroups(
		ctx context.Context,
		Request *admin.GetGlobalIsolationGroupsRequest,
		opts ...yarpc.CallOption,
	) (*admin.GetGlobalIsolationGroupsResponse, error)

	GetReplicationMessages(
		ctx context.Context,
		Request *replicator.GetReplicationMessagesRequest,
		opts ...yarpc.CallOption,
	) (*replicator.GetReplicationMessagesResponse, error)

	GetWorkflowExecutionRawHistoryV2(
		ctx context.Context,
		GetRequest *admin.GetWorkflowExecutionRawHistoryV2Request,
		opts ...yarpc.CallOption,
	) (*admin.GetWorkflowExecutionRawHistoryV2Response, error)

	ListDynamicConfig(
		ctx context.Context,
		Request *admin.ListDynamicConfigRequest,
		opts ...yarpc.CallOption,
	) (*admin.ListDynamicConfigResponse, error)

	MaintainCorruptWorkflow(
		ctx context.Context,
		Request *admin.AdminMaintainWorkflowRequest,
		opts ...yarpc.CallOption,
	) (*admin.AdminMaintainWorkflowResponse, error)

	MergeDLQMessages(
		ctx context.Context,
		Request *replicator.MergeDLQMessagesRequest,
		opts ...yarpc.CallOption,
	) (*replicator.MergeDLQMessagesResponse, error)

	PurgeDLQMessages(
		ctx context.Context,
		Request *replicator.PurgeDLQMessagesRequest,
		opts ...yarpc.CallOption,
	) error

	ReadDLQMessages(
		ctx context.Context,
		Request *replicator.ReadDLQMessagesRequest,
		opts ...yarpc.CallOption,
	) (*replicator.ReadDLQMessagesResponse, error)

	ReapplyEvents(
		ctx context.Context,
		ReapplyEventsRequest *shared.ReapplyEventsRequest,
		opts ...yarpc.CallOption,
	) error

	RefreshWorkflowTasks(
		ctx context.Context,
		Request *shared.RefreshWorkflowTasksRequest,
		opts ...yarpc.CallOption,
	) error

	RemoveTask(
		ctx context.Context,
		Request *shared.RemoveTaskRequest,
		opts ...yarpc.CallOption,
	) error

	ResendReplicationTasks(
		ctx context.Context,
		Request *admin.ResendReplicationTasksRequest,
		opts ...yarpc.CallOption,
	) error

	ResetQueue(
		ctx context.Context,
		Request *shared.ResetQueueRequest,
		opts ...yarpc.CallOption,
	) error

	RespondCrossClusterTasksCompleted(
		ctx context.Context,
		Request *shared.RespondCrossClusterTasksCompletedRequest,
		opts ...yarpc.CallOption,
	) (*shared.RespondCrossClusterTasksCompletedResponse, error)

	RestoreDynamicConfig(
		ctx context.Context,
		Request *admin.RestoreDynamicConfigRequest,
		opts ...yarpc.CallOption,
	) error

	UpdateDomainAsyncWorkflowConfiguraton(
		ctx context.Context,
		Request *admin.UpdateDomainAsyncWorkflowConfiguratonRequest,
		opts ...yarpc.CallOption,
	) (*admin.UpdateDomainAsyncWorkflowConfiguratonResponse, error)

	UpdateDomainIsolationGroups(
		ctx context.Context,
		Request *admin.UpdateDomainIsolationGroupsRequest,
		opts ...yarpc.CallOption,
	) (*admin.UpdateDomainIsolationGroupsResponse, error)

	UpdateDynamicConfig(
		ctx context.Context,
		Request *admin.UpdateDynamicConfigRequest,
		opts ...yarpc.CallOption,
	) error

	UpdateGlobalIsolationGroups(
		ctx context.Context,
		Request *admin.UpdateGlobalIsolationGroupsRequest,
		opts ...yarpc.CallOption,
	) (*admin.UpdateGlobalIsolationGroupsResponse, error)
}

// New builds a new client for the AdminService service.
//
//	client := adminserviceclient.New(dispatcher.ClientConfig("adminservice"))
func New(c transport.ClientConfig, opts ...thrift.ClientOption) Interface {
	return client{
		c: thrift.New(thrift.Config{
			Service:      "AdminService",
			ClientConfig: c,
		}, opts...),
		nwc: thrift.NewNoWire(thrift.Config{
			Service:      "AdminService",
			ClientConfig: c,
		}, opts...),
	}
}

func init() {
	yarpc.RegisterClientBuilder(
		func(c transport.ClientConfig, f reflect.StructField) Interface {
			return New(c, thrift.ClientBuilderOptions(c, f)...)
		},
	)
}

type client struct {
	c   thrift.Client
	nwc thrift.NoWireClient
}

func (c client) AddSearchAttribute(
	ctx context.Context,
	_Request *admin.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_AddSearchAttribute_Result
	args := admin.AdminService_AddSearchAttribute_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_AddSearchAttribute_Helper.UnwrapResponse(&result)
	return
}

func (c client) CloseShard(
	ctx context.Context,
	_Request *shared.CloseShardRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_CloseShard_Result
	args := admin.AdminService_CloseShard_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_CloseShard_Helper.UnwrapResponse(&result)
	return
}

func (c client) DeleteWorkflow(
	ctx context.Context,
	_Request *admin.AdminDeleteWorkflowRequest,
	opts ...yarpc.CallOption,
) (success *admin.AdminDeleteWorkflowResponse, err error) {

	var result admin.AdminService_DeleteWorkflow_Result
	args := admin.AdminService_DeleteWorkflow_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DeleteWorkflow_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (success *admin.DescribeClusterResponse, err error) {

	var result admin.AdminService_DescribeCluster_Result
	args := admin.AdminService_DescribeCluster_Helper.Args()

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DescribeCluster_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeHistoryHost(
	ctx context.Context,
	_Request *shared.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeHistoryHostResponse, err error) {

	var result admin.AdminService_DescribeHistoryHost_Result
	args := admin.AdminService_DescribeHistoryHost_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DescribeHistoryHost_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeQueue(
	ctx context.Context,
	_Request *shared.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeQueueResponse, err error) {

	var result admin.AdminService_DescribeQueue_Result
	args := admin.AdminService_DescribeQueue_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DescribeQueue_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeShardDistribution(
	ctx context.Context,
	_Request *shared.DescribeShardDistributionRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeShardDistributionResponse, err error) {

	var result admin.AdminService_DescribeShardDistribution_Result
	args := admin.AdminService_DescribeShardDistribution_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DescribeShardDistribution_Helper.UnwrapResponse(&result)
	return
}

func (c client) DescribeWorkflowExecution(
	ctx context.Context,
	_Request *admin.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (success *admin.DescribeWorkflowExecutionResponse, err error) {

	var result admin.AdminService_DescribeWorkflowExecution_Result
	args := admin.AdminService_DescribeWorkflowExecution_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_DescribeWorkflowExecution_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetCrossClusterTasks(
	ctx context.Context,
	_Request *shared.GetCrossClusterTasksRequest,
	opts ...yarpc.CallOption,
) (success *shared.GetCrossClusterTasksResponse, err error) {

	var result admin.AdminService_GetCrossClusterTasks_Result
	args := admin.AdminService_GetCrossClusterTasks_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetCrossClusterTasks_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetDLQReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetDLQReplicationMessagesResponse, err error) {

	var result admin.AdminService_GetDLQReplicationMessages_Result
	args := admin.AdminService_GetDLQReplicationMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetDLQReplicationMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetDomainAsyncWorkflowConfiguraton(
	ctx context.Context,
	_Request *admin.GetDomainAsyncWorkflowConfiguratonRequest,
	opts ...yarpc.CallOption,
) (success *admin.GetDomainAsyncWorkflowConfiguratonResponse, err error) {

	var result admin.AdminService_GetDomainAsyncWorkflowConfiguraton_Result
	args := admin.AdminService_GetDomainAsyncWorkflowConfiguraton_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetDomainAsyncWorkflowConfiguraton_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetDomainIsolationGroups(
	ctx context.Context,
	_Request *admin.GetDomainIsolationGroupsRequest,
	opts ...yarpc.CallOption,
) (success *admin.GetDomainIsolationGroupsResponse, err error) {

	var result admin.AdminService_GetDomainIsolationGroups_Result
	args := admin.AdminService_GetDomainIsolationGroups_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetDomainIsolationGroups_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetDomainReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetDomainReplicationMessagesResponse, err error) {

	var result admin.AdminService_GetDomainReplicationMessages_Result
	args := admin.AdminService_GetDomainReplicationMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetDomainReplicationMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetDynamicConfig(
	ctx context.Context,
	_Request *admin.GetDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (success *admin.GetDynamicConfigResponse, err error) {

	var result admin.AdminService_GetDynamicConfig_Result
	args := admin.AdminService_GetDynamicConfig_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetDynamicConfig_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetGlobalIsolationGroups(
	ctx context.Context,
	_Request *admin.GetGlobalIsolationGroupsRequest,
	opts ...yarpc.CallOption,
) (success *admin.GetGlobalIsolationGroupsResponse, err error) {

	var result admin.AdminService_GetGlobalIsolationGroups_Result
	args := admin.AdminService_GetGlobalIsolationGroups_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetGlobalIsolationGroups_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetReplicationMessagesResponse, err error) {

	var result admin.AdminService_GetReplicationMessages_Result
	args := admin.AdminService_GetReplicationMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetReplicationMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	_GetRequest *admin.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (success *admin.GetWorkflowExecutionRawHistoryV2Response, err error) {

	var result admin.AdminService_GetWorkflowExecutionRawHistoryV2_Result
	args := admin.AdminService_GetWorkflowExecutionRawHistoryV2_Helper.Args(_GetRequest)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_GetWorkflowExecutionRawHistoryV2_Helper.UnwrapResponse(&result)
	return
}

func (c client) ListDynamicConfig(
	ctx context.Context,
	_Request *admin.ListDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (success *admin.ListDynamicConfigResponse, err error) {

	var result admin.AdminService_ListDynamicConfig_Result
	args := admin.AdminService_ListDynamicConfig_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_ListDynamicConfig_Helper.UnwrapResponse(&result)
	return
}

func (c client) MaintainCorruptWorkflow(
	ctx context.Context,
	_Request *admin.AdminMaintainWorkflowRequest,
	opts ...yarpc.CallOption,
) (success *admin.AdminMaintainWorkflowResponse, err error) {

	var result admin.AdminService_MaintainCorruptWorkflow_Result
	args := admin.AdminService_MaintainCorruptWorkflow_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_MaintainCorruptWorkflow_Helper.UnwrapResponse(&result)
	return
}

func (c client) MergeDLQMessages(
	ctx context.Context,
	_Request *replicator.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.MergeDLQMessagesResponse, err error) {

	var result admin.AdminService_MergeDLQMessages_Result
	args := admin.AdminService_MergeDLQMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_MergeDLQMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) PurgeDLQMessages(
	ctx context.Context,
	_Request *replicator.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_PurgeDLQMessages_Result
	args := admin.AdminService_PurgeDLQMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_PurgeDLQMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) ReadDLQMessages(
	ctx context.Context,
	_Request *replicator.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.ReadDLQMessagesResponse, err error) {

	var result admin.AdminService_ReadDLQMessages_Result
	args := admin.AdminService_ReadDLQMessages_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_ReadDLQMessages_Helper.UnwrapResponse(&result)
	return
}

func (c client) ReapplyEvents(
	ctx context.Context,
	_ReapplyEventsRequest *shared.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_ReapplyEvents_Result
	args := admin.AdminService_ReapplyEvents_Helper.Args(_ReapplyEventsRequest)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_ReapplyEvents_Helper.UnwrapResponse(&result)
	return
}

func (c client) RefreshWorkflowTasks(
	ctx context.Context,
	_Request *shared.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_RefreshWorkflowTasks_Result
	args := admin.AdminService_RefreshWorkflowTasks_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_RefreshWorkflowTasks_Helper.UnwrapResponse(&result)
	return
}

func (c client) RemoveTask(
	ctx context.Context,
	_Request *shared.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_RemoveTask_Result
	args := admin.AdminService_RemoveTask_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_RemoveTask_Helper.UnwrapResponse(&result)
	return
}

func (c client) ResendReplicationTasks(
	ctx context.Context,
	_Request *admin.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_ResendReplicationTasks_Result
	args := admin.AdminService_ResendReplicationTasks_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_ResendReplicationTasks_Helper.UnwrapResponse(&result)
	return
}

func (c client) ResetQueue(
	ctx context.Context,
	_Request *shared.ResetQueueRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_ResetQueue_Result
	args := admin.AdminService_ResetQueue_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_ResetQueue_Helper.UnwrapResponse(&result)
	return
}

func (c client) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	_Request *shared.RespondCrossClusterTasksCompletedRequest,
	opts ...yarpc.CallOption,
) (success *shared.RespondCrossClusterTasksCompletedResponse, err error) {

	var result admin.AdminService_RespondCrossClusterTasksCompleted_Result
	args := admin.AdminService_RespondCrossClusterTasksCompleted_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_RespondCrossClusterTasksCompleted_Helper.UnwrapResponse(&result)
	return
}

func (c client) RestoreDynamicConfig(
	ctx context.Context,
	_Request *admin.RestoreDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_RestoreDynamicConfig_Result
	args := admin.AdminService_RestoreDynamicConfig_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_RestoreDynamicConfig_Helper.UnwrapResponse(&result)
	return
}

func (c client) UpdateDomainAsyncWorkflowConfiguraton(
	ctx context.Context,
	_Request *admin.UpdateDomainAsyncWorkflowConfiguratonRequest,
	opts ...yarpc.CallOption,
) (success *admin.UpdateDomainAsyncWorkflowConfiguratonResponse, err error) {

	var result admin.AdminService_UpdateDomainAsyncWorkflowConfiguraton_Result
	args := admin.AdminService_UpdateDomainAsyncWorkflowConfiguraton_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_UpdateDomainAsyncWorkflowConfiguraton_Helper.UnwrapResponse(&result)
	return
}

func (c client) UpdateDomainIsolationGroups(
	ctx context.Context,
	_Request *admin.UpdateDomainIsolationGroupsRequest,
	opts ...yarpc.CallOption,
) (success *admin.UpdateDomainIsolationGroupsResponse, err error) {

	var result admin.AdminService_UpdateDomainIsolationGroups_Result
	args := admin.AdminService_UpdateDomainIsolationGroups_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_UpdateDomainIsolationGroups_Helper.UnwrapResponse(&result)
	return
}

func (c client) UpdateDynamicConfig(
	ctx context.Context,
	_Request *admin.UpdateDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (err error) {

	var result admin.AdminService_UpdateDynamicConfig_Result
	args := admin.AdminService_UpdateDynamicConfig_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	err = admin.AdminService_UpdateDynamicConfig_Helper.UnwrapResponse(&result)
	return
}

func (c client) UpdateGlobalIsolationGroups(
	ctx context.Context,
	_Request *admin.UpdateGlobalIsolationGroupsRequest,
	opts ...yarpc.CallOption,
) (success *admin.UpdateGlobalIsolationGroupsResponse, err error) {

	var result admin.AdminService_UpdateGlobalIsolationGroups_Result
	args := admin.AdminService_UpdateGlobalIsolationGroups_Helper.Args(_Request)

	if c.nwc != nil && c.nwc.Enabled() {
		if err = c.nwc.Call(ctx, args, &result, opts...); err != nil {
			return
		}
	} else {
		var body wire.Value
		if body, err = c.c.Call(ctx, args, opts...); err != nil {
			return
		}

		if err = result.FromWire(body); err != nil {
			return
		}
	}

	success, err = admin.AdminService_UpdateGlobalIsolationGroups_Helper.UnwrapResponse(&result)
	return
}
