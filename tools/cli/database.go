// Copyright (c) 2022 Uber Technologies, Inc.
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

package cli

import (
	"fmt"
	"net"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/client"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/tools/common/flag"
)

var supportedDBs = append(sql.GetRegisteredPluginNames(), "cassandra")

func getDBFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   FlagServiceConfigDirWithAlias,
			Value:  "config",
			Usage:  "service configuration dir",
			EnvVar: config.EnvKeyConfigDir,
		},
		cli.StringFlag{
			Name:   FlagServiceEnvWithAlias,
			Usage:  "service env for loading service configuration",
			EnvVar: config.EnvKeyEnvironment,
		},
		cli.StringFlag{
			Name:   FlagServiceZoneWithAlias,
			Usage:  "service zone for loading service configuration",
			EnvVar: config.EnvKeyAvailabilityZone,
		},
		cli.StringFlag{
			Name:  FlagDBType,
			Value: "cassandra",
			Usage: fmt.Sprintf("persistence type. Current supported options are %v", supportedDBs),
		},
		cli.StringFlag{
			Name:  FlagDBAddress,
			Value: "127.0.0.1",
			Usage: "persistence address (right now only cassandra is fully supported)",
		},
		cli.IntFlag{
			Name:  FlagDBPort,
			Value: 9042,
			Usage: "persistence port",
		},
		cli.StringFlag{
			Name:  FlagDBRegion,
			Usage: "persistence region",
		},
		cli.StringFlag{
			Name:  FlagUsername,
			Usage: "persistence username",
		},
		cli.StringFlag{
			Name:  FlagPassword,
			Usage: "persistence password",
		},
		cli.StringFlag{
			Name:  FlagKeyspace,
			Value: "cadence",
			Usage: "cassandra keyspace",
		},
		cli.StringFlag{
			Name:  FlagDatabaseName,
			Value: "cadence",
			Usage: "sql database name",
		},
		cli.StringFlag{
			Name:  FlagEncodingType,
			Value: "thriftrw",
			Usage: "sql database encoding type",
		},
		cli.StringSliceFlag{
			Name: FlagDecodingTypes,
			Value: &cli.StringSlice{
				"thriftrw",
			},
			Usage: "sql database decoding types",
		},
		cli.IntFlag{
			Name:  FlagProtoVersion,
			Value: 4,
			Usage: "cassandra protocol version",
		},
		cli.BoolFlag{
			Name:  FlagEnableTLS,
			Usage: "enable TLS over cassandra connection",
		},
		cli.StringFlag{
			Name:  FlagTLSCertPath,
			Usage: "cassandra tls client cert path (tls must be enabled)",
		},
		cli.StringFlag{
			Name:  FlagTLSKeyPath,
			Usage: "cassandra tls client key path (tls must be enabled)",
		},
		cli.StringFlag{
			Name:  FlagTLSCaPath,
			Usage: "cassandra tls client ca path (tls must be enabled)",
		},
		cli.BoolFlag{
			Name:  FlagTLSEnableHostVerification,
			Usage: "cassandra tls verify hostname and server cert (tls must be enabled)",
		},
		cli.GenericFlag{
			Name:  FlagConnectionAttributes,
			Usage: "a key-value set of sql database connection attributes (must be in key1=value1,key2=value2,...,keyN=valueN format, e.g. cluster=dca or cluster=dca,instance=cadence)",
			Value: &flag.StringMap{},
		},
		cli.IntFlag{
			Name:  FlagRPS,
			Usage: "target rps of database queries",
			Value: 100,
		},
	}
}

func initializeExecutionStore(c *cli.Context, shardID int) persistence.ExecutionManager {
	factory := getPersistenceFactory(c)
	historyManager, err := factory.NewExecutionManager(shardID)
	if err != nil {
		ErrorAndExit("Failed to initialize history manager", err)
	}
	return historyManager
}

func initializeHistoryManager(c *cli.Context) persistence.HistoryManager {
	factory := getPersistenceFactory(c)
	historyManager, err := factory.NewHistoryManager()
	if err != nil {
		ErrorAndExit("Failed to initialize history manager", err)
	}
	return historyManager
}

func initializeShardManager(c *cli.Context) persistence.ShardManager {
	factory := getPersistenceFactory(c)
	shardManager, err := factory.NewShardManager()
	if err != nil {
		ErrorAndExit("Failed to initialize shard manager", err)
	}
	return shardManager
}

func initializeDomainManager(c *cli.Context) persistence.DomainManager {
	factory := getPersistenceFactory(c)
	domainManager, err := factory.NewDomainManager()
	if err != nil {
		ErrorAndExit("Failed to initialize domain manager", err)
	}
	return domainManager
}

var persistenceFactory client.Factory

func getPersistenceFactory(c *cli.Context) client.Factory {
	if persistenceFactory == nil {
		persistenceFactory = initPersistenceFactory(c)
	}
	return persistenceFactory
}

func initPersistenceFactory(c *cli.Context) client.Factory {
	cfg, err := cFactory.ServerConfig(c)

	if err != nil {
		cfg = &config.Config{
			Persistence: config.Persistence{
				DefaultStore: "default",
				DataStores: map[string]config.DataStore{
					"default": {NoSQL: &config.NoSQL{PluginName: cassandra.PluginName}},
				},
			},
			ClusterGroupMetadata: &config.ClusterGroupMetadata{
				CurrentClusterName: "current-cluster",
			},
		}
	}

	// If there are any overrides provided via CLI flags, apply them here
	defaultStore := cfg.Persistence.DataStores[cfg.Persistence.DefaultStore]
	defaultStore = overrideDataStore(c, defaultStore)
	cfg.Persistence.DataStores[cfg.Persistence.DefaultStore] = defaultStore

	cfg.Persistence.TransactionSizeLimit = dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit)
	cfg.Persistence.ErrorInjectionRate = dynamicconfig.GetFloatPropertyFn(0.0)

	rps := c.Float64(FlagRPS)

	return client.NewFactory(
		&cfg.Persistence,
		func() float64 { return rps },
		cfg.ClusterGroupMetadata.CurrentClusterName,
		metrics.NewNoopMetricsClient(),
		log.NewNoop(),
	)
}

func overrideDataStore(c *cli.Context, ds config.DataStore) config.DataStore {
	if c.IsSet(FlagDBType) {
		// overriding DBType will wipe out all settings, everything will be set from flags only
		ds = createDataStore(c)
	}

	if ds.NoSQL != nil {
		overrideNoSQLDataStore(c, ds.NoSQL)
	}
	if ds.SQL != nil {
		overrideSQLDataStore(c, ds.SQL)
	}

	return ds
}

func createDataStore(c *cli.Context) config.DataStore {
	dbType := c.String(FlagDBType)
	switch dbType {
	case cassandra.PluginName:
		return config.DataStore{NoSQL: &config.NoSQL{PluginName: cassandra.PluginName}}
	default:
		if sql.PluginRegistered(dbType) {
			return config.DataStore{SQL: &config.SQL{PluginName: dbType}}
		}
	}
	ErrorAndExit(fmt.Sprintf("The DB type is not supported. Options are: %s.", supportedDBs), nil)
	return config.DataStore{}
}

func overrideNoSQLDataStore(c *cli.Context, cfg *config.NoSQL) {
	if c.IsSet(FlagDBAddress) || cfg.Hosts == "" {
		cfg.Hosts = c.String(FlagDBAddress)
	}
	if c.IsSet(FlagDBPort) || cfg.Port == 0 {
		cfg.Port = c.Int(FlagDBPort)
	}
	if c.IsSet(FlagDBRegion) || cfg.Region == "" {
		cfg.Region = c.String(FlagDBRegion)
	}
	if c.IsSet(FlagUsername) || cfg.User == "" {
		cfg.User = c.String(FlagUsername)
	}
	if c.IsSet(FlagPassword) || cfg.Password == "" {
		cfg.Password = c.String(FlagPassword)
	}
	if c.IsSet(FlagKeyspace) || cfg.Keyspace == "" {
		cfg.Keyspace = c.String(FlagKeyspace)
	}
	if c.IsSet(FlagProtoVersion) || cfg.ProtoVersion == 0 {
		cfg.ProtoVersion = c.Int(FlagProtoVersion)
	}

	if cfg.TLS == nil {
		cfg.TLS = &config.TLS{}
	}
	overrideTLS(c, cfg.TLS)
}

func overrideSQLDataStore(c *cli.Context, cfg *config.SQL) {
	host, port, _ := net.SplitHostPort(cfg.ConnectAddr)
	if c.IsSet(FlagDBAddress) || cfg.ConnectAddr == "" {
		host = c.String(FlagDBAddress)
	}
	if c.IsSet(FlagDBPort) || cfg.ConnectAddr == "" {
		port = c.String(FlagDBPort)
	}
	cfg.ConnectAddr = net.JoinHostPort(host, port)

	if c.IsSet(FlagDBType) || cfg.PluginName == "" {
		cfg.PluginName = c.String(FlagDBType)
	}
	if c.IsSet(FlagUsername) || cfg.User == "" {
		cfg.User = c.String(FlagUsername)
	}
	if c.IsSet(FlagPassword) || cfg.Password == "" {
		cfg.Password = c.String(FlagPassword)
	}
	if c.IsSet(FlagDatabaseName) || cfg.DatabaseName == "" {
		cfg.DatabaseName = c.String(FlagDatabaseName)
	}
	if c.IsSet(FlagEncodingType) || cfg.EncodingType == "" {
		cfg.EncodingType = c.String(FlagEncodingType)
	}
	if c.IsSet(FlagDecodingTypes) || len(cfg.DecodingTypes) == 0 {
		cfg.DecodingTypes = c.StringSlice(FlagDecodingTypes)
	}
	if c.IsSet(FlagConnectionAttributes) || cfg.ConnectAttributes == nil {
		cfg.ConnectAttributes = c.Generic(FlagConnectionAttributes).(*flag.StringMap).Value()
	}

	if cfg.TLS == nil {
		cfg.TLS = &config.TLS{}
	}
	overrideTLS(c, cfg.TLS)
}

func overrideTLS(c *cli.Context, tls *config.TLS) {
	tls.Enabled = c.Bool(FlagEnableTLS)

	if c.IsSet(FlagTLSCertPath) {
		tls.CertFile = c.String(FlagTLSCertPath)
	}
	if c.IsSet(FlagTLSKeyPath) {
		tls.KeyFile = c.String(FlagTLSKeyPath)
	}
	if c.IsSet(FlagTLSCaPath) {
		tls.CaFile = c.String(FlagTLSCaPath)
	}
	if c.IsSet(FlagTLSEnableHostVerification) {
		tls.EnableHostVerification = c.Bool(FlagTLSEnableHostVerification)
	}
}
