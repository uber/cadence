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

package apiv1

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

// MetaAPIClient is the client API for MetaAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetaAPIClient interface {
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error)
}

type metaAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewMetaAPIClient(cc grpc.ClientConnInterface) MetaAPIClient {
	return &metaAPIClient{cc}
}

func (c *metaAPIClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error) {
	out := new(HealthResponse)
	err := c.cc.Invoke(ctx, "/uber.cadence.api.v1.MetaAPI/Health", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetaAPIServer is the server API for MetaAPI service.
// All implementations must embed UnimplementedMetaAPIServer
// for forward compatibility
type MetaAPIServer interface {
	Health(context.Context, *HealthRequest) (*HealthResponse, error)
	mustEmbedUnimplementedMetaAPIServer()
}

// UnimplementedMetaAPIServer must be embedded to have forward compatible implementations.
type UnimplementedMetaAPIServer struct {
}

func (UnimplementedMetaAPIServer) Health(context.Context, *HealthRequest) (*HealthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Health not implemented")
}
func (UnimplementedMetaAPIServer) mustEmbedUnimplementedMetaAPIServer() {}

// UnsafeMetaAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetaAPIServer will
// result in compilation errors.
type UnsafeMetaAPIServer interface {
	mustEmbedUnimplementedMetaAPIServer()
}

func RegisterMetaAPIServer(s grpc.ServiceRegistrar, srv MetaAPIServer) {
	s.RegisterService(&MetaAPI_ServiceDesc, srv)
}

func _MetaAPI_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetaAPIServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/uber.cadence.api.v1.MetaAPI/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetaAPIServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetaAPI_ServiceDesc is the grpc.ServiceDesc for MetaAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetaAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "uber.cadence.api.v1.MetaAPI",
	HandlerType: (*MetaAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _MetaAPI_Health_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "uber/cadence/api/v1/service_meta.proto",
}
