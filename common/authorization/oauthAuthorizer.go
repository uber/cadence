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
	log "github.com/sirupsen/logrus"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/tag"
)

type oauthAuthority struct {
	authorizationCfg config.OAuthAuthorizer
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
) Authorizer {
	return &oauthAuthority{
		authorizationCfg: authorizationCfg,
	}
}

// Authorize defines the logic to verify get claims from token
func (a *oauthAuthority) Authorize(
	ctx context.Context,
	attributes *Attributes,
) (Result, error) {
	call := yarpc.CallFromContext(ctx)
	token := call.Header(common.AuthorizationTokenHeaderName)
	// parseTokenAndVerify could either return the claims or a bool. I'm returning the claims currently
	// because we might use it to populate values in the context
	claims, err := a.parseTokenAndVerify(token, attributes)
	if err != nil {
		return Result{Decision: DecisionDeny}, err
	}
	validationErr := a.validateClaims(claims, attributes)
	if validationErr != nil {
		log.Debug("request is not authorized", tag.Error(err))
		return Result{Decision: DecisionDeny}, nil
	}
	return Result{Decision: DecisionAllow}, nil
}

func (a *oauthAuthority) parseTokenAndVerify(tokenStr string, attributes *Attributes) (*jwtClaims, error) {
	publicKey, err := common.StringToRSAPublicKey(a.authorizationCfg.JwtCredentials.PublicKey)
	if err != nil {
		return nil, err
	}
	algorithm := jwt.Algorithm(a.authorizationCfg.JwtCredentials.Algorithm)
	verifier, err := jwt.NewVerifierRS(algorithm, publicKey)
	if err != nil {
		return nil, err
	}
	token, verifyErr := jwt.ParseAndVerifyString(tokenStr, verifier)
	if verifyErr != nil {
		return nil, verifyErr
	}
	var claims jwtClaims
	_ = json.Unmarshal(token.RawClaims(), &claims)
	return &claims, nil
}

func (a *oauthAuthority) validateClaims(claims *jwtClaims, attributes *Attributes) error {
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
