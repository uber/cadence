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
	"strconv"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	cassp "github.com/uber/cadence/common/persistence/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/persistence/sql"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

const (
	maxEventID = 9999

	cassandraProtoVersion = 4
)

// AdminShowWorkflow shows history
func AdminShowWorkflow(c *cli.Context) {
	tid := c.String(FlagTreeID)
	bid := c.String(FlagBranchID)
	sid := c.Int(FlagShardID)
	outputFileName := c.String(FlagOutputFilename)

	ctx, cancel := newContext(c)
	defer cancel()
	client, session := connectToCassandra(c)
	serializer := persistence.NewPayloadSerializer()
	var history []*persistence.DataBlob
	if len(tid) != 0 {
		histV2 := cassp.NewHistoryV2PersistenceFromSession(client, session, loggerimpl.NewNopLogger())
		resp, err := histV2.ReadHistoryBranch(ctx, &persistence.InternalReadHistoryBranchRequest{
			TreeID:    tid,
			BranchID:  bid,
			MinNodeID: 1,
			MaxNodeID: maxEventID,
			PageSize:  maxEventID,
			ShardID:   sid,
		})
		if err != nil {
			ErrorAndExit("ReadHistoryBranch err", err)
		}

		history = resp.History
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

	resp, err := adminClient.DescribeWorkflowExecution(ctx, &types.AdminDescribeWorkflowExecutionRequest{
		Domain: domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: wid,
			RunID:      rid,
		},
	})
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

	ctx, cancel := newContext(c)
	defer cancel()
	client, session := connectToCassandra(c)
	shardID := resp.GetShardID()
	shardIDInt, err := strconv.Atoi(shardID)
	if err != nil {
		ErrorAndExit("strconv.Atoi(shardID) err", err)
	}

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
		histV2 := cassp.NewHistoryV2PersistenceFromSession(client, session, loggerimpl.NewNopLogger())
		err = histV2.DeleteHistoryBranch(ctx, &persistence.InternalDeleteHistoryBranchRequest{
			BranchInfo: *thrift.ToHistoryBranch(&branchInfo),
			ShardID:    shardIDInt,
		})
		if err != nil {
			if skipError {
				fmt.Println("failed to delete history, ", err)
			} else {
				ErrorAndExit("DeleteHistoryBranch err", err)
			}
		}
	}

	exeStore, _ := cassp.NewWorkflowExecutionPersistence(shardIDInt, client, session, loggerimpl.NewNopLogger())
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

func readOneRow(query gocql.Query) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := query.MapScan(result)
	return result, err
}

func connectToCassandra(c *cli.Context) (gocql.Client, gocql.Session) {
	host := getRequiredOption(c, FlagDBAddress)
	if !c.IsSet(FlagDBPort) {
		ErrorAndExit("cassandra port is required", nil)
	}

	clusterConfig := gocql.ClusterConfig{
		Hosts:             host,
		Port:              c.Int(FlagDBPort),
		Region:            c.String(FlagDBRegion),
		User:              c.String(FlagUsername),
		Password:          c.String(FlagPassword),
		Keyspace:          getRequiredOption(c, FlagKeyspace),
		ProtoVersion:      cassandraProtoVersion,
		SerialConsistency: gocql.LocalSerial,
		MaxConns:          20,
		Consistency:       gocql.LocalQuorum,
		Timeout:           10 * time.Second,
	}
	if c.Bool(FlagEnableTLS) {
		clusterConfig.TLS = &config.TLS{
			Enabled:                true,
			CertFile:               c.String(FlagTLSCertPath),
			KeyFile:                c.String(FlagTLSKeyPath),
			CaFile:                 c.String(FlagTLSCaPath),
			EnableHostVerification: c.Bool(FlagTLSEnableHostVerification),
		}
	}

	client := gocql.NewClient()
	session, err := client.CreateSession(clusterConfig)
	if err != nil {
		ErrorAndExit("connect to Cassandra failed", err)
	}
	return client, session
}

func connectToSQL(c *cli.Context) sqlplugin.DB {
	host := getRequiredOption(c, FlagDBAddress)
	if !c.IsSet(FlagDBPort) {
		ErrorAndExit("sql port is required", nil)
	}
	encodingType := c.String(FlagEncodingType)
	decodingTypesStr := c.StringSlice(FlagDecodingTypes)

	sqlConfig := &config.SQL{
		ConnectAddr: net.JoinHostPort(
			host,
			c.String(FlagDBPort),
		),
		PluginName:    c.String(FlagDBType),
		User:          c.String(FlagUsername),
		Password:      c.String(FlagPassword),
		DatabaseName:  getRequiredOption(c, FlagDatabaseName),
		EncodingType:  encodingType,
		DecodingTypes: decodingTypesStr,
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

	_, session := connectToCassandra(c)

	if len(domainID) > 0 {
		tmpl := "select domain from domains where id = ? "
		query := session.Query(tmpl, domainID)
		res, err := readOneRow(query)
		if err != nil {
			ErrorAndExit("readOneRow", err)
		}
		domain := res["domain"].(map[string]interface{})
		domainName := domain["name"].(string)
		fmt.Printf("domainName for domainID %v is %v \n", domainID, domainName)
	} else {
		tmpl := "select domain from domains_by_name where name = ?"
		tmplV2 := "select domain from domains_by_name_v2 where domains_partition=0 and name = ?"

		query := session.Query(tmpl, domainName)
		res, err := readOneRow(query)
		if err != nil {
			fmt.Printf("v1 return error: %v , trying v2...\n", err)

			query := session.Query(tmplV2, domainName)
			res, err := readOneRow(query)
			if err != nil {
				ErrorAndExit("readOneRow for v2", err)
			}
			domain := res["domain"].(map[string]interface{})
			domainID := domain["id"].(gocql.UUID).String()
			fmt.Printf("domainID for domainName %v is %v \n", domainName, domainID)
		} else {
			domain := res["domain"].(map[string]interface{})
			domainID := domain["id"].(gocql.UUID).String()
			fmt.Printf("domainID for domainName %v is %v \n", domainName, domainID)
		}
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

	ctx, cancel := newContext(c)
	defer cancel()

	req := &types.RemoveTaskRequest{
		ShardID:             int32(shardID),
		Type:                common.Int32Ptr(int32(typeID)),
		TaskID:              taskID,
		VisibilityTimestamp: common.Int64Ptr(visibilityTimestamp),
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
	client, session := connectToCassandra(c)
	shardStore := cassp.NewShardPersistenceFromSession(client, session, "current-cluster", loggerimpl.NewNopLogger())
	shardManager := persistence.NewShardManager(shardStore)

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
	client, session := connectToCassandra(c)
	shardStore := cassp.NewShardPersistenceFromSession(client, session, "current-cluster", loggerimpl.NewNopLogger())
	shardManager := persistence.NewShardManager(shardStore)

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
