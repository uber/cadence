// Copyright (c) 2017 Uber Technologies, Inc.
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
	"net/url"

	"github.com/uber/cadence/common"
)

var errAWSSigningCredential = fmt.Errorf("must provide exactly one type of credential, EnvironmentCredential or StaticCredential")

// ElasticSearchConfig for connecting to ElasticSearch
type (
	ElasticSearchConfig struct {
		URL     url.URL           `yaml:"url"`     //nolint:govet
		Indices map[string]string `yaml:"indices"` //nolint:govet
		// supporting v6 and v7. Default to v6 if empty.
		Version string `yaml:"version"` //nolint:govet
		// optional username to communicate with ElasticSearch
		Username string `yaml:"username"` //nolint:govet
		// optional password to communicate with ElasticSearch
		Password string `yaml:"password"` //nolint:govet
		// optional to disable sniff, according to issues on Github,
		// Sniff could cause issue like "no Elasticsearch node available"
		DisableSniff bool `yaml:"disableSniff"`
		// optional to disable health check
		DisableHealthCheck bool `yaml:"disableHealthCheck"`
		// optional to use AWS signing client
		// See more info https://github.com/olivere/elastic/wiki/Using-with-AWS-Elasticsearch-Service
		AWSSigning AWSSigning `yaml:"awsSigning"`
		// optional to use Signed Certificates over https
		TLS TLS `yaml:"tls"`
	}

	// AWSSigning contains config to enable signing,
	// Must provide either StaticCredential or EnvironmentCredential
	AWSSigning struct {
		Enable                bool                      `yaml:"enable"`
		StaticCredential      *AWSStaticCredential      `yaml:"staticCredential"`
		EnvironmentCredential *AWSEnvironmentCredential `yaml:"environmentCredential"`
	}

	// AWSStaticCredential to create a static credentials value provider.
	// SessionToken is only required for temporary security credentials retrieved via STS,
	// otherwise an empty string can be passed for this parameter.
	// See more in https://github.com/aws/aws-sdk-go/blob/master/aws/credentials/static_provider.go#L21
	AWSStaticCredential struct {
		AccessKey    string `yaml:"accessKey"`
		SecretKey    string `yaml:"secretKey"`
		Region       string `yaml:"region"`
		SessionToken string `yaml:"sessionToken"`
	}

	// AWSEnvironmentCredential will make a new Session created from SDK defaults, config files,
	// environment, and user provided config files.
	// See more in https://github.com/aws/aws-sdk-go/blob/3974dd034387fbc7cf09c8cd2400787ce07f3285/aws/session/session.go#L147
	AWSEnvironmentCredential struct {
		Region string `yaml:"region"`
	}
)

// GetVisibilityIndex return visibility index name
func (cfg *ElasticSearchConfig) GetVisibilityIndex() string {
	return cfg.Indices[common.VisibilityAppName]
}

// SetUsernamePassword set the username/password into URL
// It is a bit tricky here because url.URL doesn't expose the username/password in the struct
// because of the security concern.
func (cfg *ElasticSearchConfig) SetUsernamePassword() {
	if cfg.Username != "" {
		cfg.URL.User = url.UserPassword(cfg.Username, cfg.Password)
	}
}

// CheckAWSSigningConfig checks if the AWSSigning configuration is valid
func CheckAWSSigningConfig(config AWSSigning) error {
	if config.EnvironmentCredential == nil && config.StaticCredential == nil {
		return errAWSSigningCredential
	}
	if config.EnvironmentCredential != nil && config.StaticCredential != nil {
		return errAWSSigningCredential
	}
	return nil
}
