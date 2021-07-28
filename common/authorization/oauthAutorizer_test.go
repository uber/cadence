// Copyright (c) 2019 Uber Technologies, Inc.
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
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc/api/encoding"
	"go.uber.org/yarpc/api/transport"
	"golang.org/x/net/context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type (
	oauthSuite struct {
		suite.Suite
		logger          *log.MockLogger
		cfg             config.OAuthAuthorizer
		att             Attributes
		token           string
		tokenExpiredIat string
		ctx             context.Context
		claims          jwtClaims
	}
)

func TestOAuthSuite(t *testing.T) {
	suite.Run(t, new(oauthSuite))
}

func (s *oauthSuite) SetupTest() {
	s.logger = &log.MockLogger{}
	s.cfg = config.OAuthAuthorizer{
		Enable: true,
		JwtCredentials: config.JwtCredentials{
			Algorithm:  jwt.RS256.String(),
			PublicKey:  "../../config/credentials/keytest.pub",
			PrivateKey: "../../config/credentials/keytest",
		},
		MaxJwtTTL: 300000001,
	}
	// https://jwt.io/#debugger-io?token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwicGVybWlzc2lvbiI6InJlYWQiLCJkb21haW4iOiJ0ZXN0LWRvbWFpbiIsImlhdCI6MTYyNjMzNjQ2MywiVFRMIjozMDAwMDAwMDB9.r1e83j6J392u4oAM7S7RYEDpeEilGThev2rK6RxqRXJIYiQlqKo1siDQjgHmj5PNUyEAQJF54CcXiaWJpTPWiPOxuRGtfJbUjSTnU2TiLvUiYU9bYt5U1w_UdlGzOD0ULhXPv2bzujAgtuQiRutwpljuQZwqqSDzILAMZlD5NMhEajYbE1P_0kv7esHO4oofTh__G3VZ_2fEi52GA8lwqoqBH3tQ1RK5QblnK5zMG5zBy8yK6JUmdoAGnKugjkJdDu8ERI4lNeIaWhD6kV8lksmPY0CxLfbmqLP3BIhvRF7zOeI1ocwa_4lpk4U6QRZ2w4hyGSEtD3sMmz1wl_uQCw&publicKey=-----BEGIN%20PUBLIC%20KEY-----%0AMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAscukltHilaq%2Bo5gIVE4P%0AGwWl%2BesvJ2EaEpWw6ogr98Un11YJ4oKkwIkLw4iIo0tveCINA3cZmxaW1RejRWKE%0AqYFtQ1rYd6BsnFAHXWh2R3A1FtpG6ANUEGkE7OAJe2%2FL42E%2FImJ%2BGQxRvartInDM%0AyfiRfB7%2BL2n3wG%2BNi%2BhBNMtAaX4Wwbj2hup21Jjuo96TuhcGImBFBATGWaYR2wqe%0A%2F6by9wJexPHlY%2F1uDp3SnzF1dCLjp76SGCfyYqOGC%2FPxhQi7mDxeH9%2FtIC%2Blt%2FSz%0Awc1n8gZLtlRlZHinvYa8lhWXqVYw6WD8h4LTgALq9iY%2BbeD1PFQSY1GkQtt0RhRw%0AeQIDAQAB%0A-----END%20PUBLIC%20KEY-----
	s.token = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ik
		pvaG4gRG9lIiwicGVybWlzc2lvbiI6InJlYWQiLCJkb21haW4iOiJ0ZXN0LWRvbWFpbiIsImlhdCI6MTYyNjMzNjQ
		2MywiVFRMIjozMDAwMDAwMDB9.r1e83j6J392u4oAM7S7RYEDpeEilGThev2rK6RxqRXJIYiQlqKo1siDQjgHmj5P
		NUyEAQJF54CcXiaWJpTPWiPOxuRGtfJbUjSTnU2TiLvUiYU9bYt5U1w_UdlGzOD0ULhXPv2bzujAgtuQiRutwplju
		QZwqqSDzILAMZlD5NMhEajYbE1P_0kv7esHO4oofTh__G3VZ_2fEi52GA8lwqoqBH3tQ1RK5QblnK5zMG5zBy8yK6
		JUmdoAGnKugjkJdDu8ERI4lNeIaWhD6kV8lksmPY0CxLfbmqLP3BIhvRF7zOeI1ocwa_4lpk4U6QRZ2w4hyGSEtD3
		sMmz1wl_uQCw`
	// https://jwt.io/#debugger-io?token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwicGVybWlzc2lvbiI6InJlYWQiLCJkb21haW4iOiJ0ZXN0LWRvbWFpbiIsImlhdCI6MTYyNjMzNjQ2MywiVFRMIjoxfQ.P_T3O54F_aiHcaMwyeh2GXtzgWhyKSLkuu8rtGAylK0HOsHYRIkbjdx251kaDEf2B-QP6KKCiXhDgZ_Q42Tb477zjl9IYGRqEj9JZ7PwGuRWCEZWUaFHgB4XmkviHDMamBB5jqg2I2XYklyNO3r2m45_AcQ3dAU4uLiwBwSVKy_YsMldEvGKMC86JvGcYPhu-LLvrJSViQVyuBGjUor6YREuadAZHyKuoMunLq5b_BW2hTf_67kGiyRL5_DxBBGbiNeHDPNoBUNUAx4Nbe1rAckREL8VULVFC_HZ0bDiM7KMJJ0t6zLcgP8Z3Q3341nfhv9r3qG_6U343ZgTPZfQNQ&publicKey=-----BEGIN%20PUBLIC%20KEY-----%0AMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAscukltHilaq%2Bo5gIVE4P%0AGwWl%2BesvJ2EaEpWw6ogr98Un11YJ4oKkwIkLw4iIo0tveCINA3cZmxaW1RejRWKE%0AqYFtQ1rYd6BsnFAHXWh2R3A1FtpG6ANUEGkE7OAJe2%2FL42E%2FImJ%2BGQxRvartInDM%0AyfiRfB7%2BL2n3wG%2BNi%2BhBNMtAaX4Wwbj2hup21Jjuo96TuhcGImBFBATGWaYR2wqe%0A%2F6by9wJexPHlY%2F1uDp3SnzF1dCLjp76SGCfyYqOGC%2FPxhQi7mDxeH9%2FtIC%2Blt%2FSz%0Awc1n8gZLtlRlZHinvYa8lhWXqVYw6WD8h4LTgALq9iY%2BbeD1PFQSY1GkQtt0RhRw%0AeQIDAQAB%0A-----END%20PUBLIC%20KEY-----
	s.tokenExpiredIat = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwib
		mFtZSI6IkpvaG4gRG9lIiwicGVybWlzc2lvbiI6InJlYWQiLCJkb21haW4iOiJ0ZXN0LWRvbWFpbiIsImlhdCI6MTY
		yNjMzNjQ2MywiVFRMIjoxfQ.P_T3O54F_aiHcaMwyeh2GXtzgWhyKSLkuu8rtGAylK0HOsHYRIkbjdx251kaDEf2B-
		QP6KKCiXhDgZ_Q42Tb477zjl9IYGRqEj9JZ7PwGuRWCEZWUaFHgB4XmkviHDMamBB5jqg2I2XYklyNO3r2m45_AcQ3
		dAU4uLiwBwSVKy_YsMldEvGKMC86JvGcYPhu-LLvrJSViQVyuBGjUor6YREuadAZHyKuoMunLq5b_BW2hTf_67kGiy
		RL5_DxBBGbiNeHDPNoBUNUAx4Nbe1rAckREL8VULVFC_HZ0bDiM7KMJJ0t6zLcgP8Z3Q3341nfhv9r3qG_6U343ZgT
		PZfQNQ`

	re := regexp.MustCompile(`\r?\n?\t`)
	s.token = re.ReplaceAllString(s.token, "")
	s.tokenExpiredIat = re.ReplaceAllString(s.tokenExpiredIat, "")

	ctx := context.Background()
	ctx, call := encoding.NewInboundCall(ctx)
	err := call.ReadFromRequest(&transport.Request{
		Headers: transport.NewHeaders().With(common.AuthorizationTokenHeaderName, s.token),
	})
	s.NoError(err)
	s.att = Attributes{
		Actor:      "John Doe",
		APIName:    "",
		DomainName: "test-domain",
		TaskList:   nil,
		Permission: PermissionRead,
	}
	s.ctx = ctx
	s.claims = jwtClaims{
		Sub:        "test",
		Name:       "Test",
		Permission: "write",
		Domain:     "test-domain",
		Iat:        time.Now().Unix(),
		TTL:        300000,
	}
}

func (s *oauthSuite) TearDownTest() {
	s.logger.AssertExpectations(s.T())
}

func (s *oauthSuite) TestCorrectPayloadAuthorize() {
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	result, err := authorizer.Authorize(s.ctx, &s.att)
	s.NoError(err)
	s.Equal(result.Decision, DecisionAllow)
}

func (s *oauthSuite) TestIncorrectPublicKeyAuthorize() {
	s.cfg.JwtCredentials.PublicKey = "incorrectPublicKey"
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	result, err := authorizer.Authorize(s.ctx, &s.att)
	s.EqualError(err, "invalid public key path incorrectPublicKey")
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestIncorrectAlgorithmAuthorize() {
	s.cfg.JwtCredentials.Algorithm = "SHA256"
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	result, err := authorizer.Authorize(s.ctx, &s.att)
	s.EqualError(err, "jwt: algorithm is not supported")
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestMaxTTLLargerInTokenAuthorize() {
	s.cfg.MaxJwtTTL = 1
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	s.logger.On("Debug", "request is not authorized", mock.MatchedBy(func(t []tag.Tag) bool {
		return fmt.Sprintf("%v", t[0].Field().Interface) == "TTL in token is larger than MaxTTL allowed"
	}))
	result, _ := authorizer.Authorize(s.ctx, &s.att)
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestIncorrectTokenAuthorize() {
	ctx := context.Background()
	ctx, call := encoding.NewInboundCall(ctx)
	err := call.ReadFromRequest(&transport.Request{
		Headers: transport.NewHeaders().With(common.AuthorizationTokenHeaderName, "test"),
	})
	s.NoError(err)
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	s.logger.On("Debug", "request is not authorized", mock.MatchedBy(func(t []tag.Tag) bool {
		return fmt.Sprintf("%v", t[0].Field().Interface) == "jwt: token format is not valid"
	}))
	result, _ := authorizer.Authorize(ctx, &s.att)
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestIatExpiredTokenAuthorize() {
	ctx := context.Background()
	ctx, call := encoding.NewInboundCall(ctx)
	err := call.ReadFromRequest(&transport.Request{
		Headers: transport.NewHeaders().With(common.AuthorizationTokenHeaderName, s.tokenExpiredIat),
	})
	s.NoError(err)
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	s.logger.On("Debug", "request is not authorized", mock.MatchedBy(func(t []tag.Tag) bool {
		return fmt.Sprintf("%v", t[0].Field().Interface) == "JWT has expired"
	}))
	result, _ := authorizer.Authorize(ctx, &s.att)
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestIncorrectPermissionInAttributesAuthorize() {
	s.att.Permission = PermissionWrite
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	s.logger.On("Debug", "request is not authorized", mock.MatchedBy(func(t []tag.Tag) bool {
		return fmt.Sprintf("%v", t[0].Field().Interface) == "token doesn't have the right permission"
	}))
	result, _ := authorizer.Authorize(s.ctx, &s.att)
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestIncorrectDomainInAttributesAuthorize() {
	s.att.DomainName = "myotherdomain"
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	s.logger.On("Debug", "request is not authorized", mock.MatchedBy(func(t []tag.Tag) bool {
		return fmt.Sprintf("%v", t[0].Field().Interface) == "domain in token doesn't match with current domain"
	}))
	result, _ := authorizer.Authorize(s.ctx, &s.att)
	s.Equal(result.Decision, DecisionDeny)
}

func (s *oauthSuite) TestCorrectTokenCreationAuthenticate() {
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	result, err := authorizer.Authenticate(s.ctx, s.claims)
	s.NoError(err)
	ctx := context.Background()
	ctx, call := encoding.NewInboundCall(ctx)
	token := result.(*string)
	err = call.ReadFromRequest(&transport.Request{
		Headers: transport.NewHeaders().With(common.AuthorizationTokenHeaderName, *token),
	})
	s.NoError(err)
	decision, err := authorizer.Authorize(ctx, &s.att)
	s.NoError(err)
	s.Equal(decision.Decision, DecisionAllow)
}

func (s *oauthSuite) TestIncorrectPrivateKeyForTokenCreationAuthenticate() {
	s.cfg.JwtCredentials.PrivateKey = "NotAPrivateKey"
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	token, err := authorizer.Authenticate(s.ctx, s.claims)
	s.EqualError(err, "invalid private key path NotAPrivateKey")
	s.Nil(token)
}

func (s *oauthSuite) TestIncorrectAlgorithmAuthenticate() {
	s.cfg.JwtCredentials.Algorithm = "SHA256"
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	token, err := authorizer.Authenticate(s.ctx, s.claims)
	s.EqualError(err, "jwt: algorithm is not supported")
	s.Nil(token)
}

func (s *oauthSuite) TestIncorrectClaimsType() {
	authorizer := NewOAuthAuthorizer(s.cfg, s.logger)
	token, err := authorizer.Authenticate(s.ctx, "claims")
	s.EqualError(err, "jwtClaims type expected, string provided")
	s.Nil(token)
}
