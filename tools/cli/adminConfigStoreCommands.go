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

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

// AdminGetDynamicConfig gets value of specified dynamic config parameter matching specified filter
func AdminGetDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	filters := c.String(FlagDynamicConfigFilters)

	ctx, cancel := newContext(c)
	defer cancel()

	var parsedFilters []*types.DynamicConfigFilter
	if filters != "" {
		err := json.Unmarshal([]byte(filters), &parsedFilters)
		if err != nil {
			ErrorAndExit("Unable to parse filters string into type []*DynamicConfigFilter", err)
		}
	} else {
		parsedFilters = nil
	}

	req := &types.GetDynamicConfigRequest{
		ConfigName: dcName,
		Filters:    parsedFilters,
	}

	val, err := adminClient.GetDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to get dynamic config value", err)
	}

	prettyPrintJSONObject(val)
}

// AdminUpdateDynamicConfig updates specified dynamic config parameter with specified values
func AdminUpdateDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	dcValue := c.String(FlagDynamicConfigFilters)

	ctx, cancel := newContext(c)
	defer cancel()

	var parsedValues []*types.DynamicConfigValue
	if dcValue != "" {
		err := json.Unmarshal([]byte(dcValue), &parsedValues)
		if err != nil {
			ErrorAndExit("Unable to parse value string into type []*DynamicConfigValue", err)
		}
	} else {
		parsedValues = nil
	}

	req := &types.UpdateDynamicConfigRequest{
		ConfigName:   dcName,
		ConfigValues: parsedValues,
	}

	err := adminClient.UpdateDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to update dynamic config value", err)
	}
	fmt.Println("Dynamic Config %s updated", dcName)
}

// AdminRestoreDynamicConfig removes values of specified dynamic config parameter matching specified filter
func AdminRestoreDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := getRequiredOption(c, FlagDynamicConfigName)
	filters := c.String(FlagDynamicConfigFilters)

	ctx, cancel := newContext(c)
	defer cancel()

	var parsedFilters []*types.DynamicConfigFilter
	if filters != "" {
		err := json.Unmarshal([]byte(filters), &parsedFilters)
		if err != nil {
			ErrorAndExit("Unable to parse filters string into type []*DynamicConfigFilter", err)
		}
	} else {
		parsedFilters = nil
	}

	req := &types.RestoreDynamicConfigRequest{
		ConfigName: dcName,
		Filters:    parsedFilters,
	}

	err := adminClient.RestoreDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to restore dynamic config value", err)
	}
	fmt.Println("Dynamic Config %s restored", dcName)
}

// AdminListDynamicConfig lists all values associated with specified dynamic config parameter or all values for all dc parameter if none is specified.
func AdminListDynamicConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	dcName := c.String(FlagDynamicConfigName)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.ListDynamicConfigRequest{
		ConfigName: dcName,
	}

	val, err := adminClient.ListDynamicConfig(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to list dynamic config value(s)", err)
	}

	prettyPrintJSONObject(val)
}
