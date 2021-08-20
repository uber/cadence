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

package config

import (
	"fmt"

	"github.com/cristalhq/jwt/v3"
)

// Validate validates the persistence config
func (a *Authorization) Validate() error {
	if a.OAuthAuthorizer.Enable && a.NoopAuthorizer.Enable {
		return fmt.Errorf("[AuthorizationConfig] More than one authorizer is enabled")
	}

	if a.OAuthAuthorizer.Enable {
		if oauthError := a.validateOAuth(); oauthError != nil {
			return oauthError
		}
	}

	return nil
}

func (a *Authorization) validateOAuth() error {
	oauthConfig := a.OAuthAuthorizer

	if oauthConfig.MaxJwtTTL <= 0 {
		return fmt.Errorf("[OAuthConfig] MaxTTL must be greater than 0")
	}
	if oauthConfig.JwtCredentials.PublicKey == "" {
		return fmt.Errorf("[OAuthConfig] PublicKey can't be empty")
	}
	if oauthConfig.JwtCredentials.Algorithm != jwt.RS256.String() {
		return fmt.Errorf("[OAuthConfig] The only supported Algorithm is RS256")
	}
	return nil
}
