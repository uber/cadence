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

	"github.com/cristalhq/jwt/v3"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
)

type (
	factorySuite struct {
		suite.Suite
		logger log.Logger
	}
)

func TestFactorySuite(t *testing.T) {
	suite.Run(t, new(factorySuite))
}

func (s *factorySuite) SetupTest() {
	s.logger = loggerimpl.NewLoggerForTest(s.Suite)
}

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
				Algorithm: jwt.RS256.String(),
				PublicKey: "../../config/credentials/keytest.pub",
			},
			MaxJwtTTL: 12345,
		},
	}
}

func (s *factorySuite) TestFactoryNoopAuthorizer() {
	cfgOAuthVar := cfgOAuth()
	publicKey, _ := common.LoadRSAPublicKey(cfgOAuthVar.OAuthAuthorizer.JwtCredentials.PublicKey)
	var tests = []struct {
		cfg      config.Authorization
		expected Authorizer
		err      error
	}{
		{cfgNoop(), &nopAuthority{}, nil},
		{cfgOAuthVar, &oauthAuthority{authorizationCfg: cfgOAuthVar.OAuthAuthorizer, log: s.logger, publicKey: publicKey}, nil},
	}

	for _, test := range tests {
		authorizer, err := NewAuthorizer(test.cfg, s.logger, nil)
		s.Equal(authorizer, test.expected)
		s.Equal(err, test.err)
	}
}
