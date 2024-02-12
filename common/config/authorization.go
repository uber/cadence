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
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type (
	Authorization struct {
		OAuthAuthorizer OAuthAuthorizer `yaml:"oauthAuthorizer"`
		NoopAuthorizer  NoopAuthorizer  `yaml:"noopAuthorizer"`
	}

	NoopAuthorizer struct {
		Enable bool `yaml:"enable"`
	}

	OAuthAuthorizer struct {
		Enable bool `yaml:"enable"`
		// Max of TTL in the claim
		MaxJwtTTL int64 `yaml:"maxJwtTTL"`
		// Credentials to verify/create the JWT using public/private keys
		JwtCredentials *JwtCredentials `yaml:"jwtCredentials"`
		// Provider
		Provider *OAuthProvider `yaml:"provider"`
	}

	JwtCredentials struct {
		// support: RS256 (RSA using SHA256)
		Algorithm string `yaml:"algorithm"`
		// Public Key Path for verifying JWT token passed in from external clients
		PublicKey string `yaml:"publicKey"`
	}

	// OAuthProvider is used to validate tokens provided by 3rd party Identity Provider service
	OAuthProvider struct {
		JWKSURL             string `yaml:"jwksURL"`
		GroupsAttributePath string `yaml:"groupsAttributePath"`
		AdminAttributePath  string `yaml:"adminAttributePath"`
	}
)

// Validate validates the persistence config
func (a *Authorization) Validate() error {
	if a.OAuthAuthorizer.Enable && a.NoopAuthorizer.Enable {
		return fmt.Errorf("[AuthorizationConfig] More than one authorizer is enabled")
	}

	if a.OAuthAuthorizer.Enable {
		if err := a.validateOAuth(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Authorization) validateOAuth() error {
	oauthConfig := a.OAuthAuthorizer

	if oauthConfig.MaxJwtTTL <= 0 {
		return fmt.Errorf("[OAuthConfig] MaxTTL must be greater than 0")
	}

	if oauthConfig.JwtCredentials == nil && oauthConfig.Provider == nil {
		return errors.New("jwtCredentials or provider must be provided")
	}

	if oauthConfig.JwtCredentials != nil {
		if oauthConfig.JwtCredentials.PublicKey == "" {
			return fmt.Errorf("[OAuthConfig] PublicKey can't be empty")
		}

		if oauthConfig.JwtCredentials.Algorithm != jwt.SigningMethodRS256.Name {
			return fmt.Errorf("[OAuthConfig] The only supported Algorithm is RS256")
		}
	}

	if oauthConfig.Provider != nil {
		if oauthConfig.Provider.JWKSURL == "" {
			return fmt.Errorf("[OAuthConfig] JWKSURL is not set")
		}
	}

	return nil
}
