package service

import (
	"github.com/uber-common/bark"
	ringpop "github.com/uber/ringpop-go"
	tchannel "github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/metrics"
)

type (
	// Service is the interface which must be implemented by all the services
	Service interface {
		// GetHostName returns the name of host running the service
		GetHostName() string

		// Start starts the service
		Start(thriftService []thrift.TChanServer)

		// Stop stops the service
		Stop()

		GetLogger() bark.Logger

		GetMetricsClient() metrics.Client

		GetClientFactory() client.Factory

		GetMembershipMonitor() membership.Monitor

		GetHostInfo() *membership.HostInfo
	}

	// RingpopProvider provides a bootstrapped ringpop
	RingpopProvider interface {
		Ringpop(ch *tchannel.Channel) (*ringpop.Ringpop, error)
	}
)
