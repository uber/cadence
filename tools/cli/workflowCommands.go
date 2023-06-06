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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pborman/uuid"
	"github.com/urfave/cli"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
)

// RestartWorkflow restarts a workflow execution
func RestartWorkflow(c *cli.Context) {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := wfClient.RestartWorkflowExecution(
		ctx,
		&types.RestartWorkflowExecutionRequest{
			Domain: domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			}, Identity: getCliIdentity(),
		},
	)

	if err != nil {
		ErrorAndExit("Restart workflow failed.", err)
	} else {
		fmt.Printf("Restarted Workflow Id: %s, run Id: %s\n", wid, resp.GetRunID())
	}
}

// ShowHistory shows the history of given workflow execution based on workflowID and runID.
func ShowHistory(c *cli.Context) {
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	showHistoryHelper(c, wid, rid)
}

// ShowHistoryWithWID shows the history of given workflow with workflow_id
func ShowHistoryWithWID(c *cli.Context) {
	if !c.Args().Present() {
		ErrorAndExit("Argument workflow_id is required.", nil)
	}
	wid := c.Args().First()
	rid := ""
	if c.NArg() >= 2 {
		rid = c.Args().Get(1)
	}
	showHistoryHelper(c, wid, rid)
}

func showHistoryHelper(c *cli.Context, wid, rid string) {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	printDateTime := c.Bool(FlagPrintDateTime)
	printRawTime := c.Bool(FlagPrintRawTime)
	printFully := c.Bool(FlagPrintFullyDetail)
	printVersion := c.Bool(FlagPrintEventVersion)
	outputFileName := c.String(FlagOutputFilename)
	var maxFieldLength int
	if c.IsSet(FlagMaxFieldLength) || !printFully {
		maxFieldLength = c.Int(FlagMaxFieldLength)
	}
	resetPointsOnly := c.Bool(FlagResetPointsOnly)

	ctx, cancel := newContext(c)
	defer cancel()
	history, err := GetHistory(ctx, wfClient, domain, wid, rid)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Failed to get history on workflow id: %s, run id: %s.", wid, rid), err)
	}

	prevEvent := types.HistoryEvent{}
	if printFully { // dump everything
		for _, e := range history.Events {
			if resetPointsOnly {
				if prevEvent.GetEventType() != types.EventTypeDecisionTaskStarted {
					prevEvent = *e
					continue
				}
				prevEvent = *e
			}
			fmt.Println(anyToString(e, true, maxFieldLength))
		}
	} else if c.IsSet(FlagEventID) { // only dump that event
		eventID := c.Int(FlagEventID)
		if eventID <= 0 || eventID > len(history.Events) {
			ErrorAndExit("EventId out of range.", fmt.Errorf("number should be 1 - %d inclusive", len(history.Events)))
		}
		e := history.Events[eventID-1]
		fmt.Println(anyToString(e, true, 0))
	} else { // use table to pretty output, will trim long text
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorder(false)
		table.SetColumnSeparator("")
		for _, e := range history.Events {
			if resetPointsOnly {
				if prevEvent.GetEventType() != types.EventTypeDecisionTaskStarted {
					prevEvent = *e
					continue
				}
				prevEvent = *e
			}

			columns := []string{}
			columns = append(columns, strconv.FormatInt(e.ID, 10))

			if printRawTime {
				columns = append(columns, strconv.FormatInt(e.GetTimestamp(), 10))
			} else if printDateTime {
				columns = append(columns, convertTime(e.GetTimestamp(), false))
			}
			if printVersion {
				columns = append(columns, fmt.Sprintf("(Version: %v)", e.Version))
			}

			columns = append(columns, ColorEvent(e), HistoryEventToString(e, false, maxFieldLength))
			table.Append(columns)
		}
		table.Render()
	}

	if outputFileName != "" {
		serializer := &JSONHistorySerializer{}
		data, err := serializer.Serialize(history)
		if err != nil {
			ErrorAndExit("Failed to serialize history data.", err)
		}
		if err := ioutil.WriteFile(outputFileName, data, 0666); err != nil {
			ErrorAndExit("Failed to export history data file.", err)
		}
	}

	// finally append activities with retry
	frontendClient := cFactory.ServerFrontendClient(c)
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
	})
	if err != nil {
		if _, ok := err.(*types.EntityNotExistsError); ok {
			fmt.Printf("%s %s\n", colorRed("Error:"), err)
			return
		}
		ErrorAndExit("Describe workflow execution failed, cannot get information of pending activities", err)
	}
	fmt.Println("History Source: Default Storage")

	descOutput := convertDescribeWorkflowExecutionResponse(resp, frontendClient, c)
	if len(descOutput.PendingActivities) > 0 {
		fmt.Println("============Workflow Pending activities============")
		prettyPrintJSONObject(descOutput.PendingActivities)
		fmt.Println("NOTE: ActivityStartedEvent with retry policy will be written into history when the activity is finished.")
	}

}

// StartWorkflow starts a new workflow execution
func StartWorkflow(c *cli.Context) {
	startWorkflowHelper(c, false)
}

// RunWorkflow starts a new workflow execution and print workflow progress and result
func RunWorkflow(c *cli.Context) {
	startWorkflowHelper(c, true)
}

func startWorkflowHelper(c *cli.Context, shouldPrintProgress bool) {
	serviceClient := cFactory.ServerFrontendClient(c)

	startRequest := constructStartWorkflowRequest(c)
	domain := startRequest.GetDomain()
	wid := startRequest.GetWorkflowID()
	workflowType := startRequest.WorkflowType.GetName()
	taskList := startRequest.TaskList.GetName()
	input := string(startRequest.Input)

	startFn := func() {
		tcCtx, cancel := newContext(c)
		defer cancel()
		resp, err := serviceClient.StartWorkflowExecution(tcCtx, startRequest)

		if err != nil {
			ErrorAndExit("Failed to create workflow.", err)
		} else {
			fmt.Printf("Started Workflow Id: %s, run Id: %s\n", wid, resp.GetRunID())
		}
	}

	runFn := func() {
		tcCtx, cancel := newContextForLongPoll(c)
		defer cancel()
		resp, err := serviceClient.StartWorkflowExecution(tcCtx, startRequest)

		if err != nil {
			ErrorAndExit("Failed to run workflow.", err)
		}

		// print execution summary
		fmt.Println(colorMagenta("Running execution:"))
		table := tablewriter.NewWriter(os.Stdout)
		executionData := [][]string{
			{"Workflow Id", wid},
			{"Run Id", resp.GetRunID()},
			{"Type", workflowType},
			{"Domain", domain},
			{"Task List", taskList},
			{"Args", truncate(input)}, // in case of large input
		}
		table.SetBorder(false)
		table.SetColumnSeparator(":")
		table.AppendBulk(executionData) // Add Bulk Data
		table.Render()

		printWorkflowProgress(c, domain, wid, resp.GetRunID())
	}

	if shouldPrintProgress {
		runFn()
	} else {
		startFn()
	}
}

func constructStartWorkflowRequest(c *cli.Context) *types.StartWorkflowExecutionRequest {
	domain := getRequiredGlobalOption(c, FlagDomain)
	taskList := getRequiredOption(c, FlagTaskList)
	workflowType := getRequiredOption(c, FlagWorkflowType)
	et := c.Int(FlagExecutionTimeout)
	if et == 0 {
		ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagExecutionTimeout), nil)
	}
	dt := c.Int(FlagDecisionTimeout)
	wid := c.String(FlagWorkflowID)
	if len(wid) == 0 {
		wid = uuid.New()
	}
	reusePolicy := defaultWorkflowIDReusePolicy.Ptr()
	if c.IsSet(FlagWorkflowIDReusePolicy) {
		reusePolicy = getWorkflowIDReusePolicy(c.Int(FlagWorkflowIDReusePolicy))
	}

	input := processJSONInput(c)
	startRequest := &types.StartWorkflowExecutionRequest{
		RequestID:  uuid.New(),
		Domain:     domain,
		WorkflowID: wid,
		WorkflowType: &types.WorkflowType{
			Name: workflowType,
		},
		TaskList: &types.TaskList{
			Name: taskList,
		},
		Input:                               []byte(input),
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(int32(et)),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(int32(dt)),
		Identity:                            getCliIdentity(),
		WorkflowIDReusePolicy:               reusePolicy,
	}
	if c.IsSet(FlagCronSchedule) {
		startRequest.CronSchedule = c.String(FlagCronSchedule)
	}

	if c.IsSet(FlagRetryAttempts) || c.IsSet(FlagRetryExpiration) {
		startRequest.RetryPolicy = &types.RetryPolicy{
			InitialIntervalInSeconds: int32(c.Int(FlagRetryInterval)),
			BackoffCoefficient:       c.Float64(FlagRetryBackoff),
		}

		if c.IsSet(FlagRetryAttempts) {
			startRequest.RetryPolicy.MaximumAttempts = int32(c.Int(FlagRetryAttempts))
		}
		if c.IsSet(FlagRetryExpiration) {
			startRequest.RetryPolicy.ExpirationIntervalInSeconds = int32(c.Int(FlagRetryExpiration))
		}
		if c.IsSet(FlagRetryMaxInterval) {
			startRequest.RetryPolicy.MaximumIntervalInSeconds = int32(c.Int(FlagRetryMaxInterval))
		}
	}

	if c.IsSet(DelayStartSeconds) {
		startRequest.DelayStartSeconds = common.Int32Ptr(int32(c.Int(DelayStartSeconds)))
	}

	if c.IsSet(JitterStartSeconds) {
		startRequest.JitterStartSeconds = common.Int32Ptr(int32(c.Int(JitterStartSeconds)))
	}

	headerFields := processHeader(c)
	if len(headerFields) != 0 {
		startRequest.Header = &types.Header{Fields: headerFields}
	}

	memoFields := processMemo(c)
	if len(memoFields) != 0 {
		startRequest.Memo = &types.Memo{Fields: memoFields}
	}

	searchAttrFields := processSearchAttr(c)
	if len(searchAttrFields) != 0 {
		startRequest.SearchAttributes = &types.SearchAttributes{IndexedFields: searchAttrFields}
	}

	return startRequest
}

func processSearchAttr(c *cli.Context) map[string][]byte {
	rawSearchAttrKey := c.String(FlagSearchAttributesKey)
	var searchAttrKeys []string
	if strings.TrimSpace(rawSearchAttrKey) != "" {
		searchAttrKeys = trimSpace(strings.Split(rawSearchAttrKey, searchAttrInputSeparator))
	}

	rawSearchAttrVal := c.String(FlagSearchAttributesVal)
	var searchAttrVals []interface{}
	if strings.TrimSpace(rawSearchAttrVal) != "" {
		searchAttrValsStr := trimSpace(strings.Split(rawSearchAttrVal, searchAttrInputSeparator))

		for _, v := range searchAttrValsStr {
			searchAttrVals = append(searchAttrVals, convertStringToRealType(v))
		}
	}

	if len(searchAttrKeys) != len(searchAttrVals) {
		ErrorAndExit("Number of search attributes keys and values are not equal.", nil)
	}

	fields := map[string][]byte{}
	for i, key := range searchAttrKeys {
		val, err := json.Marshal(searchAttrVals[i])
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Encode value %v error", val), err)
		}
		fields[key] = val
	}

	return fields
}

func processHeader(c *cli.Context) map[string][]byte {
	headerKeys := processMultipleKeys(c.String(FlagHeaderKey), " ")
	headerValues := processMultipleJSONValues(processJSONInputHelper(c, jsonTypeHeader))

	if len(headerKeys) != len(headerValues) {
		ErrorAndExit("Number of header keys and values are not equal.", nil)
	}

	return mapFromKeysValues(headerKeys, headerValues)
}

func processMemo(c *cli.Context) map[string][]byte {
	memoKeys := processMultipleKeys(c.String(FlagMemoKey), " ")
	memoValues := processMultipleJSONValues(processJSONInputHelper(c, jsonTypeMemo))

	if len(memoKeys) != len(memoValues) {
		ErrorAndExit("Number of memo keys and values are not equal.", nil)
	}

	return mapFromKeysValues(memoKeys, memoValues)
}

// helper function to print workflow progress with time refresh every second
func printWorkflowProgress(c *cli.Context, domain, wid, rid string) {
	fmt.Println(colorMagenta("Progress:"))

	wfClient := getWorkflowClient(c)
	timeElapse := 1
	isTimeElapseExist := false
	doneChan := make(chan bool)
	var lastEvent *types.HistoryEvent // used for print result of this run
	ticker := time.NewTicker(time.Second).C

	tcCtx, cancel := newIndefiniteContext(c)
	defer cancel()

	showDetails := c.Bool(FlagShowDetail)
	var maxFieldLength int
	if c.IsSet(FlagMaxFieldLength) {
		maxFieldLength = c.Int(FlagMaxFieldLength)
	}

	go func() {
		iterator, err := GetWorkflowHistoryIterator(tcCtx, wfClient, domain, wid, rid, true, types.HistoryEventFilterTypeAllEvent.Ptr())
		if err != nil {
			ErrorAndExit("Unable to get history events.", err)
		}
		for iterator.HasNext() {
			entity, err := iterator.Next()
			event := entity.(*types.HistoryEvent)
			if err != nil {
				ErrorAndExit("Unable to read event.", err)
			}
			if isTimeElapseExist {
				removePrevious2LinesFromTerminal()
				isTimeElapseExist = false
			}
			if showDetails {
				fmt.Printf("  %d, %s, %s, %s\n", event.ID, convertTime(event.GetTimestamp(), false), ColorEvent(event), HistoryEventToString(event, true, maxFieldLength))
			} else {
				fmt.Printf("  %d, %s, %s\n", event.ID, convertTime(event.GetTimestamp(), false), ColorEvent(event))
			}
			lastEvent = event
		}
		doneChan <- true
	}()

	for {
		select {
		case <-ticker:
			if isTimeElapseExist {
				removePrevious2LinesFromTerminal()
			}
			fmt.Printf("\nTime elapse: %ds\n", timeElapse)
			isTimeElapseExist = true
			timeElapse++
		case <-doneChan: // print result of this run
			fmt.Println(colorMagenta("\nResult:"))
			fmt.Printf("  Run Time: %d seconds\n", timeElapse)
			printRunStatus(lastEvent)
			return
		}
	}
}

// TerminateWorkflow terminates a workflow execution
func TerminateWorkflow(c *cli.Context) {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	reason := c.String(FlagReason)

	ctx, cancel := newContext(c)
	defer cancel()
	err := wfClient.TerminateWorkflowExecution(
		ctx,
		&types.TerminateWorkflowExecutionRequest{
			Domain: domain,
			Reason: reason,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			}, Identity: getCliIdentity(),
		},
	)

	if err != nil {
		ErrorAndExit("Terminate workflow failed.", err)
	} else {
		fmt.Println("Terminate workflow succeeded.")
	}
}

// CancelWorkflow cancels a workflow execution
func CancelWorkflow(c *cli.Context) {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	reason := c.String(FlagReason)

	ctx, cancel := newContext(c)
	defer cancel()

	err := wfClient.RequestCancelWorkflowExecution(
		ctx,
		&types.RequestCancelWorkflowExecutionRequest{
			Domain: domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			},
			Identity: getCliIdentity(),
			Cause:    reason,
		},
	)
	if err != nil {
		ErrorAndExit("Cancel workflow failed.", err)
	} else {
		fmt.Println("Cancel workflow succeeded.")
	}
}

// SignalWorkflow signals a workflow execution
func SignalWorkflow(c *cli.Context) {
	serviceClient := cFactory.ServerFrontendClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	name := getRequiredOption(c, FlagName)
	input := processJSONInput(c)

	tcCtx, cancel := newContext(c)
	defer cancel()
	err := serviceClient.SignalWorkflowExecution(
		tcCtx,
		&types.SignalWorkflowExecutionRequest{
			Domain: domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			},
			SignalName: name,
			Input:      []byte(input),
			Identity:   getCliIdentity(),
			RequestID:  uuid.New(),
		},
	)

	if err != nil {
		ErrorAndExit("Signal workflow failed.", err)
	} else {
		fmt.Println("Signal workflow succeeded.")
	}
}

// SignalWithStartWorkflowExecution starts a workflow execution if not already exists and signals it
func SignalWithStartWorkflowExecution(c *cli.Context) {
	serviceClient := cFactory.ServerFrontendClient(c)

	signalWithStartRequest := constructSignalWithStartWorkflowRequest(c)

	tcCtx, cancel := newContext(c)
	defer cancel()

	resp, err := serviceClient.SignalWithStartWorkflowExecution(tcCtx, signalWithStartRequest)
	if err != nil {
		ErrorAndExit("SignalWithStart workflow failed.", err)
	} else {
		fmt.Printf("SignalWithStart workflow succeeded. Workflow Id: %s, run Id: %s\n", signalWithStartRequest.GetWorkflowID(), resp.GetRunID())
	}
}

func constructSignalWithStartWorkflowRequest(c *cli.Context) *types.SignalWithStartWorkflowExecutionRequest {
	startRequest := constructStartWorkflowRequest(c)

	return &types.SignalWithStartWorkflowExecutionRequest{
		Domain:                              startRequest.Domain,
		WorkflowID:                          startRequest.WorkflowID,
		WorkflowType:                        startRequest.WorkflowType,
		TaskList:                            startRequest.TaskList,
		Input:                               startRequest.Input,
		ExecutionStartToCloseTimeoutSeconds: startRequest.ExecutionStartToCloseTimeoutSeconds,
		TaskStartToCloseTimeoutSeconds:      startRequest.TaskStartToCloseTimeoutSeconds,
		Identity:                            startRequest.Identity,
		RequestID:                           startRequest.RequestID,
		WorkflowIDReusePolicy:               startRequest.WorkflowIDReusePolicy,
		RetryPolicy:                         startRequest.RetryPolicy,
		CronSchedule:                        startRequest.CronSchedule,
		Memo:                                startRequest.Memo,
		SearchAttributes:                    startRequest.SearchAttributes,
		Header:                              startRequest.Header,
		SignalName:                          getRequiredOption(c, FlagName),
		SignalInput:                         []byte(processJSONInputSignal(c)),
		DelayStartSeconds:                   startRequest.DelayStartSeconds,
		JitterStartSeconds:                  startRequest.JitterStartSeconds,
	}
}

func processJSONInputSignal(c *cli.Context) string {
	return processJSONInputHelper(c, jsonTypeSignal)
}

// QueryWorkflow query workflow execution
func QueryWorkflow(c *cli.Context) {
	getRequiredGlobalOption(c, FlagDomain) // for pre-check and alert if not provided
	getRequiredOption(c, FlagWorkflowID)
	queryType := getRequiredOption(c, FlagQueryType)

	queryWorkflowHelper(c, queryType)
}

// QueryWorkflowUsingStackTrace query workflow execution using __stack_trace as query type
func QueryWorkflowUsingStackTrace(c *cli.Context) {
	queryWorkflowHelper(c, "__stack_trace")
}

func queryWorkflowHelper(c *cli.Context, queryType string) {
	serviceClient := cFactory.ServerFrontendClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	input := processJSONInput(c)

	tcCtx, cancel := newContext(c)
	defer cancel()
	queryRequest := &types.QueryWorkflowRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
		Query: &types.WorkflowQuery{
			QueryType: queryType,
		},
	}
	if input != "" {
		queryRequest.Query.QueryArgs = []byte(input)
	}
	if c.IsSet(FlagQueryRejectCondition) {
		var rejectCondition types.QueryRejectCondition
		switch c.String(FlagQueryRejectCondition) {
		case "not_open":
			rejectCondition = types.QueryRejectConditionNotOpen
		case "not_completed_cleanly":
			rejectCondition = types.QueryRejectConditionNotCompletedCleanly
		default:
			ErrorAndExit(fmt.Sprintf("invalid reject condition %v, valid values are \"not_open\" and \"not_completed_cleanly\"", c.String(FlagQueryRejectCondition)), nil)
		}
		queryRequest.QueryRejectCondition = &rejectCondition
	}
	if c.IsSet(FlagQueryConsistencyLevel) {
		var consistencyLevel types.QueryConsistencyLevel
		switch c.String(FlagQueryConsistencyLevel) {
		case "eventual":
			consistencyLevel = types.QueryConsistencyLevelEventual
		case "strong":
			consistencyLevel = types.QueryConsistencyLevelStrong
		default:
			ErrorAndExit(fmt.Sprintf("invalid query consistency level %v, valid values are \"eventual\" and \"strong\"", c.String(FlagQueryConsistencyLevel)), nil)
		}
		queryRequest.QueryConsistencyLevel = &consistencyLevel
	}
	queryResponse, err := serviceClient.QueryWorkflow(tcCtx, queryRequest)
	if err != nil {
		ErrorAndExit("Query workflow failed.", err)
		return
	}

	if queryResponse.QueryRejected != nil {
		fmt.Printf("Query was rejected, workflow is in state: %v\n", *queryResponse.QueryRejected.CloseStatus)
	} else {
		// assume it is json encoded
		fmt.Print(string(queryResponse.QueryResult))
	}
}

// ListWorkflow list workflow executions based on filters
func ListWorkflow(c *cli.Context) {
	displayPagedWorkflows(c, listWorkflows(c), !c.Bool(FlagMore))
}

// ListAllWorkflow list all workflow executions based on filters
func ListAllWorkflow(c *cli.Context) {
	displayAllWorkflows(c, filterExcludedWorkflows(c, listWorkflows(c)))
}

// ScanAllWorkflow list all workflow executions using Scan API.
// It should be faster than ListAllWorkflow, but result are not sorted.
func ScanAllWorkflow(c *cli.Context) {
	displayAllWorkflows(c, scanWorkflows(c))
}

func isQueryOpen(query string) bool {
	var openWFPattern = regexp.MustCompile(`CloseTime[ ]*=[ ]*missing`)
	return openWFPattern.MatchString(query)
}

// CountWorkflow count number of workflows
func CountWorkflow(c *cli.Context) {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	query := c.String(FlagListQuery)
	request := &types.CountWorkflowExecutionsRequest{
		Domain: domain,
		Query:  query,
	}

	ctx, cancel := newContextForLongPoll(c)
	defer cancel()
	response, err := wfClient.CountWorkflowExecutions(ctx, request)
	if err != nil {
		ErrorAndExit("Failed to count workflow.", err)
	}

	fmt.Println(response.GetCount())
}

// ListArchivedWorkflow lists archived workflow executions based on filters
func ListArchivedWorkflow(c *cli.Context) {
	printAll := c.Bool(FlagAll)
	if printAll {
		displayAllWorkflows(c, listArchivedWorkflows(c))
	} else {
		displayPagedWorkflows(c, listArchivedWorkflows(c), false)
	}
}

// DescribeWorkflow show information about the specified workflow execution
func DescribeWorkflow(c *cli.Context) {
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	describeWorkflowHelper(c, wid, rid)
}

// DescribeWorkflowWithID show information about the specified workflow execution
func DescribeWorkflowWithID(c *cli.Context) {
	if !c.Args().Present() {
		ErrorAndExit("Argument workflow_id is required.", nil)
	}
	wid := c.Args().First()
	rid := ""
	if c.NArg() >= 2 {
		rid = c.Args().Get(1)
	}

	describeWorkflowHelper(c, wid, rid)
}

func describeWorkflowHelper(c *cli.Context, wid, rid string) {
	frontendClient := cFactory.ServerFrontendClient(c)
	domain := getRequiredGlobalOption(c, FlagDomain)
	printRaw := c.Bool(FlagPrintRaw) // printRaw is false by default,
	// and will show datetime and decoded search attributes instead of raw timestamp and byte arrays
	printResetPointsOnly := c.Bool(FlagResetPointsOnly)

	ctx, cancel := newContext(c)
	defer cancel()

	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
	})
	if err != nil {
		ErrorAndExit("Describe workflow execution failed", err)
	}

	if printResetPointsOnly {
		printAutoResetPoints(resp)
		return
	}

	var o interface{}
	if printRaw {
		o = resp
	} else {
		o = convertDescribeWorkflowExecutionResponse(resp, frontendClient, c)
	}

	prettyPrintJSONObject(o)
}

type AutoResetPointRow struct {
	BinaryChecksum string    `header:"Binary Checksum"`
	CreateTime     time.Time `header:"Create Time"`
	RunID          string    `header:"RunID"`
	EventID        int64     `header:"EventID"`
}

func printAutoResetPoints(resp *types.DescribeWorkflowExecutionResponse) {
	fmt.Println("Auto Reset Points:")
	table := []AutoResetPointRow{}
	if resp.WorkflowExecutionInfo.AutoResetPoints == nil || len(resp.WorkflowExecutionInfo.AutoResetPoints.Points) == 0 {
		return
	}
	for _, pt := range resp.WorkflowExecutionInfo.AutoResetPoints.Points {
		table = append(table, AutoResetPointRow{
			BinaryChecksum: pt.GetBinaryChecksum(),
			CreateTime:     time.Unix(0, pt.GetCreatedTimeNano()),
			RunID:          pt.GetRunID(),
			EventID:        pt.GetFirstDecisionCompletedID(),
		})
	}
	RenderTable(os.Stdout, table, RenderOptions{Color: true, Border: true, PrintDateTime: true})
}

// describeWorkflowExecutionResponse is used to print datetime instead of print raw time
type describeWorkflowExecutionResponse struct {
	ExecutionConfiguration *types.WorkflowExecutionConfiguration
	WorkflowExecutionInfo  workflowExecutionInfo
	PendingActivities      []*pendingActivityInfo
	PendingChildren        []*types.PendingChildExecutionInfo
	PendingDecision        *pendingDecisionInfo
}

// workflowExecutionInfo has same fields as types.WorkflowExecutionInfo, but has datetime instead of raw time
type workflowExecutionInfo struct {
	Execution        *types.WorkflowExecution
	Type             *types.WorkflowType
	StartTime        *string // change from *int64
	CloseTime        *string // change from *int64
	CloseStatus      *types.WorkflowExecutionCloseStatus
	HistoryLength    int64
	ParentDomainID   *string
	ParentExecution  *types.WorkflowExecution
	Memo             *types.Memo
	SearchAttributes map[string]interface{}
	AutoResetPoints  *types.ResetPoints
	PartitionConfig  map[string]string
}

// pendingActivityInfo has same fields as types.PendingActivityInfo, but different field type for better display
type pendingActivityInfo struct {
	ActivityID             string
	ActivityType           *types.ActivityType
	State                  *types.PendingActivityState
	ScheduledTimestamp     *string `json:",omitempty"` // change from *int64
	LastStartedTimestamp   *string `json:",omitempty"` // change from *int64
	HeartbeatDetails       *string `json:",omitempty"` // change from []byte
	LastHeartbeatTimestamp *string `json:",omitempty"` // change from *int64
	Attempt                int32   `json:",omitempty"`
	MaximumAttempts        int32   `json:",omitempty"`
	ExpirationTimestamp    *string `json:",omitempty"` // change from *int64
	LastFailureReason      *string `json:",omitempty"`
	LastWorkerIdentity     string  `json:",omitempty"`
	LastFailureDetails     *string `json:",omitempty"` // change from []byte
}

type pendingDecisionInfo struct {
	State                      *types.PendingDecisionState
	OriginalScheduledTimestamp *string `json:",omitempty"` // change from *int64
	ScheduledTimestamp         *string `json:",omitempty"` // change from *int64
	StartedTimestamp           *string `json:",omitempty"` // change from *int64
	Attempt                    int64   `json:",omitempty"`
}

func convertDescribeWorkflowExecutionResponse(resp *types.DescribeWorkflowExecutionResponse,
	wfClient frontend.Client, c *cli.Context) *describeWorkflowExecutionResponse {

	info := resp.WorkflowExecutionInfo
	executionInfo := workflowExecutionInfo{
		Execution:        info.Execution,
		Type:             info.Type,
		StartTime:        common.StringPtr(convertTime(info.GetStartTime(), false)),
		CloseTime:        common.StringPtr(convertTime(info.GetCloseTime(), false)),
		CloseStatus:      info.CloseStatus,
		HistoryLength:    info.HistoryLength,
		ParentDomainID:   info.ParentDomainID,
		ParentExecution:  info.ParentExecution,
		Memo:             info.Memo,
		SearchAttributes: convertSearchAttributesToMapOfInterface(info.SearchAttributes, wfClient, c),
		AutoResetPoints:  info.AutoResetPoints,
		PartitionConfig:  info.PartitionConfig,
	}

	var pendingActs []*pendingActivityInfo
	var tmpAct *pendingActivityInfo
	for _, pa := range resp.PendingActivities {
		tmpAct = &pendingActivityInfo{
			ActivityID:             pa.ActivityID,
			ActivityType:           pa.ActivityType,
			State:                  pa.State,
			ScheduledTimestamp:     timestampPtrToStringPtr(pa.ScheduledTimestamp, false),
			LastStartedTimestamp:   timestampPtrToStringPtr(pa.LastStartedTimestamp, false),
			LastHeartbeatTimestamp: timestampPtrToStringPtr(pa.LastHeartbeatTimestamp, false),
			Attempt:                pa.Attempt,
			MaximumAttempts:        pa.MaximumAttempts,
			ExpirationTimestamp:    timestampPtrToStringPtr(pa.ExpirationTimestamp, false),
			LastFailureReason:      pa.LastFailureReason,
			LastWorkerIdentity:     pa.LastWorkerIdentity,
		}
		if pa.HeartbeatDetails != nil {
			tmpAct.HeartbeatDetails = common.StringPtr(string(pa.HeartbeatDetails))
		}
		if pa.LastFailureDetails != nil {
			tmpAct.LastFailureDetails = common.StringPtr(string(pa.LastFailureDetails))
		}
		pendingActs = append(pendingActs, tmpAct)
	}

	var pendingDecision *pendingDecisionInfo
	if resp.PendingDecision != nil {
		pendingDecision = &pendingDecisionInfo{
			State:              resp.PendingDecision.State,
			ScheduledTimestamp: timestampPtrToStringPtr(resp.PendingDecision.ScheduledTimestamp, false),
			StartedTimestamp:   timestampPtrToStringPtr(resp.PendingDecision.StartedTimestamp, false),
			Attempt:            resp.PendingDecision.Attempt,
		}
		// TODO: Idea here is only display decision task original scheduled timestamp if user are
		// using decision heartbeat. And we should be able to tell whether a decision task has heartbeat
		// or not by comparing the original scheduled timestamp and scheduled timestamp.
		// However, currently server may assign different value to original scheduled timestamp and
		// scheduled time even if there's no decision heartbeat.
		// if resp.PendingDecision.OriginalScheduledTimestamp != nil &&
		// 	resp.PendingDecision.ScheduledTimestamp != nil &&
		// 	*resp.PendingDecision.OriginalScheduledTimestamp != *resp.PendingDecision.ScheduledTimestamp {
		// 	pendingDecision.OriginalScheduledTimestamp = timestampPtrToStringPtr(resp.PendingDecision.OriginalScheduledTimestamp, false)
		// }
	}

	return &describeWorkflowExecutionResponse{
		ExecutionConfiguration: resp.ExecutionConfiguration,
		WorkflowExecutionInfo:  executionInfo,
		PendingActivities:      pendingActs,
		PendingChildren:        resp.PendingChildren,
		PendingDecision:        pendingDecision,
	}
}

func convertSearchAttributesToMapOfInterface(searchAttributes *types.SearchAttributes,
	wfClient frontend.Client, c *cli.Context) map[string]interface{} {

	if searchAttributes == nil || len(searchAttributes.GetIndexedFields()) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	ctx, cancel := newContext(c)
	defer cancel()
	validSearchAttributes, err := wfClient.GetSearchAttributes(ctx)
	if err != nil {
		ErrorAndExit("Error when get search attributes", err)
	}
	validKeys := validSearchAttributes.GetKeys()

	indexedFields := searchAttributes.GetIndexedFields()
	for k, v := range indexedFields {
		valueType := validKeys[k]
		deserializedValue, err := common.DeserializeSearchAttributeValue(v, valueType)
		if err != nil {
			ErrorAndExit("Error deserializing search attribute value", err)
		}
		result[k] = deserializedValue
	}

	return result
}

func getAllWorkflowIDsByQuery(c *cli.Context, query string) map[string]bool {
	wfClient := getWorkflowClient(c)
	pageSize := 1000
	var nextPageToken []byte
	var info []*types.WorkflowExecutionInfo
	result := map[string]bool{}
	for {
		info, nextPageToken = scanWorkflowExecutions(wfClient, pageSize, nextPageToken, query, c)
		for _, we := range info {
			wid := we.Execution.GetWorkflowID()
			result[wid] = true
		}

		if nextPageToken == nil {
			break
		}
	}
	return result
}

func printRunStatus(event *types.HistoryEvent) {
	switch event.GetEventType() {
	case types.EventTypeWorkflowExecutionCompleted:
		fmt.Printf("  Status: %s\n", colorGreen("COMPLETED"))
		fmt.Printf("  Output: %s\n", string(event.WorkflowExecutionCompletedEventAttributes.Result))
	case types.EventTypeWorkflowExecutionFailed:
		fmt.Printf("  Status: %s\n", colorRed("FAILED"))
		fmt.Printf("  Reason: %s\n", event.WorkflowExecutionFailedEventAttributes.GetReason())
		fmt.Printf("  Detail: %s\n", string(event.WorkflowExecutionFailedEventAttributes.Details))
	case types.EventTypeWorkflowExecutionTimedOut:
		fmt.Printf("  Status: %s\n", colorRed("TIMEOUT"))
		fmt.Printf("  Timeout Type: %s\n", event.WorkflowExecutionTimedOutEventAttributes.GetTimeoutType())
	case types.EventTypeWorkflowExecutionCanceled:
		fmt.Printf("  Status: %s\n", colorRed("CANCELED"))
		fmt.Printf("  Detail: %s\n", string(event.WorkflowExecutionCanceledEventAttributes.Details))
	}
}

// WorkflowRow is a presentation layer entity use to render a table of workflows
type WorkflowRow struct {
	WorkflowType     string                 `header:"Workflow Type" maxLength:"32"`
	WorkflowID       string                 `header:"Workflow ID"`
	RunID            string                 `header:"Run ID"`
	TaskList         string                 `header:"Task List"`
	IsCron           bool                   `header:"Is Cron"`
	StartTime        time.Time              `header:"Start Time"`
	ExecutionTime    time.Time              `header:"Execution Time"`
	EndTime          time.Time              `header:"End Time"`
	Memo             map[string]string      `header:"Memo"`
	SearchAttributes map[string]interface{} `header:"Search Attributes"`
}

func newWorkflowRow(workflow *types.WorkflowExecutionInfo) WorkflowRow {
	memo := map[string]string{}
	for k, v := range workflow.Memo.GetFields() {
		memo[k] = string(v)
	}

	sa := map[string]interface{}{}
	for k, v := range workflow.SearchAttributes.GetIndexedFields() {
		var decodedVal interface{}
		json.Unmarshal(v, &decodedVal)
		sa[k] = decodedVal
	}

	return WorkflowRow{
		WorkflowType:     workflow.Type.GetName(),
		WorkflowID:       workflow.Execution.GetWorkflowID(),
		RunID:            workflow.Execution.GetRunID(),
		TaskList:         workflow.TaskList,
		IsCron:           workflow.IsCron,
		StartTime:        time.Unix(0, workflow.GetStartTime()),
		ExecutionTime:    time.Unix(0, workflow.GetExecutionTime()),
		EndTime:          time.Unix(0, workflow.GetCloseTime()),
		Memo:             memo,
		SearchAttributes: sa,
	}
}

func workflowTableOptions(c *cli.Context) RenderOptions {
	isScanQueryOpen := isQueryOpen(c.String(FlagListQuery))

	return RenderOptions{
		DefaultTemplate: templateTable,
		Color:           true,
		PrintDateTime:   c.Bool(FlagPrintDateTime),
		PrintRawTime:    c.Bool(FlagPrintRawTime),
		OptionalColumns: map[string]bool{
			"End Time":          !(c.Bool(FlagOpen) || isScanQueryOpen),
			"Memo":              c.Bool(FlagPrintMemo),
			"Search Attributes": c.Bool(FlagPrintSearchAttr),
		},
	}
}

type getWorkflowPageFn func([]byte) ([]*types.WorkflowExecutionInfo, []byte)

func getAllWorkflows(getWorkflowPage getWorkflowPageFn) []*types.WorkflowExecutionInfo {
	var all, page []*types.WorkflowExecutionInfo
	var nextPageToken []byte
	for {
		page, nextPageToken = getWorkflowPage(nextPageToken)
		all = append(all, page...)
		if len(nextPageToken) == 0 {
			break
		}
	}
	return all
}

func filterExcludedWorkflows(c *cli.Context, getWorkflowPage getWorkflowPageFn) getWorkflowPageFn {
	excludeWIDs := map[string]bool{}
	if c.IsSet(FlagListQuery) && c.IsSet(FlagExcludeWorkflowIDByQuery) {
		excludeQuery := c.String(FlagExcludeWorkflowIDByQuery)
		excludeWIDs = getAllWorkflowIDsByQuery(c, excludeQuery)
		fmt.Printf("found %d workflowIDs to exclude\n", len(excludeWIDs))
	}

	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		page, nextPageToken := getWorkflowPage(nextPageToken)
		filtered := make([]*types.WorkflowExecutionInfo, 0, len(page))
		for _, workflow := range page {
			if excludeWIDs[workflow.GetExecution().GetWorkflowID()] {
				continue
			}
			filtered = append(filtered, workflow)
		}
		return filtered, nextPageToken
	}
}

func displayPagedWorkflows(c *cli.Context, getWorkflowPage getWorkflowPageFn, firstPageOnly bool) {
	var page []*types.WorkflowExecutionInfo
	var nextPageToken []byte
	for {
		page, nextPageToken = getWorkflowPage(nextPageToken)

		displayWorkflows(c, page)

		if firstPageOnly {
			break
		}
		if len(nextPageToken) == 0 {
			break
		}
		if !showNextPage() {
			break
		}
	}
}

func displayAllWorkflows(c *cli.Context, getWorkflowsPage getWorkflowPageFn) {
	displayWorkflows(c, getAllWorkflows(getWorkflowsPage))
}

func displayWorkflows(c *cli.Context, workflows []*types.WorkflowExecutionInfo) {
	printJSON := c.Bool(FlagPrintJSON)
	printDecodedRaw := c.Bool(FlagPrintFullyDetail)
	if printJSON || printDecodedRaw {
		fmt.Println("[")
		printListResults(workflows, printJSON, false)
		fmt.Println("]")
	} else {
		tableOptions := workflowTableOptions(c)
		var table []WorkflowRow
		for _, workflow := range workflows {
			table = append(table, newWorkflowRow(workflow))
		}
		Render(c, table, tableOptions)
	}
}

func listWorkflowExecutions(client frontend.Client, pageSize int, domain, query string, c *cli.Context) getWorkflowPageFn {
	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		request := &types.ListWorkflowExecutionsRequest{
			Domain:        domain,
			PageSize:      int32(pageSize),
			NextPageToken: nextPageToken,
			Query:         query,
		}

		ctx, cancel := newContextForLongPoll(c)
		defer cancel()
		response, err := client.ListWorkflowExecutions(ctx, request)
		if err != nil {
			ErrorAndExit("Failed to list workflow.", err)
		}
		return response.Executions, response.NextPageToken
	}
}

func listOpenWorkflow(client frontend.Client, pageSize int, earliestTime, latestTime int64, domain, workflowID, workflowType string, c *cli.Context) getWorkflowPageFn {
	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		request := &types.ListOpenWorkflowExecutionsRequest{
			Domain:          domain,
			MaximumPageSize: int32(pageSize),
			NextPageToken:   nextPageToken,
			StartTimeFilter: &types.StartTimeFilter{
				EarliestTime: common.Int64Ptr(earliestTime),
				LatestTime:   common.Int64Ptr(latestTime),
			},
		}
		if len(workflowID) > 0 {
			request.ExecutionFilter = &types.WorkflowExecutionFilter{WorkflowID: workflowID}
		}
		if len(workflowType) > 0 {
			request.TypeFilter = &types.WorkflowTypeFilter{Name: workflowType}
		}

		ctx, cancel := newContextForLongPoll(c)
		defer cancel()
		response, err := client.ListOpenWorkflowExecutions(ctx, request)
		if err != nil {
			ErrorAndExit("Failed to list open workflow.", err)
		}
		return response.Executions, response.NextPageToken
	}
}

func listClosedWorkflow(client frontend.Client, pageSize int, earliestTime, latestTime int64, domain, workflowID, workflowType string, workflowStatus types.WorkflowExecutionCloseStatus, c *cli.Context) getWorkflowPageFn {
	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		request := &types.ListClosedWorkflowExecutionsRequest{
			Domain:          domain,
			MaximumPageSize: int32(pageSize),
			NextPageToken:   nextPageToken,
			StartTimeFilter: &types.StartTimeFilter{
				EarliestTime: common.Int64Ptr(earliestTime),
				LatestTime:   common.Int64Ptr(latestTime),
			},
		}
		if len(workflowID) > 0 {
			request.ExecutionFilter = &types.WorkflowExecutionFilter{WorkflowID: workflowID}
		}
		if len(workflowType) > 0 {
			request.TypeFilter = &types.WorkflowTypeFilter{Name: workflowType}
		}
		if workflowStatus != workflowStatusNotSet {
			request.StatusFilter = &workflowStatus
		}

		ctx, cancel := newContextForLongPoll(c)
		defer cancel()
		response, err := client.ListClosedWorkflowExecutions(ctx, request)
		if err != nil {
			ErrorAndExit("Failed to list closed workflow.", err)
		}
		return response.Executions, response.NextPageToken
	}
}

func listWorkflows(c *cli.Context) getWorkflowPageFn {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	earliestTime := parseTime(c.String(FlagEarliestTime), 0)
	latestTime := parseTime(c.String(FlagLatestTime), time.Now().UnixNano())
	workflowID := c.String(FlagWorkflowID)
	workflowType := c.String(FlagWorkflowType)
	queryOpen := c.Bool(FlagOpen)
	pageSize := c.Int(FlagPageSize)
	if pageSize <= 0 {
		pageSize = defaultPageSizeForList
	}

	var workflowStatus types.WorkflowExecutionCloseStatus
	if c.IsSet(FlagWorkflowStatus) {
		if queryOpen {
			ErrorAndExit(optionErr, errors.New("you can only filter on status for closed workflow, not open workflow"))
		}
		workflowStatus = getWorkflowStatus(c.String(FlagWorkflowStatus))
	} else {
		workflowStatus = workflowStatusNotSet
	}

	if len(workflowID) > 0 && len(workflowType) > 0 {
		ErrorAndExit(optionErr, errors.New("you can filter on workflow_id or workflow_type, but not on both"))
	}

	ctx, cancel := newContextForLongPoll(c)
	defer cancel()
	resp, err := wfClient.CountWorkflowExecutions(
		ctx,
		&types.CountWorkflowExecutionsRequest{
			Domain: domain,
			Query:  c.String(FlagListQuery),
		},
	)
	if err != nil {
		printError("Unable to count workflows. Proceeding with fetching list of workflows...", err)
	} else {
		fmt.Printf("Fetching %v workflows...\n", resp.GetCount())
	}

	if c.IsSet(FlagListQuery) {
		listQuery := c.String(FlagListQuery)
		return listWorkflowExecutions(wfClient, pageSize, domain, listQuery, c)
	} else if queryOpen {
		return listOpenWorkflow(wfClient, pageSize, earliestTime, latestTime, domain, workflowID, workflowType, c)
	} else {
		return listClosedWorkflow(wfClient, pageSize, earliestTime, latestTime, domain, workflowID, workflowType, workflowStatus, c)
	}
}

func listArchivedWorkflows(c *cli.Context) getWorkflowPageFn {
	wfClient := getWorkflowClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	pageSize := c.Int(FlagPageSize)
	listQuery := getRequiredOption(c, FlagListQuery)
	if pageSize <= 0 {
		pageSize = defaultPageSizeForList
	}

	contextTimeout := defaultContextTimeoutForListArchivedWorkflow
	if c.GlobalIsSet(FlagContextTimeout) {
		contextTimeout = time.Duration(c.GlobalInt(FlagContextTimeout)) * time.Second
	}

	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		request := &types.ListArchivedWorkflowExecutionsRequest{
			Domain:        domain,
			PageSize:      int32(pageSize),
			Query:         listQuery,
			NextPageToken: nextPageToken,
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()

		result, err := wfClient.ListArchivedWorkflowExecutions(ctx, request)
		if err != nil {
			cancel()
			ErrorAndExit("Failed to list archived workflow.", err)
		}
		return result.Executions, result.NextPageToken
	}
}

func scanWorkflows(c *cli.Context) getWorkflowPageFn {
	wfClient := getWorkflowClient(c)
	listQuery := c.String(FlagListQuery)
	pageSize := c.Int(FlagPageSize)
	if pageSize <= 0 {
		pageSize = defaultPageSizeForScan
	}

	return func(nextPageToken []byte) ([]*types.WorkflowExecutionInfo, []byte) {
		return scanWorkflowExecutions(wfClient, pageSize, nextPageToken, listQuery, c)
	}
}

func scanWorkflowExecutions(client frontend.Client, pageSize int, nextPageToken []byte, query string, c *cli.Context) ([]*types.WorkflowExecutionInfo, []byte) {
	domain := getRequiredGlobalOption(c, FlagDomain)

	request := &types.ListWorkflowExecutionsRequest{
		Domain:        domain,
		PageSize:      int32(pageSize),
		NextPageToken: nextPageToken,
		Query:         query,
	}
	ctx, cancel := newContextForLongPoll(c)
	defer cancel()
	response, err := client.ScanWorkflowExecutions(ctx, request)
	if err != nil {
		ErrorAndExit("Failed to list workflow.", err)
	}
	return response.Executions, response.NextPageToken
}

func getWorkflowStatus(statusStr string) types.WorkflowExecutionCloseStatus {
	if status, ok := workflowClosedStatusMap[strings.ToLower(statusStr)]; ok {
		return status
	}
	ErrorAndExit(optionErr, errors.New("option status is not one of allowed values "+
		"[completed, failed, canceled, terminated, continued_as_new, timed_out]"))
	return 0
}

func getWorkflowIDReusePolicy(value int) *types.WorkflowIDReusePolicy {
	if value >= 0 && types.WorkflowIDReusePolicy(value) <= types.WorkflowIDReusePolicyTerminateIfRunning {
		return types.WorkflowIDReusePolicy(value).Ptr()
	}
	// At this point, the policy should return if the value is valid
	ErrorAndExit(fmt.Sprintf("Option %v value is not in supported range.", FlagWorkflowIDReusePolicy), nil)
	return nil
}

// default will print decoded raw
func printListResults(executions []*types.WorkflowExecutionInfo, inJSON bool, more bool) {
	for i, execution := range executions {
		if inJSON {
			j, _ := json.Marshal(execution)
			if more || i < len(executions)-1 {
				fmt.Println(string(j) + ",")
			} else {
				fmt.Println(string(j))
			}
		} else {
			if more || i < len(executions)-1 {
				fmt.Println(anyToString(execution, true, 0) + ",")
			} else {
				fmt.Println(anyToString(execution, true, 0))
			}
		}
	}
}

// ObserveHistory show the process of running workflow
func ObserveHistory(c *cli.Context) {
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	domain := getRequiredGlobalOption(c, FlagDomain)

	printWorkflowProgress(c, domain, wid, rid)
}

// ResetWorkflow reset workflow
func ResetWorkflow(c *cli.Context) {
	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	reason := getRequiredOption(c, FlagReason)
	if len(reason) == 0 {
		ErrorAndExit("wrong reason", fmt.Errorf("reason cannot be empty"))
	}
	eventID := c.Int64(FlagEventID)
	resetType := c.String(FlagResetType)
	decisionOffset := c.Int(FlagDecisionOffset)
	if decisionOffset > 0 {
		ErrorAndExit("Only decision offset <=0 is supported", nil)
	}

	extraForResetType, ok := resetTypesMap[resetType]
	if !ok && eventID <= 0 {
		ErrorAndExit("Must specify valid eventID or valid resetType", nil)
	}
	if ok && len(extraForResetType) > 0 {
		getRequiredOption(c, extraForResetType)
	}

	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.ServerFrontendClient(c)
	rid := c.String(FlagRunID)
	var err error
	if rid == "" {
		rid, err = getCurrentRunID(ctx, domain, wid, frontendClient)
		if err != nil {
			ErrorAndExit("Cannot get latest RunID as default", err)
		}
	}

	resetBaseRunID := rid
	decisionFinishID := eventID
	if resetType != "" {
		resetBaseRunID, decisionFinishID, err = getResetEventIDByType(ctx, c, resetType, decisionOffset, domain, wid, rid, frontendClient)
		if err != nil {
			ErrorAndExit("getResetEventIDByType failed", err)
		}
	}
	resp, err := frontendClient.ResetWorkflowExecution(ctx, &types.ResetWorkflowExecutionRequest{
		Domain: domain,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      resetBaseRunID,
		},
		Reason:                fmt.Sprintf("%v:%v", getCurrentUserFromEnv(), reason),
		DecisionFinishEventID: decisionFinishID,
		RequestID:             uuid.New(),
		SkipSignalReapply:     c.Bool(FlagSkipSignalReapply),
	})
	if err != nil {
		ErrorAndExit("reset failed", err)
	}
	prettyPrintJSONObject(resp)
}

func processResets(c *cli.Context, domain string, wes chan types.WorkflowExecution, done chan bool, wg *sync.WaitGroup, params batchResetParamsType) {
	for {
		select {
		case we := <-wes:
			fmt.Println("received: ", we.GetWorkflowID(), we.GetRunID())
			wid := we.GetWorkflowID()
			rid := we.GetRunID()
			var err error
			for i := 0; i < 3; i++ {
				err = doReset(c, domain, wid, rid, params)
				if err == nil {
					break
				}
				if _, ok := err.(*types.BadRequestError); ok {
					break
				}
				fmt.Println("failed and retry...: ", wid, rid, err)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			if err != nil {
				fmt.Println("[ERROR] failed processing: ", wid, rid, err.Error())
			}
		case <-done:
			wg.Done()
			return
		}
	}
}

type batchResetParamsType struct {
	reason               string
	skipCurrentOpen      bool
	skipCurrentCompleted bool
	nonDeterministicOnly bool
	skipBaseNotCurrent   bool
	dryRun               bool
	resetType            string
	decisionOffset       int
	skipSignalReapply    bool
}

// ResetInBatch resets workflow in batch
func ResetInBatch(c *cli.Context) {
	domain := getRequiredGlobalOption(c, FlagDomain)
	resetType := getRequiredOption(c, FlagResetType)
	decisionOffset := c.Int(FlagDecisionOffset)
	if decisionOffset > 0 {
		ErrorAndExit("Only decision offset <=0 is supported", nil)
	}

	inFileName := c.String(FlagInputFile)
	query := c.String(FlagListQuery)
	excludeFileName := c.String(FlagExcludeFile)
	excludeQuery := c.String(FlagExcludeWorkflowIDByQuery)
	separator := c.String(FlagInputSeparator)
	parallel := c.Int(FlagParallism)

	extraForResetType, ok := resetTypesMap[resetType]
	if !ok {
		ErrorAndExit("Not supported reset type", nil)
	} else if len(extraForResetType) > 0 {
		getRequiredOption(c, extraForResetType)
	}

	if excludeFileName != "" && excludeQuery != "" {
		ErrorAndExit("Only one of the excluding option is allowed", nil)
	}

	batchResetParams := batchResetParamsType{
		reason:               getRequiredOption(c, FlagReason),
		skipCurrentOpen:      c.Bool(FlagSkipCurrentOpen),
		skipCurrentCompleted: c.Bool(FlagSkipCurrentCompleted),
		nonDeterministicOnly: c.Bool(FlagNonDeterministicOnly),
		skipBaseNotCurrent:   c.Bool(FlagSkipBaseIsNotCurrent),
		dryRun:               c.Bool(FlagDryRun),
		resetType:            resetType,
		decisionOffset:       decisionOffset,
		skipSignalReapply:    c.Bool(FlagSkipSignalReapply),
	}

	if inFileName == "" && query == "" {
		ErrorAndExit("Must provide input file or list query to get target workflows to reset", nil)
	}

	wg := &sync.WaitGroup{}

	wes := make(chan types.WorkflowExecution)
	done := make(chan bool)
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go processResets(c, domain, wes, done, wg, batchResetParams)
	}

	// read excluded workflowIDs
	excludeWIDs := map[string]bool{}
	if excludeFileName != "" {
		excludeWIDs = loadWorkflowIDsFromFile(excludeFileName, separator)
	}
	if excludeQuery != "" {
		excludeWIDs = getAllWorkflowIDsByQuery(c, excludeQuery)
	}

	fmt.Println("num of excluded WorkflowIDs:", len(excludeWIDs))

	if len(inFileName) > 0 {
		inFile, err := os.Open(inFileName)
		if err != nil {
			ErrorAndExit("Open failed", err)
		}
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		idx := 0
		for scanner.Scan() {
			idx++
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 {
				fmt.Printf("line %v is empty, skipped\n", idx)
				continue
			}
			cols := strings.Split(line, separator)
			if len(cols) < 1 {
				ErrorAndExit("Split failed", fmt.Errorf("line %v has less than 1 cols separated by comma, only %v ", idx, len(cols)))
			}
			fmt.Printf("Start processing line %v ...\n", idx)
			wid := strings.TrimSpace(cols[0])
			rid := ""
			if len(cols) > 1 {
				rid = strings.TrimSpace(cols[1])
			}

			if excludeWIDs[wid] {
				fmt.Println("skip by exclude file: ", wid, rid)
				continue
			}

			wes <- types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			}
		}
	} else {
		wfClient := getWorkflowClient(c)
		pageSize := 1000
		var nextPageToken []byte
		var result []*types.WorkflowExecutionInfo
		for {
			result, nextPageToken = scanWorkflowExecutions(wfClient, pageSize, nextPageToken, query, c)
			for _, we := range result {
				wid := we.Execution.GetWorkflowID()
				rid := we.Execution.GetRunID()
				if excludeWIDs[wid] {
					fmt.Println("skip by exclude file: ", wid, rid)
					continue
				}

				wes <- types.WorkflowExecution{
					WorkflowID: wid,
					RunID:      rid,
				}
			}

			if nextPageToken == nil {
				break
			}
		}
	}

	close(done)
	fmt.Println("wait for all goroutines...")
	wg.Wait()
}

func loadWorkflowIDsFromFile(excludeFileName, separator string) map[string]bool {
	excludeWIDs := map[string]bool{}
	if len(excludeFileName) > 0 {
		// This code is only used in the CLI. The input provided is from a trusted user.
		// #nosec
		excFile, err := os.Open(excludeFileName)
		if err != nil {
			ErrorAndExit("Open failed2", err)
		}
		defer excFile.Close()
		scanner := bufio.NewScanner(excFile)
		idx := 0
		for scanner.Scan() {
			idx++
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 {
				fmt.Printf("line %v is empty, skipped\n", idx)
				continue
			}
			cols := strings.Split(line, separator)
			if len(cols) < 1 {
				ErrorAndExit("Split failed", fmt.Errorf("line %v has less than 1 cols separated by comma, only %v ", idx, len(cols)))
			}
			wid := strings.TrimSpace(cols[0])
			excludeWIDs[wid] = true
		}
	}
	return excludeWIDs
}

func printErrorAndReturn(msg string, err error) error {
	fmt.Println(msg)
	return err
}

func doReset(c *cli.Context, domain, wid, rid string, params batchResetParamsType) error {
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.ServerFrontendClient(c)
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
		},
	})
	if err != nil {
		return printErrorAndReturn("DescribeWorkflowExecution failed", err)
	}

	currentRunID := resp.WorkflowExecutionInfo.Execution.GetRunID()
	if currentRunID != rid && params.skipBaseNotCurrent {
		fmt.Println("skip because base run is different from current run: ", wid, rid, currentRunID)
		return nil
	}
	if rid == "" {
		rid = currentRunID
	}

	if resp.WorkflowExecutionInfo.CloseStatus == nil || resp.WorkflowExecutionInfo.CloseTime == nil {
		if params.skipCurrentOpen {
			fmt.Println("skip because current run is open: ", wid, rid, currentRunID)
			return nil
		}
	}

	if resp.WorkflowExecutionInfo.GetCloseStatus() == types.WorkflowExecutionCloseStatusCompleted {
		if params.skipCurrentCompleted {
			fmt.Println("skip because current run is completed: ", wid, rid, currentRunID)
			return nil
		}
	}

	if params.nonDeterministicOnly {
		isLDN, err := isLastEventDecisionTaskFailedWithNonDeterminism(ctx, domain, wid, rid, frontendClient)
		if err != nil {
			return printErrorAndReturn("check isLastEventDecisionTaskFailedWithNonDeterminism failed", err)
		}
		if !isLDN {
			fmt.Println("skip because last event is not DecisionTaskFailedWithNonDeterminism")
			return nil
		}
	}

	resetBaseRunID, decisionFinishID, err := getResetEventIDByType(ctx, c, params.resetType, params.decisionOffset, domain, wid, rid, frontendClient)
	if err != nil {
		return printErrorAndReturn("getResetEventIDByType failed", err)
	}
	fmt.Println("DecisionFinishEventId for reset:", wid, rid, resetBaseRunID, decisionFinishID)

	if params.dryRun {
		fmt.Printf("dry run to reset wid: %v, rid:%v to baseRunID:%v, eventID:%v \n", wid, rid, resetBaseRunID, decisionFinishID)
	} else {
		resp2, err := frontendClient.ResetWorkflowExecution(ctx, &types.ResetWorkflowExecutionRequest{
			Domain: domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      resetBaseRunID,
			},
			DecisionFinishEventID: decisionFinishID,
			RequestID:             uuid.New(),
			Reason:                fmt.Sprintf("%v:%v", getCurrentUserFromEnv(), params.reason),
			SkipSignalReapply:     params.skipSignalReapply,
		})

		if err != nil {
			return printErrorAndReturn("ResetWorkflowExecution failed", err)
		}
		fmt.Println("new runID for wid/rid is ,", wid, rid, resp2.GetRunID())
	}

	return nil
}

func isLastEventDecisionTaskFailedWithNonDeterminism(ctx context.Context, domain, wid, rid string, frontendClient frontend.Client) (bool, error) {
	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

	var firstEvent, decisionFailed *types.HistoryEvent
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return false, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if firstEvent == nil {
				firstEvent = e
			}
			if e.GetEventType() == types.EventTypeDecisionTaskFailed {
				decisionFailed = e
			} else if e.GetEventType() == types.EventTypeDecisionTaskCompleted {
				decisionFailed = nil
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}

	if decisionFailed != nil {
		attr := decisionFailed.GetDecisionTaskFailedEventAttributes()
		if attr.GetCause() == types.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure ||
			strings.Contains(string(attr.GetDetails()), "nondeterministic") {
			fmt.Printf("found non-deterministic workflow wid:%v, rid:%v, originalStartTime:%v \n", wid, rid, time.Unix(0, firstEvent.GetTimestamp()))
			return true, nil
		}
	}

	return false, nil
}

func getResetEventIDByType(
	ctx context.Context,
	c *cli.Context,
	resetType string, decisionOffset int,
	domain, wid, rid string,
	frontendClient frontend.Client,
) (resetBaseRunID string, decisionFinishID int64, err error) {
	// default to the same runID
	resetBaseRunID = rid

	fmt.Println("resetType:", resetType)
	switch resetType {
	case resetTypeLastDecisionCompleted:
		decisionFinishID, err = getLastDecisionTaskByType(ctx, domain, wid, rid, frontendClient, types.EventTypeDecisionTaskCompleted, decisionOffset)
		if err != nil {
			return
		}
	case resetTypeLastContinuedAsNew:
		// this reset type may change the base runID
		resetBaseRunID, decisionFinishID, err = getLastContinueAsNewID(ctx, domain, wid, rid, frontendClient)
		if err != nil {
			return
		}
	case resetTypeFirstDecisionCompleted:
		decisionFinishID, err = getFirstDecisionTaskByType(ctx, domain, wid, rid, frontendClient, types.EventTypeDecisionTaskCompleted)
		if err != nil {
			return
		}
	case resetTypeBadBinary:
		binCheckSum := c.String(FlagResetBadBinaryChecksum)
		decisionFinishID, err = getBadDecisionCompletedID(ctx, domain, wid, rid, binCheckSum, frontendClient)
		if err != nil {
			return
		}
	case resetTypeDecisionCompletedTime:
		earliestTime := parseTime(c.String(FlagEarliestTime), 0)
		decisionFinishID, err = getEarliestDecisionID(ctx, domain, wid, rid, earliestTime, frontendClient)
		if err != nil {
			return
		}
	case resetTypeFirstDecisionScheduled:
		decisionFinishID, err = getFirstDecisionTaskByType(ctx, domain, wid, rid, frontendClient, types.EventTypeDecisionTaskScheduled)
		if err != nil {
			return
		}
		// decisionFinishID is exclusive in reset API
		decisionFinishID++
	case resetTypeLastDecisionScheduled:
		decisionFinishID, err = getLastDecisionTaskByType(ctx, domain, wid, rid, frontendClient, types.EventTypeDecisionTaskScheduled, decisionOffset)
		if err != nil {
			return
		}
		// decisionFinishID is exclusive in reset API
		decisionFinishID++
	default:
		panic("not supported resetType")
	}
	return
}

func getFirstDecisionTaskByType(
	ctx context.Context,
	domain string,
	workflowID string,
	runID string,
	frontendClient frontend.Client,
	decisionType types.EventType,
) (decisionFinishID int64, err error) {

	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}

		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == decisionType {
				decisionFinishID = e.ID
				return decisionFinishID, nil
			}
		}

		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if decisionFinishID == 0 {
		return 0, printErrorAndReturn("Get DecisionFinishID failed", fmt.Errorf("no DecisionFinishID"))
	}
	return
}

func getCurrentRunID(ctx context.Context, domain, wid string, frontendClient frontend.Client) (string, error) {
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.WorkflowExecutionInfo.Execution.GetRunID(), nil
}

func getBadDecisionCompletedID(ctx context.Context, domain, wid, rid, binChecksum string, frontendClient frontend.Client) (decisionFinishID int64, err error) {
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
	})
	if err != nil {
		return 0, printErrorAndReturn("DescribeWorkflowExecution failed", err)
	}

	_, p := execution.FindAutoResetPoint(clock.NewRealTimeSource(), &types.BadBinaries{
		Binaries: map[string]*types.BadBinaryInfo{
			binChecksum: {},
		},
	}, resp.WorkflowExecutionInfo.AutoResetPoints)
	if p != nil {
		decisionFinishID = p.GetFirstDecisionCompletedID()
	}

	if decisionFinishID == 0 {
		return 0, printErrorAndReturn("Get DecisionFinishID failed", &types.BadRequestError{Message: "no DecisionFinishID"})
	}
	return
}

func getLastDecisionTaskByType(
	ctx context.Context,
	domain string,
	workflowID string,
	runID string,
	frontendClient frontend.Client,
	decisionType types.EventType,
	decisionOffset int,
) (int64, error) {

	// this fixedSizeQueue is for remembering the offset decision eventID
	fixedSizeQueue := make([]int64, 0)
	size := int(math.Abs(float64(decisionOffset))) + 1

	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}

		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == decisionType {
				decisionEventID := e.ID
				fixedSizeQueue = append(fixedSizeQueue, decisionEventID)
				if len(fixedSizeQueue) > size {
					fixedSizeQueue = fixedSizeQueue[1:]
				}
			}
		}

		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if len(fixedSizeQueue) == 0 {
		return 0, printErrorAndReturn("Get DecisionFinishID failed", fmt.Errorf("no DecisionFinishID"))
	}
	return fixedSizeQueue[0], nil
}

func getLastContinueAsNewID(ctx context.Context, domain, wid, rid string, frontendClient frontend.Client) (resetBaseRunID string, decisionFinishID int64, err error) {
	// get first event
	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
		MaximumPageSize: 1,
		NextPageToken:   nil,
	}
	resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
	if err != nil {
		return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
	}
	firstEvent := resp.History.Events[0]
	resetBaseRunID = firstEvent.GetWorkflowExecutionStartedEventAttributes().GetContinuedExecutionRunID()
	if resetBaseRunID == "" {
		return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", fmt.Errorf("cannot get resetBaseRunID"))
	}

	req = &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      resetBaseRunID,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == types.EventTypeDecisionTaskCompleted {
				decisionFinishID = e.ID
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if decisionFinishID == 0 {
		return "", 0, printErrorAndReturn("Get DecisionFinishID failed", fmt.Errorf("no DecisionFinishID"))
	}
	return
}

// CompleteActivity completes an activity
func CompleteActivity(c *cli.Context) {
	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := getRequiredOption(c, FlagRunID)
	activityID := getRequiredOption(c, FlagActivityID)
	if len(activityID) == 0 {
		ErrorAndExit("Invalid activityID", fmt.Errorf("activityID cannot be empty"))
	}
	result := getRequiredOption(c, FlagResult)
	identity := getRequiredOption(c, FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.ServerFrontendClient(c)
	err := frontendClient.RespondActivityTaskCompletedByID(ctx, &types.RespondActivityTaskCompletedByIDRequest{
		Domain:     domain,
		WorkflowID: wid,
		RunID:      rid,
		ActivityID: activityID,
		Result:     []byte(result),
		Identity:   identity,
	})
	if err != nil {
		ErrorAndExit("Completing activity failed", err)
	} else {
		fmt.Println("Complete activity successfully.")
	}
}

// FailActivity fails an activity
func FailActivity(c *cli.Context) {
	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := getRequiredOption(c, FlagRunID)
	activityID := getRequiredOption(c, FlagActivityID)
	if len(activityID) == 0 {
		ErrorAndExit("Invalid activityID", fmt.Errorf("activityID cannot be empty"))
	}
	reason := getRequiredOption(c, FlagReason)
	detail := getRequiredOption(c, FlagDetail)
	identity := getRequiredOption(c, FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.ServerFrontendClient(c)
	err := frontendClient.RespondActivityTaskFailedByID(ctx, &types.RespondActivityTaskFailedByIDRequest{
		Domain:     domain,
		WorkflowID: wid,
		RunID:      rid,
		ActivityID: activityID,
		Reason:     common.StringPtr(reason),
		Details:    []byte(detail),
		Identity:   identity,
	})
	if err != nil {
		ErrorAndExit("Failing activity failed", err)
	} else {
		fmt.Println("Fail activity successfully.")
	}
}

// ObserveHistoryWithID show the process of running workflow
func ObserveHistoryWithID(c *cli.Context) {
	domain := getRequiredGlobalOption(c, FlagDomain)
	if !c.Args().Present() {
		ErrorAndExit("Argument workflow_id is required.", nil)
	}
	wid := c.Args().First()
	rid := ""
	if c.NArg() >= 2 {
		rid = c.Args().Get(1)
	}

	printWorkflowProgress(c, domain, wid, rid)
}

func getEarliestDecisionID(
	ctx context.Context,
	domain string, wid string,
	rid string, earliestTime int64,
	frontendClient frontend.Client,
) (decisionFinishID int64, err error) {
	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

OuterLoop:
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == types.EventTypeDecisionTaskCompleted {
				if e.GetTimestamp() >= earliestTime {
					decisionFinishID = e.ID
					break OuterLoop
				}
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if decisionFinishID == 0 {
		return 0, printErrorAndReturn("Get DecisionFinishID failed", fmt.Errorf("no DecisionFinishID"))
	}
	return
}
