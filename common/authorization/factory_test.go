// Copyright (c) 2021 Uber Technologies, Inc.
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

package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
)

func cfgNoop() config.Authorization {
	return config.Authorization{
		OAuthAuthorizer: config.OAuthAuthorizer{
			Enable: false,
		},
		NoopAuthorizer: config.NoopAuthorizer{
			Enable: true,
		},
	}
}

func cfgOAuth() config.Authorization {
	return config.Authorization{
		OAuthAuthorizer: config.OAuthAuthorizer{
			Enable: true,
			JwtCredentials: config.JwtCredentials{
				Algorithm:  "RS256",
				PublicKey:  "public",
				PrivateKey: "private",
			},
			JwtTTL: 12345,
		},
	}
}

func TestFactoryNoopAuthorizer(t *testing.T) {
	cfgOAuthVar := cfgOAuth()

	var tests = []struct {
		cfg      config.Authorization
		expected Authorizer
	}{
		{cfgNoop(), &nopAuthority{}},
		{cfgOAuthVar, &oauthAuthority{authorizationCfg: cfgOAuthVar.OAuthAuthorizer}},
	}

	for _, test := range tests {
		authorizer := NewAuthorizer(test.cfg)
		assert.Equal(t, authorizer, test.expected)
	}
}
