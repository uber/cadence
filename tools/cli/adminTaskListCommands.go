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
	"strings"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/types"
)

type (
	TaskListRow struct {
		Name        string `header:"Task List Name"`
		Type        string `header:"Type"`
		PollerCount int    `header:"Poller Count"`
	}
	TaskListStatusRow struct {
		ReadLevel int64 `header:"Read Level"`
		AckLevel  int64 `header:"Ack Level"`
		Backlog   int64 `header:"Backlog"`
		StartID   int64 `header:"Lease Start TaskID"`
		EndID     int64 `header:"Lease End TaskID"`
	}
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
	printTaskListPollers(pollers, taskListType)
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
	table := []TaskListRow{}
	for name, taskList := range response.GetDecisionTaskListMap() {
		table = append(table, TaskListRow{name, "Decision", len(taskList.GetPollers())})
	}
	for name, taskList := range response.GetActivityTaskListMap() {
		table = append(table, TaskListRow{name, "Activity", len(taskList.GetPollers())})
	}
	RenderTable(os.Stdout, table, RenderOptions{Color: true, Border: true})
}

func printTaskListStatus(taskListStatus *types.TaskListStatus) {
	table := []TaskListStatusRow{{
		ReadLevel: taskListStatus.GetReadLevel(),
		AckLevel:  taskListStatus.GetAckLevel(),
		Backlog:   taskListStatus.GetBacklogCountHint(),
		StartID:   taskListStatus.GetTaskIDBlock().GetStartID(),
		EndID:     taskListStatus.GetTaskIDBlock().GetEndID(),
	}}
	RenderTable(os.Stdout, table, RenderOptions{Color: true})
}
