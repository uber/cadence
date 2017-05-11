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
	"github.com/urfave/cli"
	"log"
)

// setupSchema executes the setupSchemaTask
// using the given command line arguments
// as input
func setupSchema(cli *cli.Context) error {
	config, err := newSetupSchemaConfig(cli)
	if err != nil {
		err = newConfigError(err.Error())
		log.Println(err)
		return err
	}
	if err := handleSetupSchema(config); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// updateSchema executes the updateSchemaTask
// using the given command lien args as input
func updateSchema(cli *cli.Context) error {
	config, err := newUpdateSchemaConfig(cli)
	if err != nil {
		err = newConfigError(err.Error())
		log.Println(err)
		return err
	}
	if err := handleUpdateSchema(config); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func handleUpdateSchema(config *UpdateSchemaConfig) error {
	task, err := NewUpdateSchemaTask(config)
	if err != nil {
		return fmt.Errorf("Error creating task, err=%v\n", err)
	}
	if err := task.run(); err != nil {
		return fmt.Errorf("Error setting up schema, err=%v\n", err)
	}
	return nil
}

func handleSetupSchema(config *SetupSchemaConfig) error {
	task, err := newSetupSchemaTask(config)
	if err != nil {
		return fmt.Errorf("Error creating task, err=%v\n", err)
	}
	if err := task.run(); err != nil {
		return fmt.Errorf("Error setting up schema, err=%v\n", err)
	}
	return nil
}

func validateSetupSchemaConfig(config *SetupSchemaConfig) error {
	if len(config.CassHosts) == 0 {
		return newConfigError("missing cassandra host")
	}
	if len(config.CassKeyspace) == 0 {
		return newConfigError("missing keyspace")
	}
	if len(config.SchemaFilePath) == 0 && config.DisableVersioning {
		return newConfigError("missing schemaFilePath")
	}
	if (config.DisableVersioning && len(config.InitialVersion) > 0) ||
		(!config.DisableVersioning && len(config.InitialVersion) == 0) {
		return newConfigError("either disableVersioning or initialVersion must be specified")
	}
	if !config.DisableVersioning {
		ver, err := parseValidateVersion(config.InitialVersion)
		if err != nil {
			return newConfigError("invalid intialVersion arg:" + err.Error())
		}
		config.InitialVersion = ver
	}
	return nil
}

func newSetupSchemaConfig(cli *cli.Context) (*SetupSchemaConfig, error) {

	config := new(SetupSchemaConfig)
	config.CassHosts = cli.GlobalString(cliOptEndpoint)
	config.CassKeyspace = cli.GlobalString(cliOptKeyspace)
	config.SchemaFilePath = cli.String(cliOptSchemaFile)
	config.InitialVersion = cli.String(cliOptVersion)
	config.DisableVersioning = cli.Bool(cliOptDisableVersioning)
	config.Overwrite = cli.Bool(cliOptOverwrite)

	if len(config.CassHosts) == 0 {
		return nil, fmt.Errorf("'%v' flag cannot be empty\n", cliOptEndpoint)
	}
	if len(config.CassKeyspace) == 0 {
		return nil, fmt.Errorf("'%v' flag cannot be empty\n", cliOptKeyspace)
	}
	if len(config.SchemaFilePath) == 0 && config.DisableVersioning {
		return nil, fmt.Errorf("both '%v' and '%v' flag cannot be empty\n", cliOptSchemaFile, cliOptDisableVersioning)
	}
	if config.DisableVersioning && len(config.InitialVersion) > 0 {
		return nil, fmt.Errorf("either specify '%v' or '%v', but not both", cliOptDisableVersioning, cliOptVersion)
	}
	if !config.DisableVersioning && len(config.InitialVersion) == 0 {
		return nil, fmt.Errorf("must specify a value for either '%v' or '%v'\n", cliOptDisableVersioning, cliOptVersion)
	}
	if !config.DisableVersioning {
		ver, err := parseValidateVersion(config.InitialVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %v flag: %v\n", cliOptVersion, err.Error())
		}
		config.InitialVersion = ver
	}
	return config, nil
}

func newUpdateSchemaConfig(cli *cli.Context) (*UpdateSchemaConfig, error) {

	config := new(UpdateSchemaConfig)
	config.CassHosts = cli.GlobalString(cliOptEndpoint)
	config.CassKeyspace = cli.GlobalString(cliOptKeyspace)
	config.SchemaDir = cli.String(cliOptSchemaDir)
	config.IsDryRun = cli.Bool(cliOptDryrun)
	config.TargetVersion = cli.String(cliOptTargetVersion)

	if len(config.CassHosts) == 0 {
		return nil, fmt.Errorf("'%v' flag cannot be empty\n", cliOptEndpoint)
	}
	if len(config.CassKeyspace) == 0 {
		return nil, fmt.Errorf("'%v' flag cannot be empty\n", cliOptKeyspace)
	}
	if len(config.SchemaDir) == 0 {
		return nil, fmt.Errorf("%v flag cannot be empty\n", cliOptSchemaDir)
	}
	if len(config.TargetVersion) > 0 {
		ver, err := parseValidateVersion(config.TargetVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %v flag:%v\n", cliOptTargetVersion, err.Error())
		}
		config.TargetVersion = ver
	}
	return config, nil
}
