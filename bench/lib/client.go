// Copyright (c) 2017-2021 Uber Technologies Inc.

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

package lib

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
)

const workflowRetentionDays = 1

// CadenceClient is an abstraction on top of
// the cadence library client that serves as
// a union of all the client interfaces that
// the library exposes
type CadenceClient struct {
	client.Client
	// domainClient only exposes domain API
	client.DomainClient
	// this is the service needed to start the workers
	Service workflowserviceclient.Interface
}

// CreateDomain creates a cadence domain with the given name and description
// if the domain already exist, this method silently returns success
func (client CadenceClient) CreateDomain(name string, desc string, owner string) error {
	emitMetric := true
	isGlobalDomain := false
	retention := int32(workflowRetentionDays)
	req := &shared.RegisterDomainRequest{
		Name:                                   &name,
		Description:                            &desc,
		OwnerEmail:                             &owner,
		WorkflowExecutionRetentionPeriodInDays: &retention,
		EmitMetric:                             &emitMetric,
		IsGlobalDomain:                         &isGlobalDomain,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := client.Register(ctx, req)
	if err != nil {
		if _, ok := err.(*shared.DomainAlreadyExistsError); !ok {
			return err
		}
	}
	return nil
}

// NewCadenceClients builds a CadenceClient for each domain from the runtimeContext
func NewCadenceClients(runtime *RuntimeContext) (map[string]CadenceClient, error) {
	cadenceClients := make(map[string]CadenceClient)
	for _, domain := range runtime.Bench.Domains {
		client, err := NewCadenceClientForDomain(runtime, domain)
		if err != nil {
			return nil, err
		}

		cadenceClients[domain] = client
	}

	return cadenceClients, nil
}

// NewCadenceClientForDomain builds a CadenceClient for a specified domain based on runtimeContext
func NewCadenceClientForDomain(
	runtime *RuntimeContext,
	domain string,
) (CadenceClient, error) {

	transport := grpc.NewTransport(
		grpc.ServiceName(runtime.Bench.Name),
	)

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: runtime.Bench.Name,
		Outbounds: yarpc.Outbounds{
			runtime.Cadence.ServiceName: {Unary: transport.NewSingleOutbound(runtime.Cadence.HostNameAndPort)},
		},
	})

	if err := dispatcher.Start(); err != nil {
		dispatcher.Stop()
		return CadenceClient{}, fmt.Errorf("failed to create outbound transport channel: %v", err)
	}

	var cadenceClient CadenceClient
	cadenceClient.Service = workflowserviceclient.New(dispatcher.ClientConfig(runtime.Cadence.ServiceName))
	cadenceClient.Client = client.NewClient(
		cadenceClient.Service,
		domain,
		&client.Options{
			MetricsScope: runtime.Metrics,
		},
	)
	cadenceClient.DomainClient = client.NewDomainClient(
		cadenceClient.Service,
		&client.Options{
			MetricsScope: runtime.Metrics,
		},
	)

	return cadenceClient, nil
}
