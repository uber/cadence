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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

// AdminDescribeTaskList displays poller and status information of task list.
func AdminDescribeTaskList(c *cli.Context) {
	frontendClient := cFactory.ServerFrontendClient(c)
	domain := getRequiredGlobalOption(c, FlagDomain)
	taskList := getRequiredOption(c, FlagTaskList)
	taskListType := types.TaskListTypeDecision
	if strings.ToLower(c.String(FlagTaskListType)) == "activity" {
		taskListType = types.TaskListTypeActivity
	}

	ctx, cancel := newContext(c)
	defer cancel()
	request := &types.DescribeTaskListRequest{
		Domain:                domain,
		TaskList:              &types.TaskList{Name: taskList},
		TaskListType:          &taskListType,
		IncludeTaskListStatus: true,
	}

	response, err := frontendClient.DescribeTaskList(ctx, request)
	if err != nil {
		ErrorAndExit("Operation DescribeTaskList failed.", err)
	}

	taskListStatus := response.GetTaskListStatus()
	if taskListStatus == nil {
		ErrorAndExit(colorMagenta("No tasklist status information."), nil)
	}
	printTaskListStatus(taskListStatus)
	fmt.Printf("\n")

	pollers := response.Pollers
	if len(pollers) == 0 {
		ErrorAndExit(colorMagenta("No poller for tasklist: "+taskList), nil)
	}
	printPollerInfo(pollers, taskListType)
}

// AdminListTaskList displays all task lists under a domain.
func AdminListTaskList(c *cli.Context) {
	frontendClient := cFactory.ServerFrontendClient(c)
	domain := getRequiredGlobalOption(c, FlagDomain)

	ctx, cancel := newContext(c)
	defer cancel()
	request := &types.GetTaskListsByDomainRequest{
		Domain: domain,
	}

	response, err := frontendClient.GetTaskListsByDomain(ctx, request)
	if err != nil {
		ErrorAndExit("Operation GetTaskListByDomain failed.", err)
	}

	fmt.Println("Task Lists for domain " + domain + ":")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(true)
	table.SetColumnSeparator("|")
	table.SetHeader([]string{"Task List Name", "Type", "Poller Count"})
	table.SetHeaderLine(true)
	table.SetHeaderColor(tableHeaderBlue, tableHeaderBlue, tableHeaderBlue)
	for name, taskList := range response.GetDecisionTaskListMap() {
		table.Append([]string{name, strconv.Itoa(len(taskList.GetPollers()))})
	}
	for name, taskList := range response.GetActivityTaskListMap() {
		table.Append([]string{name, strconv.Itoa(len(taskList.GetPollers()))})
	}
	table.Render()
}

func printTaskListStatus(taskListStatus *types.TaskListStatus) {
	taskIDBlock := taskListStatus.GetTaskIDBlock()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("|")
	table.SetHeader([]string{"Read Level", "Ack Level", "Backlog", "Lease Start TaskID", "Lease End TaskID"})
	table.SetHeaderLine(false)
	table.SetHeaderColor(tableHeaderBlue, tableHeaderBlue, tableHeaderBlue, tableHeaderBlue, tableHeaderBlue)
	table.Append([]string{strconv.FormatInt(taskListStatus.GetReadLevel(), 10),
		strconv.FormatInt(taskListStatus.GetAckLevel(), 10),
		strconv.FormatInt(taskListStatus.GetBacklogCountHint(), 10),
		strconv.FormatInt(taskIDBlock.GetStartID(), 10),
		strconv.FormatInt(taskIDBlock.GetEndID(), 10)})
	table.Render()
}

func printPollerInfo(pollers []*types.PollerInfo, taskListType types.TaskListType) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("|")
	if taskListType == types.TaskListTypeActivity {
		table.SetHeader([]string{"Activity Poller Identity", "Last Access Time"})
	} else {
		table.SetHeader([]string{"Decision Poller Identity", "Last Access Time"})
	}
	table.SetHeaderLine(false)
	table.SetHeaderColor(tableHeaderBlue, tableHeaderBlue)
	for _, poller := range pollers {
		table.Append([]string{poller.GetIdentity(), convertTime(poller.GetLastAccessTime(), false)})
	}
	table.Render()
}
