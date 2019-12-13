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

import "github.com/urfave/cli/v2"

// Flags used to specify cli command line arguments
const (
	FlagUsername                  = "username"
	FlagPassword                  = "password"
	FlagKeyspace                  = "keyspace"
	FlagAddress                   = "address"
	FlagHistoryAddress            = "history_address"
	FlagDBAddress                 = "db_address"
	FlagDBPort                    = "db_port"
	FlagDomainID                  = "domain_id"
	FlagDomain                    = "domain"
	FlagShardID                   = "shard_id"
	FlagWorkflowID                = "workflow_id"
	FlagRunID                     = "run_id"
	FlagTreeID                    = "tree_id"
	FlagBranchID                  = "branch_id"
	FlagNumberOfShards            = "number_of_shards"
	FlagTargetCluster             = "target_cluster"
	FlagMinEventID                = "min_event_id"
	FlagMaxEventID                = "max_event_id"
	FlagTaskList                  = "tasklist"
	FlagTaskListType              = "tasklisttype"
	FlagWorkflowIDReusePolicy     = "workflowidreusepolicy"
	FlagCronSchedule              = "cron"
	FlagWorkflowType              = "workflow_type"
	FlagWorkflowStatus            = "status"
	FlagExecutionTimeout          = "execution_timeout"
	FlagDecisionTimeout           = "decision_timeout"
	FlagContextTimeout            = "context_timeout"
	FlagInput                     = "input"
	FlagInputFile                 = "input_file"
	FlagExcludeFile               = "exclude_file"
	FlagInputSeparator            = "input_separator"
	FlagParallism                 = "input_parallism"
	FlagSkipCurrent               = "skip_current_open"
	FlagInputTopic                = "input_topic"
	FlagHostFile                  = "host_file"
	FlagCluster                   = "cluster"
	FlagInputCluster              = "input_cluster"
	FlagStartOffset               = "start_offset"
	FlagTopic                     = "topic"
	FlagGroup                     = "group"
	FlagResult                    = "result"
	FlagIdentity                  = "identity"
	FlagDetail                    = "detail"
	FlagReason                    = "reason"
	FlagOpen                      = "open"
	FlagMore                      = "more"
	FlagAll                       = "all"
	FlagPageSize                  = "pagesize"
	FlagEarliestTime              = "earliest_time"
	FlagLatestTime                = "latest_time"
	FlagPrintEventVersion         = "print_event_version"
	FlagPrintFullyDetail          = "print_full"
	FlagPrintRawTime              = "print_raw_time"
	FlagPrintRaw                  = "print_raw"
	FlagPrintDateTime             = "print_datetime"
	FlagPrintMemo                 = "print_memo"
	FlagPrintSearchAttr           = "print_search_attr"
	FlagPrintJSON                 = "print_json"
	FlagDescription               = "description"
	FlagOwnerEmail                = "owner_email"
	FlagRetentionDays             = "retention"
	FlagEmitMetric                = "emit_metric"
	FlagHistoryArchivalStatus     = "history_archival_status"
	FlagHistoryArchivalURI        = "history_uri"
	FlagVisibilityArchivalStatus  = "visibility_archival_status"
	FlagVisibilityArchivalURI     = "visibility_uri"
	FlagName                      = "name"
	FlagOutputFilename            = "output_filename"
	FlagOutputFormat              = "output"
	FlagQueryType                 = "query_type"
	FlagQueryRejectCondition      = "query_reject_condition"
	FlagQueryConsistencyLevel     = "query_consistency_level"
	FlagShowDetail                = "show_detail"
	FlagActiveClusterName         = "active_cluster"
	FlagClusters                  = "clusters"
	FlagIsGlobalDomain            = "global_domain"
	FlagDomainData                = "domain_data"
	FlagEventID                   = "event_id"
	FlagActivityID                = "activity_id"
	FlagMaxFieldLength            = "max_field_length"
	FlagSecurityToken             = "security_token"
	FlagSkipErrorMode             = "skip_errors"
	FlagHeadersMode               = "headers"
	FlagMessageType               = "message_type"
	FlagURL                       = "url"
	FlagMuttleyDestination        = "muttely_destination"
	FlagIndex                     = "index"
	FlagBatchSize                 = "batch_size"
	FlagMemoKey                   = "memo_key"
	FlagMemo                      = "memo"
	FlagMemoFile                  = "memo_file"
	FlagSearchAttributesKey       = "search_attr_key"
	FlagSearchAttributesVal       = "search_attr_value"
	FlagSearchAttributesType      = "search_attr_type"
	FlagAddBadBinary              = "add_bad_binary"
	FlagRemoveBadBinary           = "remove_bad_binary"
	FlagResetType                 = "reset_type"
	FlagResetPointsOnly           = "reset_points_only"
	FlagResetBadBinaryChecksum    = "reset_bad_binary_checksum"
	FlagListQuery                 = "query"
	FlagBatchType                 = "batch_type"
	FlagSignalName                = "signal_name"
	FlagRemoveTaskID              = "task_id"
	FlagRemoveTypeID              = "type_id"
	FlagRPS                       = "rps"
	FlagJobID                     = "job_id"
	FlagYes                       = "yes"
	FlagServiceConfigDir          = "service_config_dir"
	FlagServiceEnv                = "service_env"
	FlagServiceZone               = "service_zone"
	FlagEnableTLS                 = "tls"
	FlagTLSCertPath               = "tls_cert_path"
	FlagTLSKeyPath                = "tls_key_path"
	FlagTLSCaPath                 = "tls_ca_path"
	FlagTLSEnableHostVerification = "tls_enable_host_verification"
)

// Flag aliases used to specify cli command line arguments
var (
	FlagAddressAlias                  = []string{"ad"}
	FlagHistoryAddressAlias           = []string{"had"}
	FlagDomainAlias                   = []string{"do"}
	FlagShardIDAlias                  = []string{"sid"}
	FlagWorkflowIDAlias               = []string{"wid", "w"}
	FlagRunIDAlias                    = []string{"rid", "r"}
	FlagTaskListAlias                 = []string{"tl"}
	FlagTaskListTypeAlias             = []string{"tlt"}
	FlagWorkflowIDReusePolicyAlias    = []string{"wrp"}
	FlagWorkflowTypeAlias             = []string{"wt"}
	FlagWorkflowStatusAlias           = []string{"s"}
	FlagExecutionTimeoutAlias         = []string{"et"}
	FlagDecisionTimeoutAlias          = []string{"dt"}
	FlagContextTimeoutAlias           = []string{"ct"}
	FlagInputAlias                    = []string{"i"}
	FlagInputFileAlias                = []string{"if"}
	FlagInputTopicAlias               = []string{"it"}
	FlagReasonAlias                   = []string{"re"}
	FlagOpenAlias                     = []string{"op"}
	FlagMoreAlias                     = []string{"m"}
	FlagAllAlias                      = []string{"a"}
	FlagPageSizeAlias                 = []string{"ps"}
	FlagEarliestTimeAlias             = []string{"et"}
	FlagLatestTimeAlias               = []string{"lt"}
	FlagPrintEventVersionAlias        = []string{"pev"}
	FlagPrintFullyDetailAlias         = []string{"pf"}
	FlagPrintRawTimeAlias             = []string{"prt"}
	FlagPrintRawAlias                 = []string{"praw"}
	FlagPrintDateTimeAlias            = []string{"pdt"}
	FlagPrintMemoAlias                = []string{"pme"}
	FlagPrintSearchAttrAlias          = []string{"psa"}
	FlagPrintJSONAlias                = []string{"pjson"}
	FlagDescriptionAlias              = []string{"desc"}
	FlagOwnerEmailAlias               = []string{"oe"}
	FlagRetentionDaysAlias            = []string{"rd"}
	FlagEmitMetricAlias               = []string{"em"}
	FlagHistoryArchivalStatusAlias    = []string{"has"}
	FlagHistoryArchivalURIAlias       = []string{"huri"}
	FlagVisibilityArchivalStatusAlias = []string{"vas"}
	FlagVisibilityArchivalURIAlias    = []string{"vuri"}
	FlagNameAlias                     = []string{"n"}
	FlagOutputFilenameAlias           = []string{"of"}
	FlagQueryTypeAlias                = []string{"qt"}
	FlagQueryRejectConditionAlias     = []string{"qrc"}
	FlagQueryConsistencyLevelAlias    = []string{"qcl"}
	FlagShowDetailAlias               = []string{"sd"}
	FlagActiveClusterNameAlias        = []string{"ac"}
	FlagClustersAlias                 = []string{"cl"}
	FlagIsGlobalDomainAlias           = []string{"gd"}
	FlagDomainDataAlias               = []string{"dmd"}
	FlagEventIDAlias                  = []string{"eid"}
	FlagActivityIDAlias               = []string{"aid"}
	FlagMaxFieldLengthAlias           = []string{"maxl"}
	FlagSecurityTokenAlias            = []string{"st"}
	FlagSkipErrorModeAlias            = []string{"serr"}
	FlagHeadersModeAlias              = []string{"he"}
	FlagMessageTypeAlias              = []string{"mt"}
	FlagMuttleyDestinationAlias       = []string{"muttley"}
	FlagBatchSizeAlias                = []string{"bs"}
	FlagListQueryAlias                = []string{"q"}
	FlagBatchTypeAlias                = []string{"bt"}
	FlagSignalNameAlias               = []string{"sig"}
	FlagJobIDAlias                    = []string{"jid"}
	FlagServiceConfigDirAlias         = []string{"scd"}
	FlagServiceEnvAlias               = []string{"se"}
	FlagServiceZoneAlias              = []string{"sz"}
)

var flagsForExecution = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagWorkflowID,
		Aliases: FlagWorkflowIDAlias,
		Usage:   "WorkflowID",
	},
	&cli.StringFlag{
		Name:    FlagRunID,
		Aliases: FlagRunIDAlias,
		Usage:   "RunID",
	},
}

func getFlagsForShow() []cli.Flag {
	return append(flagsForExecution, getFlagsForShowID()...)
}

func getFlagsForShowID() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagPrintDateTime,
			Aliases: FlagPrintDateTimeAlias,
			Usage:   "Print timestamp",
		},
		&cli.BoolFlag{
			Name:    FlagPrintRawTime,
			Aliases: FlagPrintRawTimeAlias,
			Usage:   "Print raw timestamp",
		},
		&cli.StringFlag{
			Name:    FlagOutputFilename,
			Aliases: FlagOutputFilenameAlias,
			Usage:   "Serialize history event to a file",
		},
		&cli.BoolFlag{
			Name:    FlagPrintFullyDetail,
			Aliases: FlagPrintFullyDetailAlias,
			Usage:   "Print fully event detail",
		},
		&cli.BoolFlag{
			Name:    FlagPrintEventVersion,
			Aliases: FlagPrintEventVersionAlias,
			Usage:   "Print event version",
		},
		&cli.IntFlag{
			Name:    FlagEventID,
			Aliases: FlagEventIDAlias,
			Usage:   "Print specific event details",
		},
		&cli.IntFlag{
			Name:    FlagMaxFieldLength,
			Aliases: FlagMaxFieldLengthAlias,
			Usage:   "Maximum length for each attribute field",
			Value:   defaultMaxFieldLength,
		},
		&cli.BoolFlag{
			Name:  FlagResetPointsOnly,
			Usage: "Only show events that are eligible for reset",
		},
	}
}

func getFlagsForStart() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagTaskList,
			Aliases: FlagTaskListAlias,
			Usage:   "TaskList",
		},
		&cli.StringFlag{
			Name:    FlagWorkflowID,
			Aliases: FlagWorkflowIDAlias,
			Usage:   "WorkflowID",
		},
		&cli.StringFlag{
			Name:    FlagWorkflowType,
			Aliases: FlagWorkflowTypeAlias,
			Usage:   "WorkflowTypeName",
		},
		&cli.IntFlag{
			Name:    FlagExecutionTimeout,
			Aliases: FlagExecutionTimeoutAlias,
			Usage:   "Execution start to close timeout in seconds",
		},
		&cli.IntFlag{
			Name:    FlagDecisionTimeout,
			Aliases: FlagDecisionTimeoutAlias,
			Value:   defaultDecisionTimeoutInSeconds,
			Usage:   "Decision task start to close timeout in seconds",
		},
		&cli.StringFlag{
			Name: FlagCronSchedule,
			Usage: "Optional cron schedule for the workflow. Cron spec is as following: \n" +
				"\t┌───────────── minute (0 - 59) \n" +
				"\t│ ┌───────────── hour (0 - 23) \n" +
				"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
				"\t│ │ │ ┌───────────── month (1 - 12) \n" +
				"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
				"\t│ │ │ │ │ \n" +
				"\t* * * * *",
		},
		&cli.IntFlag{
			Name:    FlagWorkflowIDReusePolicy,
			Aliases: FlagWorkflowIDReusePolicyAlias,
			Usage: "Optional input to configure if the same workflow ID is allow to use for new workflow execution. " +
				"Available options: 0: AllowDuplicateFailedOnly, 1: AllowDuplicate, 2: RejectDuplicate",
		},
		&cli.StringFlag{
			Name:    FlagInput,
			Aliases: FlagInputAlias,
			Usage:   "Optional input for the workflow, in JSON format. If there are multiple parameters, concatenate them and separate by space.",
		},
		&cli.StringFlag{
			Name:    FlagInputFile,
			Aliases: FlagInputFileAlias,
			Usage: "Optional input for the workflow from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
				"Input from file will be overwrite by input from command line",
		},
		&cli.StringFlag{
			Name:  FlagMemoKey,
			Usage: "Optional key of memo. If there are multiple keys, concatenate them and separate by space",
		},
		&cli.StringFlag{
			Name: FlagMemo,
			Usage: "Optional info that can be showed when list workflow, in JSON format. If there are multiple JSON, concatenate them and separate by space. " +
				"The order must be same as memo_key",
		},
		&cli.StringFlag{
			Name: FlagMemoFile,
			Usage: "Optional info that can be listed in list workflow, from JSON format file. If there are multiple JSON, concatenate them and separate by space or newline. " +
				"The order must be same as memo_key",
		},
		&cli.StringFlag{
			Name: FlagSearchAttributesKey,
			Usage: "Optional search attributes keys that can be be used in list query. If there are multiple keys, concatenate them and separate by |. " +
				"Use 'cluster get-search-attr' cmd to list legal keys.",
		},
		&cli.StringFlag{
			Name: FlagSearchAttributesVal,
			Usage: "Optional search attributes value that can be be used in list query. If there are multiple keys, concatenate them and separate by |. " +
				"If value is array, use json array like [\"a\",\"b\"], [1,2], [\"true\",\"false\"], [\"2019-06-07T17:16:34-08:00\",\"2019-06-07T18:16:34-08:00\"]. " +
				"Use 'cluster get-search-attr' cmd to list legal keys and value types",
		},
	}
}

func getFlagsForRun() []cli.Flag {
	flagsForRun := []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagShowDetail,
			Aliases: FlagShowDetailAlias,
			Usage:   "Show event details",
		},
		&cli.IntFlag{
			Name:    FlagMaxFieldLength,
			Aliases: FlagMaxFieldLengthAlias,
			Usage:   "Maximum length for each attribute field",
		},
	}
	flagsForRun = append(getFlagsForStart(), flagsForRun...)
	return flagsForRun
}

func getCommonFlagsForVisibility() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagPrintRawTime,
			Aliases: FlagPrintRawTimeAlias,
			Usage:   "Print raw timestamp",
		},
		&cli.BoolFlag{
			Name:    FlagPrintDateTime,
			Aliases: FlagPrintDateTimeAlias,
			Usage:   "Print full date time in '2006-01-02T15:04:05Z07:00' format",
		},
		&cli.BoolFlag{
			Name:    FlagPrintMemo,
			Aliases: FlagPrintMemoAlias,
			Usage:   "Print memo",
		},
		&cli.BoolFlag{
			Name:    FlagPrintSearchAttr,
			Aliases: FlagPrintSearchAttrAlias,
			Usage:   "Print search attributes",
		},
		&cli.BoolFlag{
			Name:    FlagPrintFullyDetail,
			Aliases: FlagPrintFullyDetailAlias,
			Usage:   "Print full message without table format",
		},
		&cli.BoolFlag{
			Name:    FlagPrintJSON,
			Aliases: FlagPrintJSONAlias,
			Usage:   "Print in raw json format",
		},
	}
}

func getFlagsForList() []cli.Flag {
	flagsForList := []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagMore,
			Aliases: FlagMoreAlias,
			Usage:   "List more pages, default is to list one page of default page size 10",
		},
		&cli.IntFlag{
			Name:    FlagPageSize,
			Aliases: FlagPageSizeAlias,
			Value:   10,
			Usage:   "Result page size",
		},
	}
	flagsForList = append(getFlagsForListAll(), flagsForList...)
	return flagsForList
}

func getFlagsForListAll() []cli.Flag {
	flagsForListAll := []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagOpen,
			Aliases: FlagOpenAlias,
			Usage:   "List for open workflow executions, default is to list for closed ones",
		},
		&cli.StringFlag{
			Name:    FlagEarliestTime,
			Aliases: FlagEarliestTimeAlias,
			Usage:   "EarliestTime of start time, supported formats are '2006-01-02T15:04:05+07:00' and raw UnixNano",
		},
		&cli.StringFlag{
			Name:    FlagLatestTime,
			Aliases: FlagLatestTimeAlias,
			Usage:   "LatestTime of start time, supported formats are '2006-01-02T15:04:05+07:00' and raw UnixNano",
		},
		&cli.StringFlag{
			Name:    FlagWorkflowID,
			Aliases: FlagWorkflowIDAlias,
			Usage:   "WorkflowID",
		},
		&cli.StringFlag{
			Name:    FlagWorkflowType,
			Aliases: FlagWorkflowTypeAlias,
			Usage:   "WorkflowTypeName",
		},
		&cli.StringFlag{
			Name:    FlagWorkflowStatus,
			Aliases: FlagWorkflowStatusAlias,
			Usage:   "Closed workflow status [completed, failed, canceled, terminated, continuedasnew, timedout]",
		},
		&cli.StringFlag{
			Name:    FlagListQuery,
			Aliases: FlagListQueryAlias,
			Usage: "Optional SQL like query for use of search attributes. NOTE: using query will ignore all other filter flags including: " +
				"[open, earliest_time, latest_time, workflow_id, workflow_type]",
		},
	}
	flagsForListAll = append(getCommonFlagsForVisibility(), flagsForListAll...)
	return flagsForListAll
}

func getFlagsForScan() []cli.Flag {
	flagsForScan := []cli.Flag{
		&cli.IntFlag{
			Name:    FlagPageSize,
			Aliases: FlagPageSizeAlias,
			Value:   2000,
			Usage:   "Page size for each Scan API call",
		},
		&cli.StringFlag{
			Name:    FlagListQuery,
			Aliases: FlagListQueryAlias,
			Usage:   "Optional SQL like query",
		},
	}
	flagsForScan = append(getCommonFlagsForVisibility(), flagsForScan...)
	return flagsForScan
}

func getFlagsForListArchived() []cli.Flag {
	flagsForListArchived := []cli.Flag{
		&cli.StringFlag{
			Name:    FlagListQuery,
			Aliases: FlagListQueryAlias,
			Usage:   "SQL like query. Please check the documentation of the visibility archiver used by your domain for detailed instructions",
		},
		&cli.IntFlag{
			Name:    FlagPageSize,
			Aliases: FlagPageSizeAlias,
			Value:   100,
			Usage:   "Count of visibility records included in a single page, default to 100",
		},
		&cli.BoolFlag{
			Name:    FlagAll,
			Aliases: FlagAllAlias,
			Usage:   "List all pages",
		},
	}
	flagsForListArchived = append(getCommonFlagsForVisibility(), flagsForListArchived...)
	return flagsForListArchived
}

func getFlagsForCount() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagListQuery,
			Aliases: FlagListQueryAlias,
			Usage:   "Optional SQL like query. e.g count all open workflows 'CloseTime = missing'; 'WorkflowType=\"wtype\" and CloseTime > 0'",
		},
	}
}

func getFlagsForQuery() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagWorkflowID,
			Aliases: FlagWorkflowIDAlias,
			Usage:   "WorkflowID",
		},
		&cli.StringFlag{
			Name:    FlagRunID,
			Aliases: FlagRunIDAlias,
			Usage:   "RunID",
		},
		&cli.StringFlag{
			Name:    FlagQueryType,
			Aliases: FlagQueryTypeAlias,
			Usage:   "The query type you want to run",
		},
		&cli.StringFlag{
			Name:    FlagInput,
			Aliases: FlagInputAlias,
			Usage:   "Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space.",
		},
		&cli.StringFlag{
			Name:    FlagInputFile,
			Aliases: FlagInputFileAlias,
			Usage: "Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
				"Input from file will be overwrite by input from command line",
		},
		&cli.StringFlag{
			Name:    FlagQueryRejectCondition,
			Aliases: FlagQueryRejectConditionAlias,
			Usage:   "Optional flag to reject queries based on workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"",
		},
		&cli.StringFlag{
			Name:    FlagQueryConsistencyLevel,
			Aliases: FlagQueryConsistencyLevelAlias,
			Usage:   "Optional flag to set query consistency level. Valid values are \"eventual\" and \"strong\"",
		},
	}
}

// all flags of query except QueryType
func getFlagsForStack() []cli.Flag {
	flags := getFlagsForQuery()
	flagsToBeIgnored := map[string]bool{}
	flagsToBeIgnored[FlagQueryType] = true
	for _, flag := range FlagQueryTypeAlias {
		flagsToBeIgnored[flag] = true
	}

	for i := 0; i < len(flags); i++ {
		for _, flagName := range flags[i].Names() {
			if _, ok := flagsToBeIgnored[flagName]; ok {
				return append(flags[:i], flags[i+1:]...)
			}
		}
	}
	return flags
}

func getFlagsForDescribe() []cli.Flag {
	return append(flagsForExecution, getFlagsForDescribeID()...)
}

func getFlagsForDescribeID() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagPrintRaw,
			Aliases: FlagPrintRawAlias,
			Usage:   "Print properties as they are stored",
		},
		&cli.BoolFlag{
			Name:  FlagResetPointsOnly,
			Usage: "Only show auto-reset points",
		},
	}
}

func getFlagsForObserve() []cli.Flag {
	return append(flagsForExecution, getFlagsForObserveID()...)
}

func getFlagsForObserveID() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagShowDetail,
			Aliases: FlagShowDetailAlias,
			Usage:   "Optional show event details",
		},
		&cli.IntFlag{
			Name:    FlagMaxFieldLength,
			Aliases: FlagMaxFieldLengthAlias,
			Usage:   "Optional maximum length for each attribute field when show details",
		},
	}
}
