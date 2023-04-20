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

	"github.com/uber/cadence/common/types"

	"github.com/urfave/cli"
)

func AdminGetGlobalIsolationGroups(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.GetGlobalIsolationGroupsRequest{}
	igs, err := adminClient.GetGlobalIsolationGroups(ctx, req)
	if err != nil {
		ErrorAndExit("failed to get isolation-groups:", err)
	}
	prettyPrintJSONObject(igs.IsolationGroups.ToPartitionList())
}

func AdminUpdateGlobalIsolationGroups(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)
	cfgString := c.String(FlagIsolationGroupConfigurations)

	ctx, cancel := newContext(c)
	defer cancel()

	// Not using types.IsolationGroupConfiguration because the duplicate keys
	// for name is somewhat confusing to work with directly, instead expecting
	// a more IDL-like input of an array of isolation-groups
	var input []types.IsolationGroupPartition
	err := json.Unmarshal([]byte(cfgString), &input)
	if err != nil {
		ErrorAndExit(`failed to marshal input. Trying to marshal []types.IsolationGroupPartition

examples: 
- []                                    # will remove all isolation groups
- [{"Name": "zone-123", "State": 2}]    # drain zone-123
`, err)
	}

	req := types.IsolationGroupConfiguration{}
	for _, g := range input {
		req[g.Name] = g
	}

	_, err = adminClient.UpdateGlobalIsolationGroups(ctx, &types.UpdateGlobalIsolationGroupsRequest{
		IsolationGroups: req,
	})
	if err != nil {
		ErrorAndExit("failed to update isolation-groups", fmt.Errorf("used %#v, got %v", req, err))
	}
}

func AdminGetDomainIsolationGroups(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)
	domain := c.String(FlagDomain)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.GetDomainIsolationGroupsRequest{
		Domain: domain,
	}
	igs, err := adminClient.GetDomainIsolationGroups(ctx, req)
	if err != nil {
		ErrorAndExit("failed to get isolation-groups:", err)
	}
	prettyPrintJSONObject(igs.IsolationGroups.ToPartitionList())
}

func AdminUpdateDomainIsolationGroups(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)
	cfgString := c.String(FlagIsolationGroupConfigurations)
	domain := c.String(FlagDomain)

	ctx, cancel := newContext(c)
	defer cancel()

	var input []types.IsolationGroupPartition
	err := json.Unmarshal([]byte(cfgString), &input)
	if err != nil {
		ErrorAndExit(`failed to marshal input. Trying to marshal []types.IsolationGroupPartition

examples: 
- []                                    # will remove all isolation groups
- [{"Name": "zone-123", "State": 2}]    # drain zone-123
`, err)
	}

	req := types.IsolationGroupConfiguration{}
	for _, g := range input {
		req[g.Name] = g
	}

	_, err = adminClient.UpdateDomainIsolationGroups(ctx, &types.UpdateDomainIsolationGroupsRequest{
		Domain:          domain,
		IsolationGroups: req,
	})
	if err != nil {
		ErrorAndExit("failed to update isolation-groups", fmt.Errorf("used %#v, got %v", req, err))
	}
}
