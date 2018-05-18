// Code generated by thriftrw-plugin-yarpc
// @generated

package adminserviceclient

import (
	"context"
	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/encoding/thrift"
	"reflect"
)

// Interface is a client for the AdminService service.
type Interface interface {
	InquiryWorkflowExecution(
		ctx context.Context,
		InquiryRequest *shared.DescribeWorkflowExecutionRequest,
		opts ...yarpc.CallOption,
	) (*admin.InquiryWorkflowExecutionResponse, error)
}

// New builds a new client for the AdminService service.
//
// 	client := adminserviceclient.New(dispatcher.ClientConfig("adminservice"))
func New(c transport.ClientConfig, opts ...thrift.ClientOption) Interface {
	return client{
		c: thrift.New(thrift.Config{
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
	c thrift.Client
}

func (c client) InquiryWorkflowExecution(
	ctx context.Context,
	_InquiryRequest *shared.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (success *admin.InquiryWorkflowExecutionResponse, err error) {

	args := admin.AdminService_InquiryWorkflowExecution_Helper.Args(_InquiryRequest)

	var body wire.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result admin.AdminService_InquiryWorkflowExecution_Result
	if err = result.FromWire(body); err != nil {
		return
	}

	success, err = admin.AdminService_InquiryWorkflowExecution_Helper.UnwrapResponse(&result)
	return
}
