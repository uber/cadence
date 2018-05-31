// Code generated by thriftrw-plugin-yarpc
// @generated

// Package matchingservicefx provides better integration for Fx for services
// implementing or calling MatchingService.
//
// Clients
//
// If you are making requests to MatchingService, use the Client function to inject a
// MatchingService client into your container.
//
// 	fx.Provide(matchingservicefx.Client("..."))
//
// Servers
//
// If you are implementing MatchingService, provide a matchingserviceserver.Interface into
// the container and use the Server function.
//
// Given,
//
// 	func NewMatchingServiceHandler() matchingserviceserver.Interface
//
// You can do the following to have the procedures of MatchingService made available
// to an Fx application.
//
// 	fx.Provide(
// 		NewMatchingServiceHandler,
// 		matchingservicefx.Server(),
// 	)
package matchingservicefx
