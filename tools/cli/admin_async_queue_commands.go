// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

func AdminGetAsyncWFConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	domainName := getRequiredOption(c, FlagDomain)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.GetDomainAsyncWorkflowConfiguratonRequest{
		Domain: domainName,
	}

	resp, err := adminClient.GetDomainAsyncWorkflowConfiguraton(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to get async wf queue config", err)
	}

	if resp == nil || resp.Configuration == nil {
		fmt.Printf("Async workflow queue config not found for domain %s\n", domainName)
		return
	}

	fmt.Printf("Async workflow queue config for domain %s:\n", domainName)
	prettyPrintJSONObject(resp.Configuration)
}

func AdminUpdateAsyncWFConfig(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	domainName := getRequiredOption(c, FlagDomain)
	asyncWFCfgJSON := getRequiredOption(c, FlagJSON)

	var cfg types.AsyncWorkflowConfiguration
	err := json.Unmarshal([]byte(asyncWFCfgJSON), &cfg)
	if err != nil {
		ErrorAndExit("Failed to parse async workflow config", err)
	}

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
		Domain:        domainName,
		Configuration: &cfg,
	}

	_, err = adminClient.UpdateDomainAsyncWorkflowConfiguraton(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to update async workflow queue config", err)
	}

	fmt.Printf("Successfully updated async workflow queue config for domain %s\n", domainName)
}
