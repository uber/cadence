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
	"errors"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type oauthAuthority struct {
	authorizationCfg config.OAuthAuthorizer
	domainCache      cache.DomainCache
	log              log.Logger
	verifier         jwt.Verifier
}

type JWTClaims struct {
	Sub    string
	Name   string
	Groups string // separated by space
	Admin  bool
	Iat    int64
	TTL    int64
}

const groupSeparator = " "

// NewOAuthAuthorizer creates a oauth authority
func NewOAuthAuthorizer(
	authorizationCfg config.OAuthAuthorizer,
	log log.Logger,
	domainCache cache.DomainCache,
) (Authorizer, error) {
	publicKey, err := common.LoadRSAPublicKey(authorizationCfg.JwtCredentials.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("loading RSA public key: %w", err)
	}

	verifier, err := jwt.NewVerifierRS(
		jwt.Algorithm(authorizationCfg.JwtCredentials.Algorithm),
		publicKey,
	)

	if err != nil {
		return nil, fmt.Errorf("creating JWT verifier: %w", err)
	}

	return &oauthAuthority{
		authorizationCfg: authorizationCfg,
		domainCache:      domainCache,
		log:              log,
		verifier:         verifier,
	}, nil
}

// Authorize defines the logic to verify get claims from token
func (a *oauthAuthority) Authorize(
	ctx context.Context,
	attributes *Attributes,
) (Result, error) {
	call := yarpc.CallFromContext(ctx)

	token := call.Header(common.AuthorizationTokenHeaderName)
	if token == "" {
		a.log.Debug("request is not authorized", tag.Error(errors.New("token is not set in header")))
		return Result{Decision: DecisionDeny}, nil
	}

	claims, err := a.parseToken(token, a.verifier)
	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}

	if err := a.validateTTL(claims); err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}

	if claims.Admin {
		return Result{Decision: DecisionAllow}, nil
	}
	domain, err := a.domainCache.GetDomain(attributes.DomainName)
	if err != nil {
		return Result{Decision: DecisionDeny}, err
	}

	if err := validatePermission(claims, attributes, domain.GetInfo().Data); err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}

	return Result{Decision: DecisionAllow}, nil
}

func (a *oauthAuthority) parseToken(tokenStr string, verifier jwt.Verifier) (*JWTClaims, error) {
	token, err := jwt.ParseAndVerifyString(tokenStr, verifier)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	var claims JWTClaims
	_ = json.Unmarshal(token.RawClaims(), &claims)
	return &claims, nil
}

func (a *oauthAuthority) validateTTL(claims *JWTClaims) error {
	if claims.TTL > a.authorizationCfg.MaxJwtTTL {
		return fmt.Errorf("token TTL: %d is larger than MaxTTL allowed: %d", claims.TTL, a.authorizationCfg.MaxJwtTTL)
	}
	if claims.Iat+claims.TTL < time.Now().Unix() {
		return errors.New("JWT has expired")
	}
	return nil
}
