// The MIT License (MIT)
// 
// Copyright (c) 2017 Uber Technologies, Inc.
// 
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

package adminserviceserver

import (
	context "context"
	admin "github.com/uber/cadence/.gen/go/admin"
	shared "github.com/uber/cadence/.gen/go/shared"
	wire "go.uber.org/thriftrw/wire"
	transport "go.uber.org/yarpc/api/transport"
	thrift "go.uber.org/yarpc/encoding/thrift"
)

// Interface is the server-side interface for the AdminService service.
type Interface interface {
	AddSearchAttribute(
		ctx context.Context,
		Request *admin.AddSearchAttributeRequest,
	) error

	DescribeHistoryHost(
		ctx context.Context,
		Request *shared.DescribeHistoryHostRequest,
	) (*shared.DescribeHistoryHostResponse, error)

	DescribeWorkflowExecution(
		ctx context.Context,
		Request *admin.DescribeWorkflowExecutionRequest,
	) (*admin.DescribeWorkflowExecutionResponse, error)

	GetWorkflowExecutionRawHistory(
		ctx context.Context,
		GetRequest *admin.GetWorkflowExecutionRawHistoryRequest,
	) (*admin.GetWorkflowExecutionRawHistoryResponse, error)
}

// New prepares an implementation of the AdminService service for
// registration.
//
// 	handler := AdminServiceHandler{}
// 	dispatcher.Register(adminserviceserver.New(handler))
func New(impl Interface, opts ...thrift.RegisterOption) []transport.Procedure {
	h := handler{impl}
	service := thrift.Service{
		Name: "AdminService",
		Methods: []thrift.Method{

			thrift.Method{
				Name: "AddSearchAttribute",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.AddSearchAttribute),
				},
				Signature:    "AddSearchAttribute(Request *admin.AddSearchAttributeRequest)",
				ThriftModule: admin.ThriftModule,
			},

			thrift.Method{
				Name: "DescribeHistoryHost",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DescribeHistoryHost),
				},
				Signature:    "DescribeHistoryHost(Request *shared.DescribeHistoryHostRequest) (*shared.DescribeHistoryHostResponse)",
				ThriftModule: admin.ThriftModule,
			},

			thrift.Method{
				Name: "DescribeWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DescribeWorkflowExecution),
				},
				Signature:    "DescribeWorkflowExecution(Request *admin.DescribeWorkflowExecutionRequest) (*admin.DescribeWorkflowExecutionResponse)",
				ThriftModule: admin.ThriftModule,
			},

			thrift.Method{
				Name: "GetWorkflowExecutionRawHistory",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.GetWorkflowExecutionRawHistory),
				},
				Signature:    "GetWorkflowExecutionRawHistory(GetRequest *admin.GetWorkflowExecutionRawHistoryRequest) (*admin.GetWorkflowExecutionRawHistoryResponse)",
				ThriftModule: admin.ThriftModule,
			},
		},
	}

	procedures := make([]transport.Procedure, 0, 4)
	procedures = append(procedures, thrift.BuildProcedures(service, opts...)...)
	return procedures
}

type handler struct{ impl Interface }

func (h handler) AddSearchAttribute(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args admin.AdminService_AddSearchAttribute_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.AddSearchAttribute(ctx, args.Request)

	hadError := err != nil
	result, err := admin.AdminService_AddSearchAttribute_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) DescribeHistoryHost(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args admin.AdminService_DescribeHistoryHost_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.DescribeHistoryHost(ctx, args.Request)

	hadError := err != nil
	result, err := admin.AdminService_DescribeHistoryHost_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) DescribeWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args admin.AdminService_DescribeWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.DescribeWorkflowExecution(ctx, args.Request)

	hadError := err != nil
	result, err := admin.AdminService_DescribeWorkflowExecution_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) GetWorkflowExecutionRawHistory(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args admin.AdminService_GetWorkflowExecutionRawHistory_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.GetWorkflowExecutionRawHistory(ctx, args.GetRequest)

	hadError := err != nil
	result, err := admin.AdminService_GetWorkflowExecutionRawHistory_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}
