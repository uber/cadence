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
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"
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
	publicKey        *rsa.PublicKey
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
		return nil, err
	}
	return &oauthAuthority{
		authorizationCfg: authorizationCfg,
		domainCache:      domainCache,
		log:              log,
		publicKey:        publicKey,
	}, nil
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
	if token == "" {
		a.log.Debug("request is not authorized", tag.Error(fmt.Errorf("token is not set in header")))
		return Result{Decision: DecisionDeny}, nil
	}
	claims, err := a.parseToken(token, verifier)
	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}
	err = a.validateTTL(claims)
	if err != nil {
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

	err = a.validatePermission(claims, attributes, domain.GetInfo().Data)
	if err != nil {
		a.log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}
	return Result{Decision: DecisionAllow}, nil
}

func (a *oauthAuthority) getVerifier() (jwt.Verifier, error) {

	algorithm := jwt.Algorithm(a.authorizationCfg.JwtCredentials.Algorithm)
	verifier, err := jwt.NewVerifierRS(algorithm, a.publicKey)
	if err != nil {
		return nil, err
	}
	return verifier, nil
}

func (a *oauthAuthority) parseToken(tokenStr string, verifier jwt.Verifier) (*JWTClaims, error) {
	token, verifyErr := jwt.ParseAndVerifyString(tokenStr, verifier)
	if verifyErr != nil {
		return nil, verifyErr
	}
	var claims JWTClaims
	_ = json.Unmarshal(token.RawClaims(), &claims)
	return &claims, nil
}

func (a *oauthAuthority) validateTTL(claims *JWTClaims) error {
	if claims.TTL > a.authorizationCfg.MaxJwtTTL {
		return fmt.Errorf("TTL in token is larger than MaxTTL allowed")
	}
	if claims.Iat+claims.TTL < time.Now().Unix() {
		return fmt.Errorf("JWT has expired")
	}
	return nil
}

func (a *oauthAuthority) validatePermission(claims *JWTClaims, attributes *Attributes, data map[string]string) error {
	groups := ""
	switch attributes.Permission {
	case PermissionRead:
		groups = data[common.DomainDataKeyForReadGroups] + groupSeparator + data[common.DomainDataKeyForWriteGroups]
	case PermissionWrite:
		groups = data[common.DomainDataKeyForWriteGroups]
	default:
		return fmt.Errorf("token doesn't have permission for %v API", attributes.Permission)
	}
	// groups are separated by space
	allowedGroups := strings.Split(groups, groupSeparator)    // groups that allowed by domain configuration(in domainData)
	jwtGroups := strings.Split(claims.Groups, groupSeparator) // groups that the request has associated with

	for _, group1 := range allowedGroups {
		for _, group2 := range jwtGroups {
			if group1 == group2 {
				return nil
			}
		}
	}
	return fmt.Errorf("token doesn't have the right permission, jwt groups: %v, allowed groups: %v", jwtGroups, allowedGroups)
}
