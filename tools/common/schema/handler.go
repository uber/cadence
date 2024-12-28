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
	"os"

	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
)

// VerifyCompatibleVersion ensures that the installed version is greater than or equal to the expected version.
func VerifyCompatibleVersion(
	db SchemaClient,
	dbName string,
	expectedVersion string,
) error {

	version, err := db.ReadSchemaVersion()
	if err != nil {
		return fmt.Errorf("reading schema version keyspace/database: %q, error: %w", dbName, err)
	}
	// In most cases, the versions should match. However if after a schema upgrade there is a code
	// rollback, the code version (expected version) would fall lower than the actual version in
	// cassandra. This check is to allow such rollbacks since we only make backwards compatible schema
	// changes
	if cmpVersion(version, expectedVersion) < 0 {
		return fmt.Errorf(
			"version mismatch for keyspace/database: %q. Expected version: %s cannot be greater than "+
				"Actual version: %s", dbName, expectedVersion, version,
		)
	}
	return nil
}

// SetupFromConfig sets up schema tables based on the given config
func SetupFromConfig(config *SetupConfig, logger log.Logger, db SchemaClient) error {
	if err := validateSetupConfig(config); err != nil {
		return err
	}
	return newSetupSchemaTask(db, logger, config).Run()
}

// Setup sets up schema tables
func Setup(cli *cli.Context, db SchemaClient) error {
	cfg, logger, err := newSetupConfig(cli)
	if err != nil {
		return err
	}
	return SetupFromConfig(cfg, logger, db)
}

// UpdateFromConfig updates the schema for the specified database based on the given config
func UpdateFromConfig(config *UpdateConfig, logger log.Logger, db SchemaClient) error {
	if err := validateUpdateConfig(config); err != nil {
		return err
	}
	return NewUpdateSchemaTask(db, logger, config).Run()
}

// Update updates the schema for the specified database
func Update(cli *cli.Context, db SchemaClient) error {
	cfg, logger, err := newUpdateConfig(cli)
	if err != nil {
		return err
	}
	return UpdateFromConfig(cfg, logger, db)
}

func newUpdateConfig(cli *cli.Context) (*UpdateConfig, log.Logger, error) {
	config := new(UpdateConfig)
	schemaDir := cli.String(CLIOptSchemaDir)
	if len(schemaDir) == 0 {
		return nil, nil, NewConfigError("missing " + flag(CLIOptSchemaDir) + " argument ")
	}
	config.SchemaFS = os.DirFS(schemaDir)
	config.IsDryRun = cli.Bool(CLIOptDryrun)
	config.TargetVersion = cli.String(CLIOptTargetVersion)

	logger, err := loggerimpl.NewDevelopment()
	if err != nil {
		return nil, nil, fmt.Errorf("build logger: %w", err)
	}

	return config, logger, nil
}

func newSetupConfig(cli *cli.Context) (*SetupConfig, log.Logger, error) {
	config := new(SetupConfig)
	config.SchemaFilePath = cli.String(CLIOptSchemaFile)
	config.InitialVersion = cli.String(CLIOptVersion)
	config.DisableVersioning = cli.Bool(CLIOptDisableVersioning)
	config.Overwrite = cli.Bool(CLIOptOverwrite)

	logger, err := loggerimpl.NewDevelopment()
	if err != nil {
		return nil, nil, fmt.Errorf("build logger: %w", err)
	}

	return config, logger, nil
}

func validateSetupConfig(config *SetupConfig) error {
	if len(config.SchemaFilePath) == 0 && config.DisableVersioning {
		return NewConfigError("missing schemaFilePath " + flag(CLIOptSchemaFile))
	}
	if (config.DisableVersioning && len(config.InitialVersion) > 0) ||
		(!config.DisableVersioning && len(config.InitialVersion) == 0) {
		return NewConfigError("either " + flag(CLIOptDisableVersioning) + " or " +
			flag(CLIOptVersion) + " but not both must be specified")
	}
	if !config.DisableVersioning {
		ver, err := parseValidateVersion(config.InitialVersion)
		if err != nil {
			return NewConfigError("invalid " + flag(CLIOptVersion) + " argument:" + err.Error())
		}
		config.InitialVersion = ver
	}
	return nil
}

func validateUpdateConfig(config *UpdateConfig) error {
	if config.SchemaFS == nil {
		return NewConfigError("schema file system is not set")
	}
	if len(config.TargetVersion) > 0 {
		ver, err := parseValidateVersion(config.TargetVersion)
		if err != nil {
			return NewConfigError("invalid " + flag(CLIOptTargetVersion) + " argument:" + err.Error())
		}
		config.TargetVersion = ver
	}
	return nil
}

func flag(opt string) string {
	return "(-" + opt + ")"
}
