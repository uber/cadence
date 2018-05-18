// Code generated by thriftrw-plugin-yarpc
// @generated

package matchingservicefx

import (
	"github.com/uber/cadence/.gen/go/matching/matchingserviceclient"
	"go.uber.org/fx"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/encoding/thrift"
)

// Params defines the dependencies for the MatchingService client.
type Params struct {
	fx.In

	Provider yarpc.ClientConfig
}

// Result defines the output of the MatchingService client module. It provides a
// MatchingService client to an Fx application.
type Result struct {
	fx.Out

	Client matchingserviceclient.Interface

	// We are using an fx.Out struct here instead of just returning a client
	// so that we can add more values or add named versions of the client in
	// the future without breaking any existing code.
}

// Client provides a MatchingService client to an Fx application using the given name
// for routing.
//
// 	fx.Provide(
// 		matchingservicefx.Client("..."),
// 		newHandler,
// 	)
func Client(name string, opts ...thrift.ClientOption) interface{} {
	return func(p Params) Result {
		client := matchingserviceclient.New(p.Provider.ClientConfig(name), opts...)
		return Result{Client: client}
	}
}
