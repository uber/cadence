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

	"github.com/fatih/color"
	"github.com/pborman/uuid"
	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/service/worker/failovermanager"

	"github.com/uber/cadence/common/types"
)

// An indirection for the prompt function so that it can be mocked in the unit tests
var promptFn = prompt

// AdminAddSearchAttribute to whitelist search attribute
func AdminAddSearchAttribute(c *cli.Context) {
	key := getRequiredOption(c, FlagSearchAttributesKey)
	valType := getRequiredIntOption(c, FlagSearchAttributesType)
	if !isValueTypeValid(valType) {
		ErrorAndExit("Unknown Search Attributes value type.", nil)
	}

	// ask user for confirmation
	promptMsg := fmt.Sprintf("Are you trying to add key [%s] with Type [%s]? Y/N",
		color.YellowString(key), color.YellowString(intValTypeToString(valType)))
	promptFn(promptMsg)

	adminClient := cFactory.ServerAdminClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	request := &types.AddSearchAttributeRequest{
		SearchAttribute: map[string]types.IndexedValueType{
			key: types.IndexedValueType(valType),
		},
		SecurityToken: c.String(FlagSecurityToken),
	}

	err := adminClient.AddSearchAttribute(ctx, request)
	if err != nil {
		ErrorAndExit("Add search attribute failed.", err)
	}
	fmt.Println("Success. Note that for a multil-node Cadence cluster, DynamicConfig MUST be updated separately to whitelist the new attributes.")
}

// AdminDescribeCluster is used to dump information about the cluster
func AdminDescribeCluster(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	ctx, cancel := newContext(c)
	defer cancel()
	response, err := adminClient.DescribeCluster(ctx)
	if err != nil {
		ErrorAndExit("Operation DescribeCluster failed.", err)
	}

	prettyPrintJSONObject(response)
}

func AdminRebalanceStart(c *cli.Context) {
	client := getCadenceClient(c)
	tcCtx, cancel := newContext(c)
	defer cancel()

	workflowID := failovermanager.RebalanceWorkflowID
	rbParams := &failovermanager.RebalanceParams{
		BatchFailoverSize:              100,
		BatchFailoverWaitTimeInSeconds: 10,
	}
	input, err := json.Marshal(rbParams)
	if err != nil {
		ErrorAndExit("Failed to serialize params for failover workflow", err)
	}
	memo, err := getWorkflowMemo(map[string]interface{}{
		common.MemoKeyForOperator: getOperator(),
	})
	if err != nil {
		ErrorAndExit("Failed to serialize memo", err)
	}
	request := &types.StartWorkflowExecutionRequest{
		Domain:                              common.SystemLocalDomainName,
		WorkflowID:                          workflowID,
		RequestID:                           uuid.New(),
		Identity:                            getCliIdentity(),
		WorkflowIDReusePolicy:               types.WorkflowIDReusePolicyAllowDuplicate.Ptr(),
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(60),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(int32(defaultDecisionTimeoutInSeconds)),
		Input:                               input,
		TaskList: &types.TaskList{
			Name: failovermanager.TaskListName,
		},
		Memo: memo,
		WorkflowType: &types.WorkflowType{
			Name: failovermanager.RebalanceWorkflowTypeName,
		},
	}

	resp, err := client.StartWorkflowExecution(tcCtx, request)
	if err != nil {
		ErrorAndExit("Failed to start failover workflow", err)
	}
	fmt.Println("Rebalance workflow started")
	fmt.Println("wid: " + workflowID)
	fmt.Println("rid: " + resp.GetRunID())
}

func AdminRebalanceList(c *cli.Context) {
	c.Set(FlagWorkflowID, failovermanager.RebalanceWorkflowID)
	c.GlobalSet(FlagDomain, common.SystemLocalDomainName)
	ListWorkflow(c)
}

func intValTypeToString(valType int) string {
	switch valType {
	case 0:
		return "String"
	case 1:
		return "Keyword"
	case 2:
		return "Int"
	case 3:
		return "Double"
	case 4:
		return "Bool"
	case 5:
		return "Datetime"
	default:
		return ""
	}
}

func isValueTypeValid(valType int) bool {
	return valType >= 0 && valType <= 5
}
