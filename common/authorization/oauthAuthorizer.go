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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type oauthAuthority struct {
	authorizationCfg config.OAuthAuthorizer
	log              log.Logger
}

type jwtClaims struct {
	Sub        string
	Name       string
	Permission string
	Domain     string
	Iat        int64
	TTL        int64
}

// NewOAuthAuthorizer creates a oauth authority
func NewOAuthAuthorizer(
	authorizationCfg config.OAuthAuthorizer,
	log log.Logger,
) Authorizer {
	return &oauthAuthority{
		authorizationCfg: authorizationCfg,
		log:              log,
	}
}

// Authorize defines the logic to verify get claims from token
func (a *oauthAuthority) Authorize(
	ctx context.Context,
	attributes *Attributes,
) (Result, error) {
	call := yarpc.CallFromContext(ctx)
	verifier, err := a.getVerifier()
	if err != nil {
		return Result{Decision: DecisionDeny}, err
	}
	token := call.Header(common.AuthorizationTokenHeaderName)
	claims, err := a.parseToken(token, verifier)
	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}
	err = a.validateClaims(claims, attributes)
	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}
	return Result{Decision: DecisionAllow}, nil
}

// Authenticate defines the logic to authenticate a user
func (a *oauthAuthority) Authenticate(
	_ context.Context,
	claims interface{},
) (interface{}, error) {
	claimsJwt, ok := claims.(jwtClaims)
	if !ok {
		return nil, fmt.Errorf("jwtClaims type expected, %T provided", claims)
	}
	privateKey, err := common.LoadRSAPrivateKey(a.authorizationCfg.JwtCredentials.PrivateKey)
	if err != nil {
		return nil, err
	}
	algorithm := jwt.Algorithm(a.authorizationCfg.JwtCredentials.Algorithm)
	signer, err := jwt.NewSignerRS(algorithm, privateKey)
	if err != nil {
		return nil, err
	}
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claimsJwt)
	if token == nil {
		return nil, err
	}
	tokenString := token.String()
	return &tokenString, nil
}

func (a *oauthAuthority) getVerifier() (jwt.Verifier, error) {
	publicKey, err := common.LoadRSAPublicKey(a.authorizationCfg.JwtCredentials.PublicKey)
	if err != nil {
		return nil, err
	}
	algorithm := jwt.Algorithm(a.authorizationCfg.JwtCredentials.Algorithm)
	verifier, err := jwt.NewVerifierRS(algorithm, publicKey)
	if err != nil {
		return nil, err
	}
	return verifier, nil
}

func (a *oauthAuthority) parseToken(tokenStr string, verifier jwt.Verifier) (*jwtClaims, error) {
	token, verifyErr := jwt.ParseAndVerifyString(tokenStr, verifier)
	if verifyErr != nil {
		return nil, verifyErr
	}
	var claims jwtClaims
	_ = json.Unmarshal(token.RawClaims(), &claims)
	return &claims, nil
}

func (a *oauthAuthority) validateClaims(claims *jwtClaims, attributes *Attributes) error {
	if claims.TTL > a.authorizationCfg.MaxJwtTTL {
		return fmt.Errorf("TTL in token is larger than MaxTTL allowed")
	}
	if claims.Iat+claims.TTL < time.Now().Unix() {
		return fmt.Errorf("JWT has expired")
	}
	if claims.Domain != attributes.DomainName {
		return fmt.Errorf("domain in token doesn't match with current domain")
	}
	if NewPermission(claims.Permission) < attributes.Permission {
		return fmt.Errorf("token doesn't have the right permission")
	}

	return nil
}
