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
    "github.com/uber/cadence/common"
    "go.uber.org/yarpc"
    "time"

    "github.com/cristalhq/jwt/v3"
    "github.com/uber/cadence/common/config"
)

type oauthAuthority struct {
    authorizationCfg config.OAuthAuthorizer
}

type jwtClaims struct {
    Name       string
    Permission string
    Domain     string
    Iat        int64
}

// NewOAuthAuthorizer creates a no-op authority
func NewOAuthAuthorizer(
    authorizationCfg config.OAuthAuthorizer,
) Authorizer {
    return &oauthAuthority{
        authorizationCfg: authorizationCfg,
    }
}

func (a *oauthAuthority) Authorize(
    ctx context.Context,
    attributes *Attributes,
) (Result, error) {
    call := yarpc.CallFromContext(ctx)
    token :=  call.Header(common.AuthorizationTokenHeaderName)
    // parseTokenAndVerify could either return the claims or a bool. I'm returning the claims currently
    // because we might use it to populate values in the context
    _, err := a.parseTokenAndVerify(token)
    if err != nil {
        return Result{Decision: DecisionDeny}, err
    }
    return Result{Decision: DecisionAllow}, nil
}

func (a *oauthAuthority) parseTokenAndVerify(tokenStr string) (*jwtClaims, error) {
    key := []byte(a.authorizationCfg.JwtCredentials.PrivateKey)
    algorithm := jwt.Algorithm(a.authorizationCfg.JwtCredentials.Algorithm)
    verifier, err := jwt.NewVerifierHS(algorithm, key)
    if err != nil {
        return nil, err
    }
    token, verifyErr := jwt.ParseAndVerifyString(tokenStr, verifier)
    if verifyErr != nil {
        return nil, verifyErr
    }
    var claims jwtClaims
    claimsErr := json.Unmarshal(token.RawClaims(), &claims)
    if claimsErr != nil {
        return nil, err
    }
    validationErr := a.validateClaims(claims)
    if validationErr != nil {
        return nil, validationErr
    }

    return &claims, nil
}

func (a *oauthAuthority) validateClaims(claims jwtClaims) error {
    if claims.Iat < time.Now().Unix() {
        return fmt.Errorf("JWT has expired")
    }

    return nil
}
