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

package schema

import (
	"fmt"
	"regexp"
)

type (
	// ConfigError is an error type that
	// represents a problem with the config
	ConfigError struct {
		msg string
	}
	// UpdateConfig holds the config
	// params for executing a UpdateTask
	UpdateConfig struct {
		DBName        string
		TargetVersion string
		SchemaDir     string
		IsDryRun      bool
	}
	// SetupConfig holds the config
	// params need by the SetupTask
	SetupConfig struct {
		SchemaFilePath    string
		InitialVersion    string
		Overwrite         bool // overwrite previous data
		DisableVersioning bool // do not use schema versioning
	}
	// DB is the database interface that's required to be implemented
	// for the schema-tool to work
	DB interface {
		// Exec executes a cql statement
		Exec(stmt string, args ...interface{}) error
		// DropAllTables drops all tables
		DropAllTables() error
		// CreateSchemaVersionTables sets up the schema version tables
		CreateSchemaVersionTables() error
		// ReadSchemaVersion returns the current schema version for the keyspace
		ReadSchemaVersion() (string, error)
		// UpdateSchemaVersion updates the schema version for the keyspace
		UpdateSchemaVersion(newVersion string, minCompatibleVersion string) error
		// WriteSchemaUpdateLog adds an entry to the schema update history table
		WriteSchemaUpdateLog(oldVersion string, newVersion string, manifestMD5 string, desc string) error
		// Close gracefully closes the client object
		Close()
	}
)

const (
	// CLIFlagEndpoint is the cli option for endpoint
	CLIFlagEndpoint = "endpoint"
	// CLIFlagPort is the cli option for port
	CLIFlagPort = "port"
	// CLIFlagUser is the cli option for user
	CLIFlagUser = "user"
	// CLIFlagPassword is the cli option for password
	CLIFlagPassword = "password"
	// CLIFlagTimeout is the cli option for timeout
	CLIFlagTimeout = "timeout"
	// CLIFlagKeyspace is the cli option for keyspace
	CLIFlagKeyspace = "keyspace"
	// CLIFlagDatabase is the cli option for database
	CLIFlagDatabase = "database"
	// CLIFlagPluginName is the cli option for plugin name
	CLIFlagPluginName = "plugin"
	// CLIFlagVersion is the cli option for version
	CLIFlagVersion = "version"
	// CLIFlagSchemaFile is the cli option for schema file
	CLIFlagSchemaFile = "schema-file"
	// CLIFlagOverwrite is the cli option for overwrite
	CLIFlagOverwrite = "overwrite"
	// CLIFlagDisableVersioning is the cli option to disabling versioning
	CLIFlagDisableVersioning = "disable-versioning"
	// CLIFlagTargetVersion is the cli option for target version
	CLIFlagTargetVersion = "version"
	// CLIFlagDryrun is the cli option for enabling dryrun
	CLIFlagDryrun = "dryrun"
	// CLIFlagSchemaDir is the cli option for schema directory
	CLIFlagSchemaDir = "schema-dir"
	// CLIFlagReplicationFactor is the cli option for replication factor
	CLIFlagReplicationFactor = "replication-factor"
	// CLIFlagQuiet is the cli option for quiet mode
	CLIFlagQuiet = "quiet"

	// CLIFlagEnableTLS enables cassandra client TLS
	CLIFlagEnableTLS = "tls"
	// CLIFlagTLSCertFile is the optional tls cert file (tls must be enabled)
	CLIFlagTLSCertFile = "tls-cert-file"
	// CLIFlagTLSKeyFile is the optional tls key file (tls must be enabled)
	CLIFlagTLSKeyFile = "tls-key-file"
	// CLIFlagTLSCaFile is the optional tls CA file (tls must be enabled)
	CLIFlagTLSCaFile = "tls-ca-file"
	// CLIFlagTLSEnableHostVerification enables tls host verification (tls must be enabled)
	CLIFlagTLSEnableHostVerification = "tls-enable-host-verification"
)

var (

	// CLIFlagEndpointAlias is the cli flag for endpoint
	CLIFlagEndpointAlias = []string{"ep"}
	// CLIFlagPortAlias is the cli flag for port
	CLIFlagPortAlias = []string{"p"}
	// CLIFlagUserAlias is the cli flag for user
	CLIFlagUserAlias = []string{"u"}
	// CLIFlagPasswordAlias is the cli flag for password
	CLIFlagPasswordAlias = []string{"pw"}
	// CLIFlagTimeoutAlias is the cli flag for timeout
	CLIFlagTimeoutAlias = []string{"t"}
	// CLIFlagKeyspaceAlias is the cli flag for keyspace
	CLIFlagKeyspaceAlias = []string{"k"}
	// CLIFlagDatabaseAlias is the cli flag for database
	CLIFlagDatabaseAlias = []string{"db"}
	// CLIFlagPluginNameAlias is the cli flag for plugin name
	CLIFlagPluginNameAlias = []string{"pl"}
	// CLIFlagVersionAlias is the cli flag for version
	CLIFlagVersionAlias = []string{"v"}
	// CLIFlagSchemaFileAlias is the cli flag for schema file
	CLIFlagSchemaFileAlias = []string{"f"}
	// CLIFlagOverwriteAlias is the cli flag for overwrite
	CLIFlagOverwriteAlias = []string{"o"}
	// CLIFlagDisableVersioningAlias is the cli flag for disabling versioning
	CLIFlagDisableVersioningAlias = []string{"d"}
	// CLIFlagTargetVersionAlias is the cli flag for target version
	CLIFlagTargetVersionAlias = []string{"v"}
	// CLIFlagDryrunAlias is the cli flag for dryrun
	CLIFlagDryrunAlias = []string{"y"}
	// CLIFlagSchemaDirAlias is the cli flag for schema directory
	CLIFlagSchemaDirAlias = []string{"d"}
	// CLIFlagReplicationFactorAlias is the cli flag for replication factor
	CLIFlagReplicationFactorAlias = []string{"rf"}
	// CLIFlagQuietAlias is the cli flag for quiet mode
	CLIFlagQuietAlias = []string{"q"}
)

// DryrunDBName is the db name used for dryrun
const DryrunDBName = "_cadence_dryrun_"

var rmspaceRegex = regexp.MustCompile("\\s+")

// NewConfigError creates and returns an instance of ConfigError
func NewConfigError(msg string) error {
	return &ConfigError{msg: msg}
}

// Error returns a string representation of this error
func (e *ConfigError) Error() string {
	return fmt.Sprintf("Config Error:%v", e.msg)
}
