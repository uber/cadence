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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	cassandra_db "github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
	"github.com/uber/cadence/tools/common/flag"
)

const (
	maxEventID      = 9999
	tableRenderSize = 10
)

// AdminShowWorkflow shows history
func AdminShowWorkflow(c *cli.Context) {
	tid := c.String(FlagTreeID)
	bid := c.String(FlagBranchID)
	sid := c.Int(FlagShardID)
	outputFileName := c.String(FlagOutputFilename)

	ctx, cancel := newContext(c)
	defer cancel()
	serializer := persistence.NewPayloadSerializer()
	var history []*persistence.DataBlob
	if len(tid) != 0 {
		thriftrwEncoder := codec.NewThriftRWEncoder()
		histV2 := initializeHistoryManager(c)
		branchToken, err := thriftrwEncoder.Encode(&shared.HistoryBranch{
			TreeID:   &tid,
			BranchID: &bid,
		})
		if err != nil {
			ErrorAndExit("encoding branch token err", err)
		}
		resp, err := histV2.ReadRawHistoryBranch(ctx, &persistence.ReadHistoryBranchRequest{
			BranchToken: branchToken,
			MinEventID:  1,
			MaxEventID:  maxEventID,
			PageSize:    maxEventID,
			ShardID:     &sid,
		})
		if err != nil {
			ErrorAndExit("ReadHistoryBranch err", err)
		}

		history = resp.HistoryEventBlobs
	} else {
		ErrorAndExit("need to specify TreeID/BranchID/ShardID", nil)
	}

	if len(history) == 0 {
		ErrorAndExit("no events", nil)
	}
	allEvents := &shared.History{}
	totalSize := 0
	for idx, b := range history {
		totalSize += len(b.Data)
		fmt.Printf("======== batch %v, blob len: %v ======\n", idx+1, len(b.Data))
		internalHistoryBatch, err := serializer.DeserializeBatchEvents(b)
		if err != nil {
			ErrorAndExit("DeserializeBatchEvents err", err)
		}
		historyBatch := thrift.FromHistoryEventArray(internalHistoryBatch)
		allEvents.Events = append(allEvents.Events, historyBatch...)
		for _, e := range historyBatch {
			jsonstr, err := json.Marshal(e)
			if err != nil {
				ErrorAndExit("json.Marshal err", err)
			}
			fmt.Println(string(jsonstr))
		}
	}
	fmt.Printf("======== total batches %v, total blob len: %v ======\n", len(history), totalSize)

	if outputFileName != "" {
		data, err := json.Marshal(allEvents.Events)
		if err != nil {
			ErrorAndExit("Failed to serialize history data.", err)
		}
		if err := ioutil.WriteFile(outputFileName, data, 0666); err != nil {
			ErrorAndExit("Failed to export history data file.", err)
		}
	}
}

// AdminDescribeWorkflow describe a new workflow execution for admin
func AdminDescribeWorkflow(c *cli.Context) {

	resp := describeMutableState(c)
	prettyPrintJSONObject(resp)

	if resp != nil {
		msStr := resp.GetMutableStateInDatabase()
		ms := persistence.WorkflowMutableState{}
		err := json.Unmarshal([]byte(msStr), &ms)
		if err != nil {
			ErrorAndExit("json.Unmarshal err", err)
		}
		currentBranchToken := ms.ExecutionInfo.BranchToken
		if ms.VersionHistories != nil {
			// if VersionHistories is set, then all branch infos are stored in VersionHistories
			currentVersionHistory, err := ms.VersionHistories.GetCurrentVersionHistory()
			if err != nil {
				ErrorAndExit("ms.VersionHistories.GetCurrentVersionHistory err", err)
			}
			currentBranchToken = currentVersionHistory.GetBranchToken()
		}

		branchInfo := shared.HistoryBranch{}
		thriftrwEncoder := codec.NewThriftRWEncoder()
		err = thriftrwEncoder.Decode(currentBranchToken, &branchInfo)
		if err != nil {
			ErrorAndExit("thriftrwEncoder.Decode err", err)
		}
		prettyPrintJSONObject(branchInfo)
		if ms.ExecutionInfo.AutoResetPoints != nil {
			fmt.Println("auto-reset-points:")
			for _, p := range ms.ExecutionInfo.AutoResetPoints.Points {
				createT := time.Unix(0, p.GetCreatedTimeNano())
				expireT := time.Unix(0, p.GetExpiringTimeNano())
				fmt.Println(p.GetBinaryChecksum(), p.GetRunID(), p.GetFirstDecisionCompletedID(), p.GetResettable(), createT, expireT)
			}
		}
	}
}

func describeMutableState(c *cli.Context) *types.AdminDescribeWorkflowExecutionResponse {
	adminClient := cFactory.ServerAdminClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	ctx, cancel := newContext(c)
	defer cancel()

	resp, err := adminClient.DescribeWorkflowExecution(
		ctx,
		&types.AdminDescribeWorkflowExecutionRequest{
			Domain: domain,
			Execution: &types.WorkflowExecution{
				WorkflowID: wid,
				RunID:      rid,
			},
		},
	)
	if err != nil {
		ErrorAndExit("Get workflow mutableState failed", err)
	}
	return resp
}

// AdminDeleteWorkflow delete a workflow execution for admin
func AdminDeleteWorkflow(c *cli.Context) {
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	resp := describeMutableState(c)
	msStr := resp.GetMutableStateInDatabase()
	ms := persistence.WorkflowMutableState{}
	err := json.Unmarshal([]byte(msStr), &ms)
	if err != nil {
		ErrorAndExit("json.Unmarshal err", err)
	}
	domainID := ms.ExecutionInfo.DomainID
	skipError := c.Bool(FlagSkipErrorMode)

	shardID := resp.GetShardID()
	shardIDInt, err := strconv.Atoi(shardID)
	if err != nil {
		ErrorAndExit("strconv.Atoi(shardID) err", err)
	}
	ctx, cancel := newContext(c)
	defer cancel()
	histV2 := initializeHistoryManager(c)
	defer histV2.Close()
	exeStore := initializeExecutionStore(c, shardIDInt, 0)

	branchInfo := shared.HistoryBranch{}
	thriftrwEncoder := codec.NewThriftRWEncoder()
	branchTokens := [][]byte{ms.ExecutionInfo.BranchToken}
	if ms.VersionHistories != nil {
		// if VersionHistories is set, then all branch infos are stored in VersionHistories
		branchTokens = [][]byte{}
		for _, versionHistory := range ms.VersionHistories.ToInternalType().Histories {
			branchTokens = append(branchTokens, versionHistory.BranchToken)
		}
	}

	for _, branchToken := range branchTokens {
		err = thriftrwEncoder.Decode(branchToken, &branchInfo)
		if err != nil {
			ErrorAndExit("thriftrwEncoder.Decode err", err)
		}
		fmt.Println("deleting history events for ...")
		prettyPrintJSONObject(branchInfo)
		err = histV2.DeleteHistoryBranch(ctx, &persistence.DeleteHistoryBranchRequest{
			BranchToken: branchToken,
			ShardID:     &shardIDInt,
		})
		if err != nil {
			if skipError {
				fmt.Println("failed to delete history, ", err)
			} else {
				ErrorAndExit("DeleteHistoryBranch err", err)
			}
		}
	}

	req := &persistence.DeleteWorkflowExecutionRequest{
		DomainID:   domainID,
		WorkflowID: wid,
		RunID:      rid,
	}

	err = exeStore.DeleteWorkflowExecution(ctx, req)
	if err != nil {
		if skipError {
			fmt.Println("delete mutableState row failed, ", err)
		} else {
			ErrorAndExit("delete mutableState row failed", err)
		}
	}
	fmt.Println("delete mutableState row successfully")

	deleteCurrentReq := &persistence.DeleteCurrentWorkflowExecutionRequest{
		DomainID:   domainID,
		WorkflowID: wid,
		RunID:      rid,
	}

	err = exeStore.DeleteCurrentWorkflowExecution(ctx, deleteCurrentReq)
	if err != nil {
		if skipError {
			fmt.Println("delete current row failed, ", err)
		} else {
			ErrorAndExit("delete current row failed", err)
		}
	}
	fmt.Println("delete current row successfully")
}

func connectToCassandra(c *cli.Context) (nosqlplugin.DB, nosqlplugin.AdminDB) {
	host := getRequiredOption(c, FlagDBAddress)
	if !c.IsSet(FlagDBPort) {
		ErrorAndExit("cassandra port is required", nil)
	}

	cfg := config.NoSQL{
		PluginName:   cassandra_db.PluginName,
		Hosts:        host,
		Port:         c.Int(FlagDBPort),
		Region:       c.String(FlagDBRegion),
		User:         c.String(FlagUsername),
		Password:     c.String(FlagPassword),
		Keyspace:     getRequiredOption(c, FlagKeyspace),
		ProtoVersion: c.Int(FlagProtoVersion),
		MaxConns:     20,
	}
	if c.Bool(FlagEnableTLS) {
		cfg.TLS = &config.TLS{
			Enabled:                true,
			CertFile:               c.String(FlagTLSCertPath),
			KeyFile:                c.String(FlagTLSKeyPath),
			CaFile:                 c.String(FlagTLSCaPath),
			EnableHostVerification: c.Bool(FlagTLSEnableHostVerification),
		}
	}

	db, err := nosql.NewNoSQLDB(&cfg, loggerimpl.NewNopLogger())
	if err != nil {
		ErrorAndExit("connect to Cassandra failed", err)
	}
	adminDB, err := nosql.NewNoSQLAdminDB(&cfg, loggerimpl.NewNopLogger())
	if err != nil {
		ErrorAndExit("connect to Cassandra failed", err)
	}
	return db, adminDB
}

func connectToSQL(c *cli.Context) sqlplugin.DB {
	host := getRequiredOption(c, FlagDBAddress)
	if !c.IsSet(FlagDBPort) {
		ErrorAndExit("sql port is required", nil)
	}
	encodingType := c.String(FlagEncodingType)
	decodingTypesStr := c.StringSlice(FlagDecodingTypes)
	connectAttributes := c.Generic(FlagConnectionAttributes).(*flag.StringMap)

	sqlConfig := &config.SQL{
		ConnectAddr: net.JoinHostPort(
			host,
			c.String(FlagDBPort),
		),
		PluginName:        c.String(FlagDBType),
		User:              c.String(FlagUsername),
		Password:          c.String(FlagPassword),
		DatabaseName:      getRequiredOption(c, FlagDatabaseName),
		EncodingType:      encodingType,
		DecodingTypes:     decodingTypesStr,
		ConnectAttributes: connectAttributes.Value(),
	}

	if c.Bool(FlagEnableTLS) {
		sqlConfig.TLS = &config.TLS{
			Enabled:                true,
			CertFile:               c.String(FlagTLSCertPath),
			KeyFile:                c.String(FlagTLSKeyPath),
			CaFile:                 c.String(FlagTLSCaPath),
			EnableHostVerification: c.Bool(FlagTLSEnableHostVerification),
		}
	}

	db, err := sql.NewSQLDB(sqlConfig)
	if err != nil {
		ErrorAndExit("connect to SQL failed", err)
	}
	return db
}

// AdminGetDomainIDOrName map domain
func AdminGetDomainIDOrName(c *cli.Context) {
	domainID := c.String(FlagDomainID)
	domainName := c.String(FlagDomain)
	if len(domainID) == 0 && len(domainName) == 0 {
		ErrorAndExit("Need either domainName or domainID", nil)
	}

	db, _ := connectToCassandra(c)

	ctx, cancel := newContext(c)
	defer cancel()
	if len(domainID) > 0 {
		domain, err := db.SelectDomain(ctx, &domainID, nil)
		if err != nil {
			ErrorAndExit("SelectDomain error", err)
		}
		fmt.Printf("domainName for domainID %v is %v \n", domainID, domain.Info.Name)
	} else {
		domain, err := db.SelectDomain(ctx, nil, &domainName)
		if err != nil {
			ErrorAndExit("SelectDomain error", err)
		}
		fmt.Printf("domainID for domainName %v is %v \n", domain.Info.ID, domainID)
	}
}

// AdminGetShardID get shardID
func AdminGetShardID(c *cli.Context) {
	wid := getRequiredOption(c, FlagWorkflowID)
	numberOfShards := c.Int(FlagNumberOfShards)

	if numberOfShards <= 0 {
		ErrorAndExit("numberOfShards is required", nil)
		return
	}
	shardID := common.WorkflowIDToHistoryShard(wid, numberOfShards)
	fmt.Printf("ShardID for workflowID: %v is %v \n", wid, shardID)
}

// AdminRemoveTask describes history host
func AdminRemoveTask(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	shardID := getRequiredIntOption(c, FlagShardID)
	taskID := getRequiredInt64Option(c, FlagTaskID)
	typeID := getRequiredIntOption(c, FlagTaskType)
	var visibilityTimestamp int64
	if common.TaskType(typeID) == common.TaskTypeTimer {
		visibilityTimestamp = getRequiredInt64Option(c, FlagTaskVisibilityTimestamp)
	}
	var clusterName string
	if common.TaskType(taskID) == common.TaskTypeCrossCluster {
		clusterName = getRequiredOption(c, FlagCluster)
	}

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.RemoveTaskRequest{
		ShardID:             int32(shardID),
		Type:                common.Int32Ptr(int32(typeID)),
		TaskID:              taskID,
		VisibilityTimestamp: common.Int64Ptr(visibilityTimestamp),
		ClusterName:         clusterName,
	}

	err := adminClient.RemoveTask(ctx, req)
	if err != nil {
		ErrorAndExit("Remove task has failed", err)
	}
}

// AdminDescribeShard describes shard by shard id
func AdminDescribeShard(c *cli.Context) {
	sid := getRequiredIntOption(c, FlagShardID)

	ctx, cancel := newContext(c)
	defer cancel()
	shardManager := initializeShardManager(c)

	getShardReq := &persistence.GetShardRequest{ShardID: sid}
	shard, err := shardManager.GetShard(ctx, getShardReq)
	if err != nil {
		ErrorAndExit("Failed to describe shard.", err)
	}

	prettyPrintJSONObject(shard)
}

// AdminSetShardRangeID set shard rangeID by shard id
func AdminSetShardRangeID(c *cli.Context) {
	sid := getRequiredIntOption(c, FlagShardID)
	rid := getRequiredInt64Option(c, FlagRangeID)

	ctx, cancel := newContext(c)
	defer cancel()
	shardManager := initializeShardManager(c)

	getShardResp, err := shardManager.GetShard(ctx, &persistence.GetShardRequest{ShardID: sid})
	if err != nil {
		ErrorAndExit("Failed to get shardInfo.", err)
	}

	previousRangeID := getShardResp.ShardInfo.RangeID
	updatedShardInfo := getShardResp.ShardInfo
	updatedShardInfo.RangeID = rid
	updatedShardInfo.StolenSinceRenew++
	updatedShardInfo.Owner = ""
	updatedShardInfo.UpdatedAt = time.Now()

	if err := shardManager.UpdateShard(ctx, &persistence.UpdateShardRequest{
		PreviousRangeID: previousRangeID,
		ShardInfo:       updatedShardInfo,
	}); err != nil {
		ErrorAndExit("Failed to reset shard rangeID.", err)
	}

	fmt.Printf("Successfully updated rangeID from %v to %v for shard %v.\n", previousRangeID, rid, sid)
}

// AdminCloseShard closes shard by shard id
func AdminCloseShard(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)
	sid := getRequiredIntOption(c, FlagShardID)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.CloseShardRequest{}
	req.ShardID = int32(sid)

	err := adminClient.CloseShard(ctx, req)
	if err != nil {
		ErrorAndExit("Close shard task has failed", err)
	}
}

// AdminDescribeShardDistribution describes shard distribution
func AdminDescribeShardDistribution(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.DescribeShardDistributionRequest{
		PageSize: int32(c.Int(FlagPageSize)),
		PageID:   int32(c.Int(FlagPageID)),
	}

	resp, err := adminClient.DescribeShardDistribution(ctx, req)
	if err != nil {
		ErrorAndExit("Shard list failed", err)
	}

	fmt.Printf("Total Number of Shards: %d \n", resp.NumberOfShards)
	fmt.Printf("Number of Shards Returned: %d \n", len(resp.Shards))

	if len(resp.Shards) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetColumnSeparator("|")
	header := []string{"ShardID", "Identity"}
	headerColor := []tablewriter.Colors{tableHeaderBlue, tableHeaderBlue}
	table.SetHeader(header)
	table.SetHeaderColor(headerColor...)
	table.SetHeaderLine(false)

	outputPageSize := tableRenderSize
	for shardID, identity := range resp.Shards {
		if outputPageSize == 0 {
			table.Render()
			table.ClearRows()
			if !showNextPage() {
				break
			}
			outputPageSize = tableRenderSize
		}
		table.Append([]string{strconv.Itoa(int(shardID)), identity})
		outputPageSize--
	}
	// output the remaining rows
	table.Render()
	table.ClearRows()
}

// AdminDescribeHistoryHost describes history host
func AdminDescribeHistoryHost(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	wid := c.String(FlagWorkflowID)
	sid := c.Int(FlagShardID)
	addr := c.String(FlagHistoryAddress)
	printFully := c.Bool(FlagPrintFullyDetail)

	if len(wid) == 0 && !c.IsSet(FlagShardID) && len(addr) == 0 {
		ErrorAndExit("at least one of them is required to provide to lookup host: workflowID, shardID and host address", nil)
		return
	}

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.DescribeHistoryHostRequest{}
	if len(wid) > 0 {
		req.ExecutionForHost = &types.WorkflowExecution{WorkflowID: wid}
	}
	if c.IsSet(FlagShardID) {
		req.ShardIDForHost = common.Int32Ptr(int32(sid))
	}
	if len(addr) > 0 {
		req.HostAddress = common.StringPtr(addr)
	}

	resp, err := adminClient.DescribeHistoryHost(ctx, req)
	if err != nil {
		ErrorAndExit("Describe history host failed", err)
	}

	if !printFully {
		resp.ShardIDs = nil
	}
	prettyPrintJSONObject(resp)
}

// AdminRefreshWorkflowTasks refreshes all the tasks of a workflow
func AdminRefreshWorkflowTasks(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	domain := getRequiredGlobalOption(c, FlagDomain)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	ctx, cancel := newContext(c)
	defer cancel()

	err := adminClient.RefreshWorkflowTasks(ctx, &types.RefreshWorkflowTasksRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
	})
	if err != nil {
		ErrorAndExit("Refresh workflow task failed", err)
	} else {
		fmt.Println("Refresh workflow task succeeded.")
	}
}

// AdminResetQueue resets task processing queue states
func AdminResetQueue(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	shardID := getRequiredIntOption(c, FlagShardID)
	clusterName := getRequiredOption(c, FlagCluster)
	typeID := getRequiredIntOption(c, FlagQueueType)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.ResetQueueRequest{
		ShardID:     int32(shardID),
		ClusterName: clusterName,
		Type:        common.Int32Ptr(int32(typeID)),
	}

	err := adminClient.ResetQueue(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to reset queue", err)
	}
	fmt.Println("Reset queue state succeeded")
}

// AdminDescribeQueue describes task processing queue states
func AdminDescribeQueue(c *cli.Context) {
	adminClient := cFactory.ServerAdminClient(c)

	shardID := getRequiredIntOption(c, FlagShardID)
	clusterName := getRequiredOption(c, FlagCluster)
	typeID := getRequiredIntOption(c, FlagQueueType)

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.DescribeQueueRequest{
		ShardID:     int32(shardID),
		ClusterName: clusterName,
		Type:        common.Int32Ptr(int32(typeID)),
	}

	resp, err := adminClient.DescribeQueue(ctx, req)
	if err != nil {
		ErrorAndExit("Failed to describe queue", err)
	}

	for _, state := range resp.ProcessingQueueStates {
		fmt.Println(state)
	}
}
