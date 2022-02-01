// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

// AdminMaintainCorruptWorkflow deletes workflow from DB if it's corrupt
func AdminMaintainCorruptWorkflow(c *cli.Context) error {
	domainName := getRequiredGlobalOption(c, FlagDomain)
	workflowID := c.String(FlagWorkflowID)
	runID := c.String(FlagRunID)
	skipErrors := c.Bool(FlagSkipErrorMode)
	adminClient := cFactory.ServerAdminClient(c)

	request := &types.AdminDeleteWorkflowRequest{
		Domain: domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		SkipErrors: skipErrors,
	}

	ctx, cancel := newContext(c)
	defer cancel()
	err := adminClient.MaintainCorruptWorkflow(ctx, request)
	if err != nil {
		ErrorAndExit("Operation AdminMaintainCorruptWorkflow failed.", err)
	}

	return err
}
