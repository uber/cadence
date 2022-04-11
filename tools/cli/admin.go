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
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/service/worker/scanner/executions"
)

func newAdminWorkflowCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "show",
			Aliases: []string{"show"},
			Usage:   "show workflow history from database",
			Flags: append(getDBFlags(),
				// v2 history events
				cli.StringFlag{
					Name:  FlagTreeID,
					Usage: "TreeID",
				},
				cli.StringFlag{
					Name:  FlagBranchID,
					Usage: "BranchID",
				},
				cli.StringFlag{
					Name:  FlagOutputFilenameWithAlias,
					Usage: "output file",
				},
				// support mysql query
				cli.IntFlag{
					Name:  FlagShardIDWithAlias,
					Usage: "ShardID",
				}),
			Action: func(c *cli.Context) {
				AdminShowWorkflow(c)
			},
		},
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "Describe internal information of workflow execution",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID",
				},
			},
			Action: func(c *cli.Context) {
				AdminDescribeWorkflow(c)
			},
		},
		{
			Name:    "refresh-tasks",
			Aliases: []string{"rt"},
			Usage:   "Refreshes all the tasks of a workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID",
				},
			},
			Action: func(c *cli.Context) {
				AdminRefreshWorkflowTasks(c)
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"del"},
			Usage:   "Delete current workflow execution and the mutableState record",
			Flags: append(getDBFlags(),
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID",
				},
				cli.BoolFlag{
					Name:  FlagSkipErrorModeWithAlias,
					Usage: "skip errors when deleting history",
				},
				cli.BoolFlag{
					Name:  FlagRemote,
					Usage: "Executes deletion on server side",
				}),
			Action: func(c *cli.Context) {
				AdminDeleteWorkflow(c)
			},
		},
		{
			Name:    "fix_corruption",
			Aliases: []string{"fc"},
			Usage:   "Checks if workflow record is corrupted in database and cleans up",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID",
				},
				cli.BoolFlag{
					Name:  FlagSkipErrorModeWithAlias,
					Usage: "Skip errors and tries to delete as much as possible from the DB",
				},
			},
			Action: func(c *cli.Context) {
				AdminMaintainCorruptWorkflow(c)
			},
		},
	}
}

func newAdminShardManagementCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "describe",
			Aliases: []string{"d"},
			Usage:   "Describe shard by Id",
			Flags: append(
				getDBFlags(),
				cli.IntFlag{
					Name:  FlagShardID,
					Usage: "The Id of the shard to describe",
				},
			),
			Action: func(c *cli.Context) {
				AdminDescribeShard(c)
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List shard distribution",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  FlagPageSize,
					Value: 100,
					Usage: "Max number of results to return",
				},
				cli.IntFlag{
					Name:  FlagPageID,
					Value: 0,
					Usage: "Option to show results offset from pagesize * page_id",
				},
				getFormatFlag(),
			},
			Action: func(c *cli.Context) {
				AdminDescribeShardDistribution(c)
			},
		},
		{
			Name:    "setRangeID",
			Aliases: []string{"srid"},
			Usage:   "Force update shard rangeID",
			Flags: append(
				getDBFlags(),
				cli.IntFlag{
					Name:  FlagShardIDWithAlias,
					Usage: "ID of the shard to reset",
				},
				cli.Int64Flag{
					Name:  FlagRangeIDWithAlias,
					Usage: "new shard rangeID",
				},
			),
			Action: func(c *cli.Context) {
				AdminSetShardRangeID(c)
			},
		},
		{
			Name:    "closeShard",
			Aliases: []string{"clsh"},
			Usage:   "close a shard given a shard id",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  FlagShardID,
					Usage: "ShardID for the cadence cluster to manage",
				},
			},
			Action: func(c *cli.Context) {
				AdminCloseShard(c)
			},
		},
		{
			Name:    "removeTask",
			Aliases: []string{"rmtk"},
			Usage:   "remove a task based on shardID, task type, taskID, and task visibility timestamp",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  FlagShardID,
					Usage: "shardID",
				},
				cli.Int64Flag{
					Name:  FlagTaskID,
					Usage: "taskID",
				},
				cli.IntFlag{
					Name:  FlagTaskType,
					Usage: "task type: 2 (transfer task), 3 (timer task), 4 (replication task) or 6 (cross-cluster task)",
				},
				cli.Int64Flag{
					Name:  FlagTaskVisibilityTimestamp,
					Usage: "task visibility timestamp in nano (required for removing timer task)",
				},
				cli.StringFlag{
					Name:  FlagCluster,
					Usage: "target cluster of the task (required for removing cross-cluster task)",
				},
			},
			Action: func(c *cli.Context) {
				AdminRemoveTask(c)
			},
		},
		{
			Name:  "timers",
			Usage: "get scheduled timers for a given time range",
			Flags: append(getDBFlags(),
				cli.IntFlag{
					Name:  FlagShardID,
					Usage: "shardID",
				},
				cli.IntFlag{
					Name:  FlagPageSize,
					Usage: "page size used to query db executions table",
					Value: 500,
				},
				cli.StringFlag{
					Name:  FlagStartDate,
					Usage: "start date",
					Value: time.Now().UTC().Format(time.RFC3339),
				},
				cli.StringFlag{
					Name:  FlagEndDate,
					Usage: "end date",
					Value: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
				},
				cli.StringFlag{
					Name:  FlagDomainID,
					Usage: "filter tasks by DomainID",
				},
				cli.IntSliceFlag{
					Name: FlagTimerType,
					Usage: "timer types: 0 - DecisionTimeoutTask, 1 - TaskTypeActivityTimeout, " +
						"2 - TaskTypeUserTimer, 3 - TaskTypeWorkflowTimeout, 4 - TaskTypeDeleteHistoryEvent, " +
						"5 - TaskTypeActivityRetryTimer, 6 - TaskTypeWorkflowBackoffTimer",
					Value: &cli.IntSlice{-1},
				},
				cli.BoolFlag{
					Name:  FlagPrintJSON,
					Usage: "print raw json data instead of histogram",
				},

				cli.BoolFlag{
					Name:  FlagSkipErrorMode,
					Usage: "skip errors",
				},
				cli.StringFlag{
					Name:  FlagInputFile,
					Usage: "file to use, will not connect to persistence",
				},
				cli.StringFlag{
					Name:  FlagDateFormat,
					Usage: "create buckets using time format. Use Go reference time: Mon Jan 2 15:04:05 MST 2006. If set, --" + FlagBucketSize + " is ignored",
				},
				cli.StringFlag{
					Name:  FlagBucketSize,
					Value: "hour",
					Usage: "group timers by time bucket. Available: day, hour, minute, second",
				},
				cli.IntFlag{
					Name:  FlagShardMultiplier,
					Usage: "multiply timer counters for histogram",
					Value: 16384,
				},
			),
			Action: func(c *cli.Context) {
				AdminTimers(c)
			},
		},
	}
}

func newAdminHistoryHostCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "Describe internal information of history host",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagHistoryAddressWithAlias,
					Usage: "History Host address(IP:PORT)",
				},
				cli.IntFlag{
					Name:  FlagShardIDWithAlias,
					Usage: "ShardID",
				},
				cli.BoolFlag{
					Name:  FlagPrintFullyDetailWithAlias,
					Usage: "Print fully detail",
				},
			},
			Action: func(c *cli.Context) {
				AdminDescribeHistoryHost(c)
			},
		},
		{
			Name:    "getshard",
			Aliases: []string{"gsh"},
			Usage:   "Get shardID for a workflowID",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.IntFlag{
					Name:  FlagNumberOfShards,
					Usage: "NumberOfShards for the cadence cluster(see config for numHistoryShards)",
				},
			},
			Action: func(c *cli.Context) {
				AdminGetShardID(c)
			},
		},
	}
}

func newAdminDomainCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "register",
			Aliases: []string{"re"},
			Usage:   "Register workflow domain",
			Flags:   adminRegisterDomainFlags,
			Action: func(c *cli.Context) {
				newDomainCLI(c, true).RegisterDomain(c)
			},
		},
		{
			Name:    "update",
			Aliases: []string{"up", "u"},
			Usage:   "Update existing workflow domain",
			Flags:   adminUpdateDomainFlags,
			Action: func(c *cli.Context) {
				newDomainCLI(c, true).UpdateDomain(c)
			},
		},
		{
			Name:    "deprecate",
			Aliases: []string{"dep"},
			Usage:   "Deprecate existing workflow domain",
			Flags:   adminDeprecateDomainFlags,
			Action: func(c *cli.Context) {
				newDomainCLI(c, true).DeprecateDomain(c)
			},
		},
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "Describe existing workflow domain",
			Flags:   adminDescribeDomainFlags,
			Action: func(c *cli.Context) {
				newDomainCLI(c, true).DescribeDomain(c)
			},
		},
		{
			Name:    "getdomainidorname",
			Aliases: []string{"getdn"},
			Usage:   "Get domainID or domainName",
			Flags: append(getDBFlags(),
				cli.StringFlag{
					Name:  FlagDomain,
					Usage: "DomainName",
				},
				cli.StringFlag{
					Name:  FlagDomainID,
					Usage: "Domain ID(uuid)",
				}),
			Action: func(c *cli.Context) {
				AdminGetDomainIDOrName(c)
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List all domains in the cluster",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  FlagPageSizeWithAlias,
					Value: 10,
					Usage: "Result page size",
				},
				cli.BoolFlag{
					Name:  FlagAllWithAlias,
					Usage: "List all domains, by default only domains in REGISTERED status are listed",
				},
				cli.BoolFlag{
					Name:  FlagDeprecatedWithAlias,
					Usage: "List deprecated domains only, by default only domains in REGISTERED status are listed",
				},
				cli.StringFlag{
					Name:  FlagPrefix,
					Usage: "List domains that are matching to the given prefix",
					Value: "",
				},
				cli.BoolFlag{
					Name:  FlagPrintFullyDetailWithAlias,
					Usage: "Print full domain detail",
				},
				cli.BoolFlag{
					Name:  FlagPrintJSONWithAlias,
					Usage: "Print in raw json format (DEPRECATED: instead use --format json)",
				},
				getFormatFlag(),
			},
			Action: func(c *cli.Context) {
				newDomainCLI(c, false).ListDomains(c)
			},
		},
	}
}

func newAdminKafkaCommands() []cli.Command {
	return []cli.Command{
		{
			// TODO: do we still need this command given that kafka replication has been deprecated?
			Name:    "parse",
			Aliases: []string{"par"},
			Usage:   "Parse replication tasks from kafka messages",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input file to use, if not present assumes piping",
				},
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID, if not provided then no filters by WorkflowID are applied",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID, if not provided then no filters by RunID are applied",
				},
				cli.StringFlag{
					Name:  FlagOutputFilenameWithAlias,
					Usage: "Output file to write to, if not provided output is written to stdout",
				},
				cli.BoolFlag{
					Name:  FlagSkipErrorModeWithAlias,
					Usage: "Skip errors in parsing messages",
				},
				cli.BoolFlag{
					Name:  FlagHeadersModeWithAlias,
					Usage: "Output headers of messages in format: DomainID, WorkflowID, RunID, FirstEventID, NextEventID",
				},
				cli.IntFlag{
					Name:  FlagMessageTypeWithAlias,
					Usage: "Kafka message type (0: replicationTasks; 1: visibility)",
					Value: 0,
				},
			},
			Action: func(c *cli.Context) {
				AdminKafkaParse(c)
			},
		},
		{
			// TODO: move this command be a subcommand of admin workflow
			Name:    "rereplicate",
			Aliases: []string{"rrp"},
			Usage:   "Rereplicate replication tasks from history tables",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagSourceCluster,
					Usage: "Name of source cluster to resend the replication task",
				},
				cli.StringFlag{
					Name:  FlagDomainID,
					Usage: "DomainID",
				},
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowID",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunID",
				},
				cli.Int64Flag{
					Name:  FlagMaxEventID,
					Usage: "MaxEventID Optional, default to all events",
				},
				cli.StringFlag{
					Name:  FlagEndEventVersion,
					Usage: "Workflow end event version, required if MaxEventID is specified",
				}},
			Action: func(c *cli.Context) {
				AdminRereplicate(c)
			},
		},
	}
}

func newAdminElasticSearchCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "catIndex",
			Aliases: []string{"cind"},
			Usage:   "Cat Indices on ElasticSearch",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagURL,
					Usage: "URL of ElasticSearch cluster",
				},
				getFormatFlag(),
			},
			Action: func(c *cli.Context) {
				AdminCatIndices(c)
			},
		},
		{
			Name:    "index",
			Aliases: []string{"ind"},
			Usage:   "Index docs on ElasticSearch",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagURL,
					Usage: "URL of ElasticSearch cluster",
				},
				cli.StringFlag{
					Name:  FlagIndex,
					Usage: "ElasticSearch target index",
				},
				cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input file of indexer.Message in json format, separated by newline",
				},
				cli.IntFlag{
					Name:  FlagBatchSizeWithAlias,
					Usage: "Optional batch size of actions for bulk operations",
					Value: 1000,
				},
			},
			Action: func(c *cli.Context) {
				AdminIndex(c)
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"del"},
			Usage:   "Delete docs on ElasticSearch",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagURL,
					Usage: "URL of ElasticSearch cluster",
				},
				cli.StringFlag{
					Name:  FlagIndex,
					Usage: "ElasticSearch target index",
				},
				cli.StringFlag{
					Name: FlagInputFileWithAlias,
					Usage: "Input file name. Redirect cadence wf list result (with tale format) to a file and use as delete input. " +
						"First line should be table header like WORKFLOW TYPE | WORKFLOW ID | RUN ID | ...",
				},
				cli.IntFlag{
					Name:  FlagBatchSizeWithAlias,
					Usage: "Optional batch size of actions for bulk operations",
					Value: 1000,
				},
				cli.IntFlag{
					Name:  FlagRPS,
					Usage: "Optional batch request rate per second",
					Value: 30,
				},
			},
			Action: func(c *cli.Context) {
				AdminDelete(c)
			},
		},
		{
			Name:    "report",
			Aliases: []string{"rep"},
			Usage:   "Generate Report by Aggregation functions on ElasticSearch",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagURL,
					Usage: "URL of ElasticSearch cluster",
				},
				cli.StringFlag{
					Name:  FlagIndex,
					Usage: "ElasticSearch target index",
				},
				cli.StringFlag{
					Name:  FlagListQuery,
					Usage: "SQL query of the report",
				},
				cli.StringFlag{
					Name:  FlagOutputFormat,
					Usage: "Additional output format (html or csv)",
				},
				cli.StringFlag{
					Name:  FlagOutputFilename,
					Usage: "Additional output filename with path",
				},
			},
			Action: func(c *cli.Context) {
				GenerateReport(c)
			},
		},
	}
}

func newAdminTaskListCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "Describe pollers and status information of tasklist",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagTaskListWithAlias,
					Usage: "TaskList description",
				},
				cli.StringFlag{
					Name:  FlagTaskListTypeWithAlias,
					Value: "decision",
					Usage: "Optional TaskList type [decision|activity]",
				},
			},
			Action: func(c *cli.Context) {
				AdminDescribeTaskList(c)
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List active tasklist under a domain",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagDomainWithAlias,
					Usage: "Required Domain name",
				},
			},
			Action: func(c *cli.Context) {
				AdminListTaskList(c)
			},
		},
	}
}

func newAdminClusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "add-search-attr",
			Aliases: []string{"asa"},
			Usage:   "whitelist search attribute",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagSearchAttributesKey,
					Usage: "Search Attribute key to be whitelisted",
				},
				cli.IntFlag{
					Name:  FlagSearchAttributesType,
					Value: -1,
					Usage: "Search Attribute value type. [0:String, 1:Keyword, 2:Int, 3:Double, 4:Bool, 5:Datetime]",
				},
				cli.StringFlag{
					Name:  FlagSecurityTokenWithAlias,
					Usage: "Optional token for security check",
				},
			},
			Action: func(c *cli.Context) {
				AdminAddSearchAttribute(c)
			},
		},
		{
			Name:    "describe",
			Aliases: []string{"d"},
			Usage:   "Describe cluster information",
			Action: func(c *cli.Context) {
				AdminDescribeCluster(c)
			},
		},
		{
			Name:        "failover",
			Aliases:     []string{"fo"},
			Usage:       "Failover domains with domain data IsManagedByCadence=true to target cluster",
			Subcommands: newAdminFailoverCommands(),
		},
		{
			Name:    "failover_fast",
			Aliases: []string{"fof"},
			Usage:   "Failover domains with domain data IsManagedByCadence=true to target cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagTargetClusterWithAlias,
					Usage: "Target active cluster name",
				},
			},
			Action: func(c *cli.Context) {
				newDomainCLI(c, false).FailoverDomains(c)
			},
		},
		{
			Name:        "rebalance",
			Aliases:     []string{"rb"},
			Usage:       "Rebalance the domains active cluster",
			Subcommands: newAdminRebalanceCommands(),
		},
	}
}

func getDLQFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  FlagShards,
			Usage: "Comma separated shard IDs or inclusive ranges. Example: \"2,5-6,10\".  Alternatively, feed one shard ID per line via STDIN.",
		},
		cli.StringFlag{
			Name:  FlagDLQTypeWithAlias,
			Usage: "Type of DLQ to manage. (Options: domain, history)",
			Value: "history",
		},
		cli.StringFlag{
			Name:  FlagSourceCluster,
			Usage: "The cluster where the task is generated",
		},
		cli.IntFlag{
			Name:  FlagLastMessageIDWithAlias,
			Usage: "The upper boundary of the read message",
		},
	}
}

func newAdminDLQCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "count",
			Aliases: []string{"c"},
			Usage:   "Count DLQ Messages",
			Flags: []cli.Flag{
				getFormatFlag(),
				cli.StringFlag{
					Name:  FlagDLQTypeWithAlias,
					Usage: "Type of DLQ to manage. (Options: domain, history)",
					Value: "history",
				},
				cli.BoolFlag{
					Name:  FlagForce,
					Usage: "Force fetch latest counts (will put additional stress on DB)",
				},
			},
			Action: func(c *cli.Context) {
				AdminCountDLQMessages(c)
			},
		},
		{
			Name:    "read",
			Aliases: []string{"r"},
			Usage:   "Read DLQ Messages",
			Flags: append(getDLQFlags(),
				cli.IntFlag{
					Name:  FlagMaxMessageCountWithAlias,
					Usage: "Max message size to fetch",
				},
				getFormatFlag(),
			),
			Action: func(c *cli.Context) {
				AdminGetDLQMessages(c)
			},
		},
		{
			Name:    "purge",
			Aliases: []string{"p"},
			Usage:   "Delete DLQ messages with equal or smaller ids than the provided task id",
			Flags:   getDLQFlags(),
			Action: func(c *cli.Context) {
				AdminPurgeDLQMessages(c)
			},
		},
		{
			Name:    "merge",
			Aliases: []string{"m"},
			Usage:   "Merge DLQ messages with equal or smaller ids than the provided task id",
			Flags:   getDLQFlags(),
			Action: func(c *cli.Context) {
				AdminMergeDLQMessages(c)
			},
		},
	}
}

func newAdminQueueCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "reset",
			Usage: "reset processing queue states for transfer or timer queue processor",
			Flags: getQueueCommandFlags(),
			Action: func(c *cli.Context) {
				AdminResetQueue(c)
			},
		},
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "describe processing queue states for transfer or timer queue processor",
			Flags:   getQueueCommandFlags(),
			Action: func(c *cli.Context) {
				AdminDescribeQueue(c)
			},
		},
	}
}

func newDBCommands() []cli.Command {
	var collections cli.StringSlice = invariant.CollectionStrings()

	scanFlag := cli.StringFlag{
		Name:     FlagScanType,
		Usage:    "Scan type to use: " + strings.Join(executions.ScanTypeStrings(), ", "),
		Required: true,
	}

	collectionsFlag := cli.StringSliceFlag{
		Name:  FlagInvariantCollection,
		Usage: "Scan collection type to use: " + strings.Join(collections, ", "),
		Value: &collections,
	}

	return []cli.Command{
		{
			Name:  "scan",
			Usage: "scan executions in database and detect corruptions",
			Flags: append(getDBFlags(),
				cli.IntFlag{
					Name:     FlagNumberOfShards,
					Usage:    "NumberOfShards for the cadence cluster (see config for numHistoryShards)",
					Required: true,
				},
				scanFlag,
				collectionsFlag,
				cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input file of executions to scan in JSON format {\"DomainID\":\"x\",\"WorkflowID\":\"x\",\"RunID\":\"x\"} separated by a newline",
				},
			),

			Action: func(c *cli.Context) {
				AdminDBScan(c)
			},
		},
		{
			Name:  "unsupported-workflow",
			Usage: "use this command when upgrade the Cadence server from version less than 0.16.0. This scan database and detect unsupported workflow type.",
			Flags: append(getDBFlags(),
				cli.IntFlag{
					Name:  FlagRPS,
					Usage: "NumberOfShards for the cadence cluster (see config for numHistoryShards)",
					Value: 1000,
				},
				cli.StringFlag{
					Name:  FlagOutputFilenameWithAlias,
					Usage: "Output file to write to, if not provided output is written to stdout",
				},
				cli.IntFlag{
					Name:     FlagLowerShardBound,
					Usage:    "FlagLowerShardBound for the start shard to scan. (Default: 0)",
					Value:    0,
					Required: true,
				},
				cli.IntFlag{
					Name:     FlagUpperShardBound,
					Usage:    "FlagLowerShardBound for the end shard to scan. (Default: 16383)",
					Value:    16383,
					Required: true,
				},
			),

			Action: func(c *cli.Context) {
				AdminDBScanUnsupportedWorkflow(c)
			},
		},
		{
			Name:  "clean",
			Usage: "clean up corrupted workflows",
			Flags: append(getDBFlags(),
				scanFlag,
				collectionsFlag,
				cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input file of execution to clean in JSON format. Use `scan` command to generate list of executions.",
				},
			),
			Action: func(c *cli.Context) {
				AdminDBClean(c)
			},
		},
		{
			Name:  "decode_thrift",
			Usage: "decode thrift object of HEX, print into JSON if the HEX data is matching with any supported struct",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   FlagInputWithAlias,
					EnvVar: "Input",
					Usage:  "HEX input of the binary data, you may get from database query like SELECT HEX(...) FROM ...",
				},
			},
			Action: func(c *cli.Context) {
				AdminDBDataDecodeThrift(c)
			},
		},
	}
}

func getQueueCommandFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  FlagShardIDWithAlias,
			Usage: "shardID",
		},
		cli.StringFlag{
			Name:  FlagCluster,
			Usage: "cluster the task processor is responsible for",
		},
		cli.IntFlag{
			Name:  FlagQueueType,
			Usage: "queue type: 2 (transfer queue), 3 (timer queue) or 6 (cross-cluster queue)",
		},
	}
}

func newAdminFailoverCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "start failover workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagTargetClusterWithAlias,
					Usage: "Target cluster name",
				},
				cli.StringFlag{
					Name:  FlagSourceClusterWithAlias,
					Usage: "Source cluster name",
				},
				cli.IntFlag{
					Name:  FlagFailoverTimeoutWithAlias,
					Usage: "Optional graceful failover timeout in seconds. If this field is define, the failover will use graceful failover.",
				},
				cli.IntFlag{
					Name:  FlagExecutionTimeoutWithAlias,
					Usage: "Optional Failover workflow timeout in seconds",
					Value: defaultFailoverWorkflowTimeoutInSeconds,
				},
				cli.IntFlag{
					Name:  FlagFailoverWaitTimeWithAlias,
					Usage: "Optional Failover wait time after each batch in seconds",
					Value: defaultBatchFailoverWaitTimeInSeconds,
				},
				cli.IntFlag{
					Name:  FlagFailoverBatchSizeWithAlias,
					Usage: "Optional number of domains to failover in one batch",
					Value: defaultBatchFailoverSize,
				},
				cli.StringSliceFlag{
					Name: FlagFailoverDomains,
					Usage: "Optional domains to failover, eg d1,d2..,dn. " +
						"Only provided domains in source cluster will be failover.",
				},
				cli.IntFlag{
					Name: FlagFailoverDrillWaitTimeWithAlias,
					Usage: "Optional failover drill wait time. " +
						"After the wait time, the domains will be reset to original regions." +
						"This field is required if the cron schedule is specified.",
				},
				cli.StringFlag{
					Name: FlagCronSchedule,
					Usage: "Optional cron schedule on failover drill. Please specify failover drill wait time " +
						"if this field is specific",
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverStart(c)
			},
		},
		{
			Name:    "pause",
			Aliases: []string{"p"},
			Usage:   "pause failover workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "Optional Failover workflow runID, default is latest runID",
				},
				cli.BoolFlag{
					Name: FlagFailoverDrillWithAlias,
					Usage: "Optional to pause failover workflow or failover drill workflow." +
						" The default is normal failover workflow",
				},
			},

			Action: func(c *cli.Context) {
				AdminFailoverPause(c)
			},
		},
		{
			Name:    "resume",
			Aliases: []string{"re"},
			Usage:   "resume paused failover workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "Optional Failover workflow runID, default is latest runID",
				},
				cli.BoolFlag{
					Name: FlagFailoverDrillWithAlias,
					Usage: "Optional to resume failover workflow or failover drill workflow." +
						" The default is normal failover workflow",
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverResume(c)
			},
		},
		{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "query failover workflow state",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: FlagFailoverDrillWithAlias,
					Usage: "Optional to query failover workflow or failover drill workflow." +
						" The default is normal failover workflow",
				},
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "Optional Failover workflow runID, default is latest runID",
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverQuery(c)
			},
		},
		{
			Name:    "abort",
			Aliases: []string{"a"},
			Usage:   "abort failover workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "Optional Failover workflow runID, default is latest runID",
				},
				cli.StringFlag{
					Name:  FlagReasonWithAlias,
					Usage: "Optional reason why abort",
				},
				cli.BoolFlag{
					Name: FlagFailoverDrillWithAlias,
					Usage: "Optional to abort failover workflow or failover drill workflow." +
						" The default is normal failover workflow",
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverAbort(c)
			},
		},
		{
			Name:    "rollback",
			Aliases: []string{"ro"},
			Usage:   "rollback failover workflow",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "Optional Failover workflow runID, default is latest runID",
				},
				cli.IntFlag{
					Name:  FlagFailoverTimeoutWithAlias,
					Usage: "Optional graceful failover timeout in seconds. If this field is define, the failover will use graceful failover.",
				},
				cli.IntFlag{
					Name:  FlagExecutionTimeoutWithAlias,
					Usage: "Optional Failover workflow timeout in seconds",
					Value: defaultFailoverWorkflowTimeoutInSeconds,
				},
				cli.IntFlag{
					Name:  FlagFailoverWaitTimeWithAlias,
					Usage: "Optional Failover wait time after each batch in seconds",
					Value: defaultBatchFailoverWaitTimeInSeconds,
				},
				cli.IntFlag{
					Name:  FlagFailoverBatchSizeWithAlias,
					Usage: "Optional number of domains to failover in one batch",
					Value: defaultBatchFailoverSize,
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverRollback(c)
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list failover workflow runs closed/open. This is just a simplified list cmd",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  FlagOpenWithAlias,
					Usage: "List for open workflow executions, default is to list for closed ones",
				},
				cli.IntFlag{
					Name:  FlagPageSizeWithAlias,
					Value: 10,
					Usage: "Result page size",
				},
				cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "Ignore this. It is a dummy flag which will be forced overwrite",
				},
				cli.BoolFlag{
					Name: FlagFailoverDrillWithAlias,
					Usage: "Optional to query failover workflow or failover drill workflow." +
						" The default is normal failover workflow",
				},
			},
			Action: func(c *cli.Context) {
				AdminFailoverList(c)
			},
		},
	}
}

func newAdminRebalanceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "start rebalance workflow",
			Flags:   []cli.Flag{},
			Action: func(c *cli.Context) {
				AdminRebalanceStart(c)
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list rebalance workflow runs closed/open.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  FlagOpenWithAlias,
					Usage: "List for open workflow executions, default is to list for closed ones",
				},
				cli.IntFlag{
					Name:  FlagPageSizeWithAlias,
					Value: 10,
					Usage: "Result page size",
				},
			},
			Action: func(c *cli.Context) {
				AdminRebalanceList(c)
			},
		},
	}
}

func newAdminConfigStoreCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "get-dynamic-config",
			Aliases: []string{"getdc", "g"},
			Usage:   "Get Dynamic Config Value",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagDynamicConfigName,
					Usage: "Name of Dynamic Config parameter to get value of",
				},
				cli.StringSliceFlag{
					Name:  FlagDynamicConfigFilter,
					Usage: `Optional. Can be specified multiple times for multiple filters. ex: --dynamic-config-filter '{"Name":"domainName","Value":"global-samples-domain"}'`,
				},
			},
			Action: func(c *cli.Context) {
				AdminGetDynamicConfig(c)
			},
		},
		{
			Name:    "update-dynamic-config",
			Aliases: []string{"updc", "u"},
			Usage:   "Update Dynamic Config Value",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagDynamicConfigName,
					Usage: "Name of Dynamic Config parameter to update value of",
				},
				cli.StringSliceFlag{
					Name:  FlagDynamicConfigValue,
					Usage: `Optional. Can be specified multiple times for multiple values. ex: --dynamic-config-value '{"Value":true,"Filters":[]}'`,
				},
			},
			Action: func(c *cli.Context) {
				AdminUpdateDynamicConfig(c)
			},
		},
		{
			Name:    "restore-dynamic-config",
			Aliases: []string{"resdc", "r"},
			Usage:   "Restore Dynamic Config Value",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagDynamicConfigName,
					Usage: "Name of Dynamic Config parameter to restore",
				},
				cli.StringSliceFlag{
					Name:  FlagDynamicConfigFilter,
					Usage: `Optional. Can be specified multiple times for multiple filters. ex: --dynamic-config-filter '{"Name":"domainName","Value":"global-samples-domain"}'`,
				},
			},
			Action: func(c *cli.Context) {
				AdminRestoreDynamicConfig(c)
			},
		},
		{
			Name:    "list-dynamic-config",
			Aliases: []string{"listdc", "l"},
			Usage:   "List Dynamic Config Value",
			Flags:   []cli.Flag{},
			Action: func(c *cli.Context) {
				AdminListDynamicConfig(c)
			},
		},
	}
}
