// Copyright (c) 2017 Uber Technologies, Inc.
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

package client

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"go.uber.org/yarpc"
)

const (
	// GoSDK is the header value for common.ClientImplHeaderName indicating a go sdk client
	GoSDK = "uber-go"
	// JavaSDK is the header value for common.ClientImplHeaderName indicating a java sdk client
	JavaSDK = "uber-java"
	// CLI is the header value for common.ClientImplHeaderName indicating a cli client
	CLI = "cli"

	// SupportedGoSDKVersion indicates the highest go sdk version server will accept requests from
	SupportedGoSDKVersion = "<1.5.0"
	// SupportedJavaSDKVersion indicates the highest java sdk version server will accept requests from
	SupportedJavaSDKVersion = "<1.5.0"
	// SupportedCLIVersion indicates the highest cli version server will accept requests from
	SupportedCLIVersion = "<1.5.0"

	// GoWorkerStickyQueryVersion indicates the minimum client version of go worker which supports StickyQuery
	GoWorkerStickQueryVersion = ">=1.0.0"
	// JavaWorkerStickyQueryVersion indicates the minimum client version of the java worker which supports StickyQuery
	JavaWorkerStickQueryVersion = ">=1.0.0"
	// GoWorkerConsistentQueryVersion indicates the minimum client version of the go worker which supports ConsistentQuery
	GoWorkerConsistentQueryVersion = ">=1.5.0"

	stickyQueryVersion = "stick-query-version"
	consistentQueryVersion = "consistent-query-version"
)

type (
	// VersionChecker is used to check client/server compatibility and client's capabilities
	VersionChecker interface {
		ClientVersionSupported(context.Context) error

		SupportsStickyQuery(context.Context) error
		SupportsConsistentQuery(context.Context) error
	}

	versionChecker struct {
		featureVersions map[string]map[string]version.Constraints
		supportedVersions map[string]version.Constraints
		checkVersion bool
	}
)

// NewVersionChecker constructs a new VersionChecker
func NewVersionChecker(checkVersion bool) VersionChecker {
	requiredFeatureVersions := map[string]map[string]version.Constraints{
		GoSDK: {
			stickyQueryVersion:     mustNewConstraint(GoWorkerStickQueryVersion),
			consistentQueryVersion: mustNewConstraint(GoWorkerConsistentQueryVersion),
		},
		JavaSDK: {
			stickyQueryVersion: mustNewConstraint(JavaWorkerStickQueryVersion),
		},
	}
	supportedVersions := map[string]version.Constraints{
		GoSDK: mustNewConstraint(SupportedGoSDKVersion),
		JavaSDK: mustNewConstraint(SupportedJavaSDKVersion),
		CLI: mustNewConstraint(SupportedCLIVersion),
	}
	return &versionChecker{
		featureVersions: requiredFeatureVersions,
		supportedVersions: supportedVersions,
		checkVersion: checkVersion,
	}
}

func (vc *versionChecker) ClientVersionSupported(ctx context.Context) error {
	if !vc.checkVersion {
		return nil
	}

	call := yarpc.CallFromContext(ctx)
	clientFeatureVersion := call.Header(common.FeatureVersionHeaderName)
	clientImpl := call.Header(common.ClientImplHeaderName)

	if clientFeatureVersion == "" {
		return nil
	}
	versionSupported, ok := vc.supportedVersions[clientImpl]
	if !ok {
		return nil
	}
	clientVersion, err := version.NewVersion(clientFeatureVersion)
	if clientFeatureVersion == "malformed-version" {
		fmt.Println(err)
		fmt.Println(clientVersion)
	}
	if err != nil {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl, SupportedVersions: versionSupported.String()}
	}
	if !versionSupported.Check(clientVersion) {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl, SupportedVersions: versionSupported.String()}
	}
	return nil
}

// SupportsStickyQuery returns error if sticky query is not supported otherwise nil
func (vc *versionChecker) SupportsStickyQuery(ctx context.Context) error {
	return vc.checkFeature(ctx, stickyQueryVersion)
}

// SupportConsistentQuery returns error if consistent query is not supported otherwise nil
func (vc *versionChecker) SupportsConsistentQuery(ctx context.Context) error {
	return vc.checkFeature(ctx, consistentQueryVersion)
}

func (vc *versionChecker) checkFeature(ctx context.Context, feature string) error {
	call := yarpc.CallFromContext(ctx)
	clientFeatureVersion := call.Header(common.FeatureVersionHeaderName)
	clientImpl := call.Header(common.ClientImplHeaderName)

	if clientFeatureVersion == "" || clientImpl == "" {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl}
	}
	versionSupported, ok := vc.featureVersions[clientImpl][feature]
	if !ok {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl}
	}
	workerVersion, err := version.NewVersion(clientFeatureVersion)
	if err != nil {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl, SupportedVersions: versionSupported.String()}
	}
	if !versionSupported.Check(workerVersion) {
		return &shared.ClientVersionNotSupportedError{FeatureVersion: clientFeatureVersion, ClientImpl: clientImpl, SupportedVersions: versionSupported.String()}
	}
	return nil
}

func mustNewConstraint(v string) version.Constraints {
	constraint, err := version.NewConstraint(v)
	if err != nil {
		panic("invalid version constraint " + v)
	}
	return constraint
}
