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

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package adminv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AdminAPIClient is the client API for AdminAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminAPIClient interface {
	// DescribeWorkflowExecution returns information about the internal states of workflow execution.
	DescribeWorkflowExecution(ctx context.Context, in *DescribeWorkflowExecutionRequest, opts ...grpc.CallOption) (*DescribeWorkflowExecutionResponse, error)
	// DescribeHistoryHost returns information about the internal states of a history host.
	DescribeHistoryHost(ctx context.Context, in *DescribeHistoryHostRequest, opts ...grpc.CallOption) (*DescribeHistoryHostResponse, error)
	// CloseShard closes shard.
	CloseShard(ctx context.Context, in *CloseShardRequest, opts ...grpc.CallOption) (*CloseShardResponse, error)
	// RemoveTask removes task.
	RemoveTask(ctx context.Context, in *RemoveTaskRequest, opts ...grpc.CallOption) (*RemoveTaskResponse, error)
	// ResetQueue resets queue.
	ResetQueue(ctx context.Context, in *ResetQueueRequest, opts ...grpc.CallOption) (*ResetQueueResponse, error)
	// DescribeQueue describes queue.
	DescribeQueue(ctx context.Context, in *DescribeQueueRequest, opts ...grpc.CallOption) (*DescribeQueueResponse, error)
	// Returns the raw history of specified workflow execution.
	// It fails with 'EntityNotExistError' if specified workflow execution in unknown to the service.
	// StartEventId defines the beginning of the event to fetch. The first event is inclusive.
	// EndEventId and EndEventVersion defines the end of the event to fetch. The end event is exclusive.
	GetWorkflowExecutionRawHistoryV2(ctx context.Context, in *GetWorkflowExecutionRawHistoryV2Request, opts ...grpc.CallOption) (*GetWorkflowExecutionRawHistoryV2Response, error)
	// GetReplicationMessages returns new replication tasks since the read level provided in the token.
	GetReplicationMessages(ctx context.Context, in *GetReplicationMessagesRequest, opts ...grpc.CallOption) (*GetReplicationMessagesResponse, error)
	// GetDLQReplicationMessages return replication messages based on DLQ info.
	GetDLQReplicationMessages(ctx context.Context, in *GetDLQReplicationMessagesRequest, opts ...grpc.CallOption) (*GetDLQReplicationMessagesResponse, error)
	// GetDomainReplicationMessages returns new domain replication tasks since last retrieved task id.
	GetDomainReplicationMessages(ctx context.Context, in *GetDomainReplicationMessagesRequest, opts ...grpc.CallOption) (*GetDomainReplicationMessagesResponse, error)
	// ReapplyEvents applies stale events to the current workflow and current run.
	ReapplyEvents(ctx context.Context, in *ReapplyEventsRequest, opts ...grpc.CallOption) (*ReapplyEventsResponse, error)
	// AddSearchAttribute whitelist search attribute in request.
	AddSearchAttribute(ctx context.Context, in *AddSearchAttributeRequest, opts ...grpc.CallOption) (*AddSearchAttributeResponse, error)
	// DescribeCluster returns information about Cadence cluster.
	DescribeCluster(ctx context.Context, in *DescribeClusterRequest, opts ...grpc.CallOption) (*DescribeClusterResponse, error)
	// ReadDLQMessages returns messages from DLQ.
	ReadDLQMessages(ctx context.Context, in *ReadDLQMessagesRequest, opts ...grpc.CallOption) (*ReadDLQMessagesResponse, error)
	// PurgeDLQMessages purges messages from DLQ.
	PurgeDLQMessages(ctx context.Context, in *PurgeDLQMessagesRequest, opts ...grpc.CallOption) (*PurgeDLQMessagesResponse, error)
	// MergeDLQMessages merges messages from DLQ.
	MergeDLQMessages(ctx context.Context, in *MergeDLQMessagesRequest, opts ...grpc.CallOption) (*MergeDLQMessagesResponse, error)
	// RefreshWorkflowTasks refreshes all tasks of a workflow.
	RefreshWorkflowTasks(ctx context.Context, in *RefreshWorkflowTasksRequest, opts ...grpc.CallOption) (*RefreshWorkflowTasksResponse, error)
	// ResendReplicationTasks requests replication tasks from remote cluster and apply tasks to current cluster.
	ResendReplicationTasks(ctx context.Context, in *ResendReplicationTasksRequest, opts ...grpc.CallOption) (*ResendReplicationTasksResponse, error)
}

type adminAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminAPIClient(cc grpc.ClientConnInterface) AdminAPIClient {
	return &adminAPIClient{cc}
}

func (c *adminAPIClient) DescribeWorkflowExecution(ctx context.Context, in *DescribeWorkflowExecutionRequest, opts ...grpc.CallOption) (*DescribeWorkflowExecutionResponse, error) {
	out := new(DescribeWorkflowExecutionResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/DescribeWorkflowExecution", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) DescribeHistoryHost(ctx context.Context, in *DescribeHistoryHostRequest, opts ...grpc.CallOption) (*DescribeHistoryHostResponse, error) {
	out := new(DescribeHistoryHostResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/DescribeHistoryHost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) CloseShard(ctx context.Context, in *CloseShardRequest, opts ...grpc.CallOption) (*CloseShardResponse, error) {
	out := new(CloseShardResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/CloseShard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) RemoveTask(ctx context.Context, in *RemoveTaskRequest, opts ...grpc.CallOption) (*RemoveTaskResponse, error) {
	out := new(RemoveTaskResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/RemoveTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) ResetQueue(ctx context.Context, in *ResetQueueRequest, opts ...grpc.CallOption) (*ResetQueueResponse, error) {
	out := new(ResetQueueResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/ResetQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) DescribeQueue(ctx context.Context, in *DescribeQueueRequest, opts ...grpc.CallOption) (*DescribeQueueResponse, error) {
	out := new(DescribeQueueResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/DescribeQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) GetWorkflowExecutionRawHistoryV2(ctx context.Context, in *GetWorkflowExecutionRawHistoryV2Request, opts ...grpc.CallOption) (*GetWorkflowExecutionRawHistoryV2Response, error) {
	out := new(GetWorkflowExecutionRawHistoryV2Response)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/GetWorkflowExecutionRawHistoryV2", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) GetReplicationMessages(ctx context.Context, in *GetReplicationMessagesRequest, opts ...grpc.CallOption) (*GetReplicationMessagesResponse, error) {
	out := new(GetReplicationMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/GetReplicationMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) GetDLQReplicationMessages(ctx context.Context, in *GetDLQReplicationMessagesRequest, opts ...grpc.CallOption) (*GetDLQReplicationMessagesResponse, error) {
	out := new(GetDLQReplicationMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/GetDLQReplicationMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) GetDomainReplicationMessages(ctx context.Context, in *GetDomainReplicationMessagesRequest, opts ...grpc.CallOption) (*GetDomainReplicationMessagesResponse, error) {
	out := new(GetDomainReplicationMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/GetDomainReplicationMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) ReapplyEvents(ctx context.Context, in *ReapplyEventsRequest, opts ...grpc.CallOption) (*ReapplyEventsResponse, error) {
	out := new(ReapplyEventsResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/ReapplyEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) AddSearchAttribute(ctx context.Context, in *AddSearchAttributeRequest, opts ...grpc.CallOption) (*AddSearchAttributeResponse, error) {
	out := new(AddSearchAttributeResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/AddSearchAttribute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) DescribeCluster(ctx context.Context, in *DescribeClusterRequest, opts ...grpc.CallOption) (*DescribeClusterResponse, error) {
	out := new(DescribeClusterResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/DescribeCluster", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) ReadDLQMessages(ctx context.Context, in *ReadDLQMessagesRequest, opts ...grpc.CallOption) (*ReadDLQMessagesResponse, error) {
	out := new(ReadDLQMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/ReadDLQMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) PurgeDLQMessages(ctx context.Context, in *PurgeDLQMessagesRequest, opts ...grpc.CallOption) (*PurgeDLQMessagesResponse, error) {
	out := new(PurgeDLQMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/PurgeDLQMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) MergeDLQMessages(ctx context.Context, in *MergeDLQMessagesRequest, opts ...grpc.CallOption) (*MergeDLQMessagesResponse, error) {
	out := new(MergeDLQMessagesResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/MergeDLQMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) RefreshWorkflowTasks(ctx context.Context, in *RefreshWorkflowTasksRequest, opts ...grpc.CallOption) (*RefreshWorkflowTasksResponse, error) {
	out := new(RefreshWorkflowTasksResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/RefreshWorkflowTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminAPIClient) ResendReplicationTasks(ctx context.Context, in *ResendReplicationTasksRequest, opts ...grpc.CallOption) (*ResendReplicationTasksResponse, error) {
	out := new(ResendReplicationTasksResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.admin.v1.AdminAPI/ResendReplicationTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminAPIServer is the server API for AdminAPI service.
// All implementations must embed UnimplementedAdminAPIServer
// for forward compatibility
type AdminAPIServer interface {
	// DescribeWorkflowExecution returns information about the internal states of workflow execution.
	DescribeWorkflowExecution(context.Context, *DescribeWorkflowExecutionRequest) (*DescribeWorkflowExecutionResponse, error)
	// DescribeHistoryHost returns information about the internal states of a history host.
	DescribeHistoryHost(context.Context, *DescribeHistoryHostRequest) (*DescribeHistoryHostResponse, error)
	// CloseShard closes shard.
	CloseShard(context.Context, *CloseShardRequest) (*CloseShardResponse, error)
	// RemoveTask removes task.
	RemoveTask(context.Context, *RemoveTaskRequest) (*RemoveTaskResponse, error)
	// ResetQueue resets queue.
	ResetQueue(context.Context, *ResetQueueRequest) (*ResetQueueResponse, error)
	// DescribeQueue describes queue.
	DescribeQueue(context.Context, *DescribeQueueRequest) (*DescribeQueueResponse, error)
	// Returns the raw history of specified workflow execution.
	// It fails with 'EntityNotExistError' if specified workflow execution in unknown to the service.
	// StartEventId defines the beginning of the event to fetch. The first event is inclusive.
	// EndEventId and EndEventVersion defines the end of the event to fetch. The end event is exclusive.
	GetWorkflowExecutionRawHistoryV2(context.Context, *GetWorkflowExecutionRawHistoryV2Request) (*GetWorkflowExecutionRawHistoryV2Response, error)
	// GetReplicationMessages returns new replication tasks since the read level provided in the token.
	GetReplicationMessages(context.Context, *GetReplicationMessagesRequest) (*GetReplicationMessagesResponse, error)
	// GetDLQReplicationMessages return replication messages based on DLQ info.
	GetDLQReplicationMessages(context.Context, *GetDLQReplicationMessagesRequest) (*GetDLQReplicationMessagesResponse, error)
	// GetDomainReplicationMessages returns new domain replication tasks since last retrieved task id.
	GetDomainReplicationMessages(context.Context, *GetDomainReplicationMessagesRequest) (*GetDomainReplicationMessagesResponse, error)
	// ReapplyEvents applies stale events to the current workflow and current run.
	ReapplyEvents(context.Context, *ReapplyEventsRequest) (*ReapplyEventsResponse, error)
	// AddSearchAttribute whitelist search attribute in request.
	AddSearchAttribute(context.Context, *AddSearchAttributeRequest) (*AddSearchAttributeResponse, error)
	// DescribeCluster returns information about Cadence cluster.
	DescribeCluster(context.Context, *DescribeClusterRequest) (*DescribeClusterResponse, error)
	// ReadDLQMessages returns messages from DLQ.
	ReadDLQMessages(context.Context, *ReadDLQMessagesRequest) (*ReadDLQMessagesResponse, error)
	// PurgeDLQMessages purges messages from DLQ.
	PurgeDLQMessages(context.Context, *PurgeDLQMessagesRequest) (*PurgeDLQMessagesResponse, error)
	// MergeDLQMessages merges messages from DLQ.
	MergeDLQMessages(context.Context, *MergeDLQMessagesRequest) (*MergeDLQMessagesResponse, error)
	// RefreshWorkflowTasks refreshes all tasks of a workflow.
	RefreshWorkflowTasks(context.Context, *RefreshWorkflowTasksRequest) (*RefreshWorkflowTasksResponse, error)
	// ResendReplicationTasks requests replication tasks from remote cluster and apply tasks to current cluster.
	ResendReplicationTasks(context.Context, *ResendReplicationTasksRequest) (*ResendReplicationTasksResponse, error)
	mustEmbedUnimplementedAdminAPIServer()
}

// UnimplementedAdminAPIServer must be embedded to have forward compatible implementations.
type UnimplementedAdminAPIServer struct {
}

func (UnimplementedAdminAPIServer) DescribeWorkflowExecution(context.Context, *DescribeWorkflowExecutionRequest) (*DescribeWorkflowExecutionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeWorkflowExecution not implemented")
}
func (UnimplementedAdminAPIServer) DescribeHistoryHost(context.Context, *DescribeHistoryHostRequest) (*DescribeHistoryHostResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeHistoryHost not implemented")
}
func (UnimplementedAdminAPIServer) CloseShard(context.Context, *CloseShardRequest) (*CloseShardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseShard not implemented")
}
func (UnimplementedAdminAPIServer) RemoveTask(context.Context, *RemoveTaskRequest) (*RemoveTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveTask not implemented")
}
func (UnimplementedAdminAPIServer) ResetQueue(context.Context, *ResetQueueRequest) (*ResetQueueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetQueue not implemented")
}
func (UnimplementedAdminAPIServer) DescribeQueue(context.Context, *DescribeQueueRequest) (*DescribeQueueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeQueue not implemented")
}
func (UnimplementedAdminAPIServer) GetWorkflowExecutionRawHistoryV2(context.Context, *GetWorkflowExecutionRawHistoryV2Request) (*GetWorkflowExecutionRawHistoryV2Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkflowExecutionRawHistoryV2 not implemented")
}
func (UnimplementedAdminAPIServer) GetReplicationMessages(context.Context, *GetReplicationMessagesRequest) (*GetReplicationMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReplicationMessages not implemented")
}
func (UnimplementedAdminAPIServer) GetDLQReplicationMessages(context.Context, *GetDLQReplicationMessagesRequest) (*GetDLQReplicationMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDLQReplicationMessages not implemented")
}
func (UnimplementedAdminAPIServer) GetDomainReplicationMessages(context.Context, *GetDomainReplicationMessagesRequest) (*GetDomainReplicationMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDomainReplicationMessages not implemented")
}
func (UnimplementedAdminAPIServer) ReapplyEvents(context.Context, *ReapplyEventsRequest) (*ReapplyEventsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReapplyEvents not implemented")
}
func (UnimplementedAdminAPIServer) AddSearchAttribute(context.Context, *AddSearchAttributeRequest) (*AddSearchAttributeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddSearchAttribute not implemented")
}
func (UnimplementedAdminAPIServer) DescribeCluster(context.Context, *DescribeClusterRequest) (*DescribeClusterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeCluster not implemented")
}
func (UnimplementedAdminAPIServer) ReadDLQMessages(context.Context, *ReadDLQMessagesRequest) (*ReadDLQMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadDLQMessages not implemented")
}
func (UnimplementedAdminAPIServer) PurgeDLQMessages(context.Context, *PurgeDLQMessagesRequest) (*PurgeDLQMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PurgeDLQMessages not implemented")
}
func (UnimplementedAdminAPIServer) MergeDLQMessages(context.Context, *MergeDLQMessagesRequest) (*MergeDLQMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MergeDLQMessages not implemented")
}
func (UnimplementedAdminAPIServer) RefreshWorkflowTasks(context.Context, *RefreshWorkflowTasksRequest) (*RefreshWorkflowTasksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshWorkflowTasks not implemented")
}
func (UnimplementedAdminAPIServer) ResendReplicationTasks(context.Context, *ResendReplicationTasksRequest) (*ResendReplicationTasksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResendReplicationTasks not implemented")
}
func (UnimplementedAdminAPIServer) mustEmbedUnimplementedAdminAPIServer() {}

// UnsafeAdminAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminAPIServer will
// result in compilation errors.
type UnsafeAdminAPIServer interface {
	mustEmbedUnimplementedAdminAPIServer()
}

func RegisterAdminAPIServer(s grpc.ServiceRegistrar, srv AdminAPIServer) {
	s.RegisterService(&AdminAPI_ServiceDesc, srv)
}

func _AdminAPI_DescribeWorkflowExecution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DescribeWorkflowExecutionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).DescribeWorkflowExecution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/DescribeWorkflowExecution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).DescribeWorkflowExecution(ctx, req.(*DescribeWorkflowExecutionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_DescribeHistoryHost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DescribeHistoryHostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).DescribeHistoryHost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/DescribeHistoryHost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).DescribeHistoryHost(ctx, req.(*DescribeHistoryHostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_CloseShard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseShardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).CloseShard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/CloseShard",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).CloseShard(ctx, req.(*CloseShardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_RemoveTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).RemoveTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/RemoveTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).RemoveTask(ctx, req.(*RemoveTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_ResetQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetQueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).ResetQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/ResetQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).ResetQueue(ctx, req.(*ResetQueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_DescribeQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DescribeQueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).DescribeQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/DescribeQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).DescribeQueue(ctx, req.(*DescribeQueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_GetWorkflowExecutionRawHistoryV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkflowExecutionRawHistoryV2Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).GetWorkflowExecutionRawHistoryV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/GetWorkflowExecutionRawHistoryV2",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).GetWorkflowExecutionRawHistoryV2(ctx, req.(*GetWorkflowExecutionRawHistoryV2Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_GetReplicationMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReplicationMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).GetReplicationMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/GetReplicationMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).GetReplicationMessages(ctx, req.(*GetReplicationMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_GetDLQReplicationMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDLQReplicationMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).GetDLQReplicationMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/GetDLQReplicationMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).GetDLQReplicationMessages(ctx, req.(*GetDLQReplicationMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_GetDomainReplicationMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDomainReplicationMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).GetDomainReplicationMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/GetDomainReplicationMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).GetDomainReplicationMessages(ctx, req.(*GetDomainReplicationMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_ReapplyEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReapplyEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).ReapplyEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/ReapplyEvents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).ReapplyEvents(ctx, req.(*ReapplyEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_AddSearchAttribute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddSearchAttributeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).AddSearchAttribute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/AddSearchAttribute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).AddSearchAttribute(ctx, req.(*AddSearchAttributeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_DescribeCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DescribeClusterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).DescribeCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/DescribeCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).DescribeCluster(ctx, req.(*DescribeClusterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_ReadDLQMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadDLQMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).ReadDLQMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/ReadDLQMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).ReadDLQMessages(ctx, req.(*ReadDLQMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_PurgeDLQMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PurgeDLQMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).PurgeDLQMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/PurgeDLQMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).PurgeDLQMessages(ctx, req.(*PurgeDLQMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_MergeDLQMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MergeDLQMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).MergeDLQMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/MergeDLQMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).MergeDLQMessages(ctx, req.(*MergeDLQMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_RefreshWorkflowTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshWorkflowTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).RefreshWorkflowTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/RefreshWorkflowTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).RefreshWorkflowTasks(ctx, req.(*RefreshWorkflowTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminAPI_ResendReplicationTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResendReplicationTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminAPIServer).ResendReplicationTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.admin.v1.AdminAPI/ResendReplicationTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminAPIServer).ResendReplicationTasks(ctx, req.(*ResendReplicationTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AdminAPI_ServiceDesc is the grpc.ServiceDesc for AdminAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AdminAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "uber.cadence.admin.v1.AdminAPI",
	HandlerType: (*AdminAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DescribeWorkflowExecution",
			Handler:    _AdminAPI_DescribeWorkflowExecution_Handler,
		},
		{
			MethodName: "DescribeHistoryHost",
			Handler:    _AdminAPI_DescribeHistoryHost_Handler,
		},
		{
			MethodName: "CloseShard",
			Handler:    _AdminAPI_CloseShard_Handler,
		},
		{
			MethodName: "RemoveTask",
			Handler:    _AdminAPI_RemoveTask_Handler,
		},
		{
			MethodName: "ResetQueue",
			Handler:    _AdminAPI_ResetQueue_Handler,
		},
		{
			MethodName: "DescribeQueue",
			Handler:    _AdminAPI_DescribeQueue_Handler,
		},
		{
			MethodName: "GetWorkflowExecutionRawHistoryV2",
			Handler:    _AdminAPI_GetWorkflowExecutionRawHistoryV2_Handler,
		},
		{
			MethodName: "GetReplicationMessages",
			Handler:    _AdminAPI_GetReplicationMessages_Handler,
		},
		{
			MethodName: "GetDLQReplicationMessages",
			Handler:    _AdminAPI_GetDLQReplicationMessages_Handler,
		},
		{
			MethodName: "GetDomainReplicationMessages",
			Handler:    _AdminAPI_GetDomainReplicationMessages_Handler,
		},
		{
			MethodName: "ReapplyEvents",
			Handler:    _AdminAPI_ReapplyEvents_Handler,
		},
		{
			MethodName: "AddSearchAttribute",
			Handler:    _AdminAPI_AddSearchAttribute_Handler,
		},
		{
			MethodName: "DescribeCluster",
			Handler:    _AdminAPI_DescribeCluster_Handler,
		},
		{
			MethodName: "ReadDLQMessages",
			Handler:    _AdminAPI_ReadDLQMessages_Handler,
		},
		{
			MethodName: "PurgeDLQMessages",
			Handler:    _AdminAPI_PurgeDLQMessages_Handler,
		},
		{
			MethodName: "MergeDLQMessages",
			Handler:    _AdminAPI_MergeDLQMessages_Handler,
		},
		{
			MethodName: "RefreshWorkflowTasks",
			Handler:    _AdminAPI_RefreshWorkflowTasks_Handler,
		},
		{
			MethodName: "ResendReplicationTasks",
			Handler:    _AdminAPI_ResendReplicationTasks_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "uber/cadence/admin/v1/service.proto",
}
