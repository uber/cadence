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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

type (
	// TLS describe TLS configuration
	TLS struct {
		Enabled bool `yaml:"enabled"`

		// For Postgres(https://www.postgresql.org/docs/9.1/libpq-ssl.html) and MySQL
		// default to require if Enable is true.
		// For MySQL: https://github.com/go-sql-driver/mysql , it also can be set in ConnectAttributes, default is tls-custom
		SSLMode string `yaml:"sslmode" `

		// CertPath and KeyPath are optional depending on server
		// config, but both fields must be omitted to avoid using a
		// client certificate
		CertFile string `yaml:"certFile"`
		KeyFile  string `yaml:"keyFile"`

		CaFile string `yaml:"caFile"` //optional depending on server config
		// If you want to verify the hostname and server cert (like a wildcard for cass cluster) then you should turn this on
		// This option is basically the inverse of InSecureSkipVerify
		// See InSecureSkipVerify in http://golang.org/pkg/crypto/tls/ for more info
		EnableHostVerification bool `yaml:"enableHostVerification"`

		ServerName string `yaml:"serverName"`
	}
)

// ToTLSConfig converts Cadence TLS config to crypto/tls.Config
func (config TLS) ToTLSConfig() (*tls.Config, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Setup base TLS config
	// EnableHostVerification is a secure flag vs insecureSkipVerify is insecure so inverse the value
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !config.EnableHostVerification,
	}

	// Setup server name
	if config.ServerName != "" {
		tlsConfig.ServerName = config.ServerName
	}

	// Load CA cert
	if config.CaFile != "" {
		caCert, err := ioutil.ReadFile(config.CaFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// Load client cert
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
