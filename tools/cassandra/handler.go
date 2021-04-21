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

package cassandra

import (
	"fmt"
	"log"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/schema/cassandra"
	"github.com/uber/cadence/tools/common/schema"
)

const defaultNumReplicas = 1

// SetupSchemaConfig contains the configuration params needed to setup schema tables
type SetupSchemaConfig struct {
	CQLClientConfig
	schema.SetupConfig
}

// VerifyCompatibleVersion ensures that the installed version of cadence and visibility keyspaces
// is greater than or equal to the expected version.
// In most cases, the versions should match. However if after a schema upgrade there is a code
// rollback, the code version (expected version) would fall lower than the actual version in
// cassandra.
func VerifyCompatibleVersion(
	cfg config.Persistence,
) error {

	ds, ok := cfg.DataStores[cfg.DefaultStore]
	if ok && ds.Cassandra != nil {
		err := CheckCompatibleVersion(*ds.Cassandra, cassandra.Version)
		if err != nil {
			return err
		}
	}
	ds, ok = cfg.DataStores[cfg.VisibilityStore]
	if ok && ds.Cassandra != nil {
		err := CheckCompatibleVersion(*ds.Cassandra, cassandra.VisibilityVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckCompatibleVersion check the version compatibility
func CheckCompatibleVersion(
	cfg config.Cassandra,
	expectedVersion string,
) error {

	client, err := NewCQLClient(&CQLClientConfig{
		Hosts:    cfg.Hosts,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Keyspace: cfg.Keyspace,
		Timeout:  DefaultTimeout,
		TLS:      cfg.TLS,
	})
	if err != nil {
		return fmt.Errorf("unable to create CQL Client: %v", err.Error())
	}
	defer client.Close()

	return schema.VerifyCompatibleVersion(client, cfg.Keyspace, expectedVersion)
}

// setupSchema executes the setupSchemaTask
// using the given command line arguments
// as input
func setupSchema(cli *cli.Context) error {
	config, err := newCQLClientConfig(cli)
	if err != nil {
		return handleErr(schema.NewConfigError(err.Error()))
	}
	client, err := NewCQLClient(config)
	if err != nil {
		return handleErr(err)
	}
	defer client.Close()
	if err := schema.Setup(cli, client); err != nil {
		return handleErr(err)
	}
	return nil
}

// updateSchema executes the updateSchemaTask
// using the given command lien args as input
func updateSchema(cli *cli.Context) error {
	config, err := newCQLClientConfig(cli)
	if err != nil {
		return handleErr(schema.NewConfigError(err.Error()))
	}
	client, err := NewCQLClient(config)
	if err != nil {
		return handleErr(err)
	}
	defer client.Close()
	if err := schema.Update(cli, client); err != nil {
		return handleErr(err)
	}
	return nil
}

// createKeyspace creates a cassandra Keyspace
func createKeyspace(cli *cli.Context) error {
	config, err := newCQLClientConfig(cli)
	if err != nil {
		return handleErr(schema.NewConfigError(err.Error()))
	}
	keyspace := cli.String(schema.CLIOptKeyspace)
	if keyspace == "" {
		return handleErr(schema.NewConfigError("missing " + flag(schema.CLIOptKeyspace) + " argument "))
	}
	err = doCreateKeyspace(*config, keyspace)
	if err != nil {
		return handleErr(fmt.Errorf("error creating Keyspace:%v", err))
	}
	return nil
}

func doCreateKeyspace(cfg CQLClientConfig, name string) error {
	cfg.Keyspace = SystemKeyspace
	client, err := NewCQLClient(&cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.CreateKeyspace(name)
}

func newCQLClientConfig(cli *cli.Context) (*CQLClientConfig, error) {
	cqlConfig := new(CQLClientConfig)
	cqlConfig.Hosts = cli.GlobalString(schema.CLIOptEndpoint)
	cqlConfig.Port = cli.GlobalInt(schema.CLIOptPort)
	cqlConfig.User = cli.GlobalString(schema.CLIOptUser)
	cqlConfig.Password = cli.GlobalString(schema.CLIOptPassword)
	cqlConfig.Timeout = cli.GlobalInt(schema.CLIOptTimeout)
	cqlConfig.Keyspace = cli.GlobalString(schema.CLIOptKeyspace)
	cqlConfig.NumReplicas = cli.Int(schema.CLIOptReplicationFactor)

	if cli.GlobalBool(schema.CLIFlagEnableTLS) {
		cqlConfig.TLS = &config.TLS{
			Enabled:                true,
			CertFile:               cli.GlobalString(schema.CLIFlagTLSCertFile),
			KeyFile:                cli.GlobalString(schema.CLIFlagTLSKeyFile),
			CaFile:                 cli.GlobalString(schema.CLIFlagTLSCaFile),
			EnableHostVerification: cli.GlobalBool(schema.CLIFlagTLSEnableHostVerification),
		}
	}

	if err := validateCQLClientConfig(cqlConfig); err != nil {
		return nil, err
	}
	return cqlConfig, nil
}

func validateCQLClientConfig(config *CQLClientConfig) error {
	if len(config.Hosts) == 0 {
		return schema.NewConfigError("missing cassandra endpoint argument " + flag(schema.CLIOptEndpoint))
	}
	if config.Keyspace == "" {
		return schema.NewConfigError("missing " + flag(schema.CLIOptKeyspace) + " argument ")
	}
	if config.Port == 0 {
		config.Port = DefaultCassandraPort
	}
	if config.NumReplicas == 0 {
		config.NumReplicas = defaultNumReplicas
	}

	return nil
}

func flag(opt string) string {
	return "(-" + opt + ")"
}

func handleErr(err error) error {
	log.Println(err)
	return err
}

func logErr(err error) {
	log.Println(err)
}
