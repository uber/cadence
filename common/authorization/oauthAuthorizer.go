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
	"context"
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/uber/cadence/common/config"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

type oauthAuthority struct{
	cfg *config.Config
}

type jwtClaims struct {
	name string
	permission string
	domain string
	iat int64
}

// NewOAuthAuthorizer creates a no-op authority
func NewOAuthAuthorizer(
	cfg *config.Config,
	) Authorizer {
	return &oauthAuthority{
		cfg: cfg,
	}
}

func (a *oauthAuthority) Authorize(
	ctx context.Context,
	attributes *Attributes,
) (Result, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return Result{Decision: DecisionDeny}, fmt.Errorf("ctx metadata was not found. Unable to authorize")
	}
	// potentially claims will populate the context, otherwise we can
	// return bool here
	_, err := a.validateAndGetClaims(md["authorization"])
	if err != nil {
		return Result{Decision: DecisionDeny}, fmt.Errorf("authentication required")
	}
	return Result{Decision: DecisionAllow}, nil
}

func (a *oauthAuthority) validateAndGetClaims(authorization []string) (*jwtClaims, error) {
	key := []byte(a.cfg.Authorization.OAuthAuthorizer.JwtCredentials.PrivateKey)
	algorithm := jwt.Algorithm(a.cfg.Authorization.OAuthAuthorizer.JwtCredentials.Algorithm)
	verifier, err := jwt.NewVerifierHS(algorithm, key)
	if err != nil {
		return nil, err
	}
	tokenStr := strings.Join(authorization, ".")
	token, verifyErr := jwt.ParseAndVerifyString(tokenStr, verifier)
	if verifyErr != nil {
		return nil, verifyErr
	}
	var claims jwtClaims
	claimsErr := json.Unmarshal(token.RawClaims(), &claims)
	if claimsErr != nil {
		return nil, err
	}
	if claims.iat < time.Now().Unix() {
		return nil, fmt.Errorf("JWT has expired")
	}

	return &claims, nil
}

