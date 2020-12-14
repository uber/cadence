// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package gocql

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/gocql/gocql"
)

var _ Client = client{}

type (
	client struct{}
)

var (
	defaultClient = client{}
)

// NewClient creates a default gocql client based on the open source gocql library.
func NewClient() Client {
	return defaultClient
}

func (c client) CreateSession(
	config ClusterConfig,
) (Session, error) {
	cluster := newCassandraCluster(config)
	cluster.ProtoVersion = config.ProtoVersion
	cluster.Consistency = mustConvertConsistency(config.Consistency)
	cluster.SerialConsistency = mustConvertSerialConsistency(config.SerialConsistency)
	cluster.Timeout = config.Timeout
	gocqlSession, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &session{
		Session: gocqlSession,
	}, nil
}

func (c client) IsTimeoutError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	if err == gocql.ErrTimeoutNoResponse {
		return true
	}
	if err == gocql.ErrConnectionClosed {
		return true
	}
	_, ok := err.(*gocql.RequestErrWriteTimeout)
	return ok
}

func (c client) IsNotFoundError(err error) bool {
	return err == gocql.ErrNotFound
}

func (c client) IsThrottlingError(err error) bool {
	if req, ok := err.(gocql.RequestError); ok {
		// gocql does not expose the constant errOverloaded = 0x1001
		return req.Code() == 0x1001
	}
	return false
}

func newCassandraCluster(cfg ClusterConfig) *gocql.ClusterConfig {
	hosts := parseHosts(cfg.Hosts)
	cluster := gocql.NewCluster(hosts...)
	cluster.ProtoVersion = 4
	if cfg.Port > 0 {
		cluster.Port = cfg.Port
	}
	if cfg.User != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.User,
			Password: cfg.Password,
		}
	}
	if cfg.Keyspace != "" {
		cluster.Keyspace = cfg.Keyspace
	}
	if cfg.Datacenter != "" {
		cluster.HostFilter = gocql.DataCentreHostFilter(cfg.Datacenter)
	}
	if cfg.Region != "" {
		cluster.HostFilter = regionHostFilter(cfg.Region)
	}

	if cfg.TLS != nil && cfg.TLS.Enabled {
		cluster.SslOpts = &gocql.SslOptions{
			CertPath:               cfg.TLS.CertFile,
			KeyPath:                cfg.TLS.KeyFile,
			CaPath:                 cfg.TLS.CaFile,
			EnableHostVerification: cfg.TLS.EnableHostVerification,

			Config: &tls.Config{
				ServerName: cfg.TLS.ServerName,
			},
		}
	}
	if cfg.MaxConns > 0 {
		cluster.NumConns = cfg.MaxConns
	}

	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	return cluster
}

// regionHostFilter returns a gocql host filter for the given region name
func regionHostFilter(region string) gocql.HostFilter {
	return gocql.HostFilterFunc(func(host *gocql.HostInfo) bool {
		applicationRegion := region
		if len(host.DataCenter()) < 3 {
			return false
		}
		return host.DataCenter()[:3] == applicationRegion
	})
}

// parseHosts returns parses a list of hosts separated by comma
func parseHosts(input string) []string {
	var hosts = make([]string, 0)
	for _, h := range strings.Split(input, ",") {
		if host := strings.TrimSpace(h); len(host) > 0 {
			hosts = append(hosts, host)
		}
	}
	return hosts
}
