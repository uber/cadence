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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

var _ jwt.Claims = (*JWTClaims)(nil)

type oauthAuthority struct {
	authorizationCfg config.OAuthAuthorizer
	domainCache      cache.DomainCache
	log              log.Logger
	parser           *jwt.Parser
	publicKey        interface{}
}

// JWTClaims is a Cadence specific claim with embeded Claims defined https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
type JWTClaims struct {
	jwt.RegisteredClaims

	Name   string
	Groups string // separated by space
	Admin  bool
	TTL    int64 // TODO should be removed. ExpiresAt should be used
}

func (j JWTClaims) GetGroups() []string {
	return strings.Split(j.Groups, groupSeparator)
}

const groupSeparator = " "

// NewOAuthAuthorizer creates a oauth authority
func NewOAuthAuthorizer(
	authorizationCfg config.OAuthAuthorizer,
	log log.Logger,
	domainCache cache.DomainCache,
) (Authorizer, error) {

	key, err := common.LoadRSAPublicKey(authorizationCfg.JwtCredentials.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("loading RSA public key: %w", err)
	}

	if authorizationCfg.JwtCredentials.Algorithm != jwt.SigningMethodRS256.Name {
		return nil, fmt.Errorf("algorithm %q is not supported", authorizationCfg.JwtCredentials.Algorithm)
	}

	return &oauthAuthority{
		authorizationCfg: authorizationCfg,
		domainCache:      domainCache,
		log:              log,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{authorizationCfg.JwtCredentials.Algorithm}),
			jwt.WithIssuedAt(),
		),
		publicKey: key,
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

	var claims JWTClaims

	_, err := a.parser.ParseWithClaims(token, &claims, a.keyFunc)

	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}

	if err := a.validateTTL(&claims); err != nil {
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

	if err := validatePermission(&claims, attributes, domain.GetInfo().Data); err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}

	return Result{Decision: DecisionAllow}, nil
}

// keyFunc returns correct key to check signature
func (a *oauthAuthority) keyFunc(token *jwt.Token) (interface{}, error) {
	// only local public key is supported currently
	return a.publicKey, nil
}

func (a *oauthAuthority) validateTTL(claims *JWTClaims) error {
	// Fill ExpiresAt when TTL is passed
	if claims.TTL > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(claims.IssuedAt.Time.Add(time.Second * time.Duration(claims.TTL)))
	}

	exp, err := claims.GetExpirationTime()

	if err != nil || exp == nil {
		return errors.New("ExpiresAt is not set")
	}

	timeLeft := exp.Unix() - time.Now().Unix()
	if timeLeft < 0 {
		return errors.New("token is expired")
	}

	if timeLeft > a.authorizationCfg.MaxJwtTTL {
		return fmt.Errorf("token TTL: %d is larger than MaxTTL allowed: %d", timeLeft, a.authorizationCfg.MaxJwtTTL)
	}

	return nil
}
