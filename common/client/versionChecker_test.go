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
	"github.com/uber/cadence/.gen/go/shared"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type (
	VersionCheckerSuite struct {
		*require.Assertions
		suite.Suite
	}
)

func TestVersionCheckerSuite(t *testing.T) {
	suite.Run(t, new(VersionCheckerSuite))
}

func (s *VersionCheckerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *VersionCheckerSuite) TestClientVersionSupported() {
	testCases := []struct{
		versionChecker VersionChecker
		callContext context.Context
		expectErr bool
	}{
		{
			versionChecker: NewVersionChecker(false),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: context.Background(),
			expectErr: false,
		},
		{
				versionChecker: NewVersionChecker(true),
				callContext: s.constructCallContext("unknown-client", "0.0.0"),
				expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(GoSDK, "malformed-version"),
			expectErr: true,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(GoSDK, SupportedGoSDKVersion),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(JavaSDK, SupportedJavaSDKVersion),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(CLI, SupportedCLIVersion),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(GoSDK, "999.999.999"),
			expectErr: true,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(JavaSDK, "999.999.999"),
			expectErr: true,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(CLI, "999.999.999"),
			expectErr: true,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(GoSDK, "1.4.0"),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(JavaSDK, "1.4.0"),
			expectErr: false,
		},
		{
			versionChecker: NewVersionChecker(true),
			callContext: s.constructCallContext(CLI, "1.4.0"),
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		err := tc.versionChecker.ClientVersionSupported(tc.callContext)
		if tc.expectErr {
			s.Error(err)
			s.IsType(&shared.ClientVersionNotSupportedError{}, err)
		} else {
			s.NoError(err)
		}
	}
}

func (s *VersionCheckerSuite) constructCallContext(clientImpl string, featureVersion string) context.Context {
	// TODO: this needs to be populated with yarpc headers
	return context.Background()

}