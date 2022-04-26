// Copyright (c) 2020 Uber Technologies, Inc.
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
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	defaultPageSize = 1000
)

type DLQRow struct {
	ShardID         int                        `header:"Shard ID" json:"shardID"`
	DomainName      string                     `header:"Domain Name" json:"domainName"`
	DomainID        string                     `header:"Domain ID" json:"domainID"`
	WorkflowID      string                     `header:"Workflow ID" json:"workflowID"`
	RunID           string                     `header:"Run ID" json:"runID"`
	TaskID          int64                      `header:"Task ID" json:"taskID"`
	TaskType        *types.ReplicationTaskType `header:"Task Type" json:"taskType"`
	Version         int64                      `json:"version"`
	FirstEventID    int64                      `json:"firstEventID"`
	NextEventID     int64                      `json:"nextEventID"`
	ScheduledID     int64                      `json:"scheduledID"`
	ReplicationTask *types.ReplicationTask     `json:"replicationTask"`

	// Those are deserialized variants from history replications task
	Events       []*types.HistoryEvent `json:"events"`
	NewRunEvents []*types.HistoryEvent `json:"newRunEvents,omitempty"`

	// Only event IDs for compact table representation
	EventIDs       []int64 `header:"Event IDs"`
	NewRunEventIDs []int64 `header:"New Run Event IDs"`
}

type HistoryDLQCountRow struct {
	SourceCluster string `header:"Source Cluster" json:"sourceCluster"`
	ShardID       int32  `header:"Shard ID" json:"shardID"`
	Count         int64  `header:"Count" json:"count"`
}

// AdminCountDLQMessages returns info how many and where DLQ messages are queued
func AdminCountDLQMessages(c *cli.Context) {
	force := c.Bool(FlagForce)
	ctx, cancel := newContext(c)
	defer cancel()

	adminClient := cFactory.ServerAdminClient(c)
	response, err := adminClient.CountDLQMessages(ctx, &types.CountDLQMessagesRequest{ForceFetch: force})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while getting DLQ count, results may be partial: %v\n", err)
	}

	if c.String(FlagDLQType) == "domain" {
		fmt.Println(response.Domain)
		return
	}

	table := []HistoryDLQCountRow{}
	for key, count := range response.History {
		table = append(table, HistoryDLQCountRow{
			SourceCluster: key.SourceCluster,
			ShardID:       key.ShardID,
			Count:         count,
		})
	}
	sort.Slice(table, func(i, j int) bool {
		// First sort by source cluster
		switch strings.Compare(table[i].SourceCluster, table[j].SourceCluster) {
		case -1:
			return true
		case 1:
			return false
		}

		// Then by count in decreasing order
		diff := table[i].Count - table[j].Count
		if diff > 0 {
			return true
		}
		if diff < 0 {
			return false
		}

		// Finally by shard in increasing order
		return table[i].ShardID < table[j].ShardID
	})

	Render(c, table, RenderOptions{Color: true, DefaultTemplate: templateTable})
}

// AdminGetDLQMessages gets DLQ metadata
func AdminGetDLQMessages(c *cli.Context) {
	ctx, cancel := newContext(c)
	defer cancel()

	client := cFactory.ServerFrontendClient(c)
	adminClient := cFactory.ServerAdminClient(c)

	dlqType := toQueueType(getRequiredOption(c, FlagDLQType))
	sourceCluster := getRequiredOption(c, FlagSourceCluster)

	remainingMessageCount := common.EndMessageID
	if c.IsSet(FlagMaxMessageCount) {
		remainingMessageCount = c.Int64(FlagMaxMessageCount)
	}
	lastMessageID := common.EndMessageID
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = c.Int64(FlagLastMessageID)
	}

	// Cache for domain names
	domainNames := map[string]string{}
	getDomainName := func(domainId string) string {
		if domainName, ok := domainNames[domainId]; ok {
			return domainName
		}

		resp, err := client.DescribeDomain(ctx, &types.DescribeDomainRequest{UUID: common.StringPtr(domainId)})
		if err != nil {
			ErrorAndExit("failed to describe domain", err)
		}
		domainNames[domainId] = resp.DomainInfo.Name
		return resp.DomainInfo.Name
	}

	readShard := func(shardID int) []DLQRow {
		var rows []DLQRow
		var pageToken []byte

		for {
			resp, err := adminClient.ReadDLQMessages(ctx, &types.ReadDLQMessagesRequest{
				Type:                  dlqType,
				SourceCluster:         sourceCluster,
				ShardID:               int32(shardID),
				InclusiveEndMessageID: common.Int64Ptr(lastMessageID),
				MaximumPageSize:       defaultPageSize,
				NextPageToken:         pageToken,
			})
			if err != nil {
				ErrorAndExit(fmt.Sprintf("fail to read dlq message for shard: %d", shardID), err)
			}

			replicationTasks := map[int64]*types.ReplicationTask{}
			for _, task := range resp.ReplicationTasks {
				replicationTasks[task.SourceTaskID] = task
			}

			for _, info := range resp.ReplicationTasksInfo {
				task := replicationTasks[info.TaskID]

				var taskType *types.ReplicationTaskType
				if task != nil {
					taskType = task.TaskType
				}

				events := deserializeBatchEvents(task.GetHistoryTaskV2Attributes().GetEvents())
				newRunEvents := deserializeBatchEvents(task.GetHistoryTaskV2Attributes().GetNewRunEvents())

				rows = append(rows, DLQRow{
					ShardID:         shardID,
					DomainName:      getDomainName(info.DomainID),
					DomainID:        info.DomainID,
					WorkflowID:      info.WorkflowID,
					RunID:           info.RunID,
					TaskType:        taskType,
					TaskID:          info.TaskID,
					Version:         info.Version,
					FirstEventID:    info.FirstEventID,
					NextEventID:     info.NextEventID,
					ScheduledID:     info.ScheduledID,
					ReplicationTask: task,
					Events:          events,
					EventIDs:        collectEventIDs(events),
					NewRunEvents:    newRunEvents,
					NewRunEventIDs:  collectEventIDs(newRunEvents),
				})

				remainingMessageCount--
				if remainingMessageCount <= 0 {
					return rows
				}
			}

			if len(resp.NextPageToken) == 0 {
				break
			}
			pageToken = resp.NextPageToken
		}
		return rows
	}

	table := []DLQRow{}
	for shardID := range getShards(c) {
		if remainingMessageCount <= 0 {
			break
		}
		table = append(table, readShard(shardID)...)
	}

	Render(c, table, RenderOptions{DefaultTemplate: templateTable, Color: true})
}

// AdminPurgeDLQMessages deletes messages from DLQ
func AdminPurgeDLQMessages(c *cli.Context) {
	dlqType := getRequiredOption(c, FlagDLQType)
	sourceCluster := getRequiredOption(c, FlagSourceCluster)
	var lastMessageID *int64
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = common.Int64Ptr(c.Int64(FlagLastMessageID))
	}

	adminClient := cFactory.ServerAdminClient(c)
	for shardID := range getShards(c) {
		ctx, cancel := newContext(c)
		err := adminClient.PurgeDLQMessages(ctx, &types.PurgeDLQMessagesRequest{
			Type:                  toQueueType(dlqType),
			SourceCluster:         sourceCluster,
			ShardID:               int32(shardID),
			InclusiveEndMessageID: lastMessageID,
		})
		cancel()
		if err != nil {
			fmt.Printf("Failed to purge DLQ message in shard %v with error: %v.\n", shardID, err)
			continue
		}
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("Successfully purge DLQ Messages in shard %v.\n", shardID)
	}
}

// AdminMergeDLQMessages merges message from DLQ
func AdminMergeDLQMessages(c *cli.Context) {
	dlqType := getRequiredOption(c, FlagDLQType)
	sourceCluster := getRequiredOption(c, FlagSourceCluster)
	var lastMessageID *int64
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = common.Int64Ptr(c.Int64(FlagLastMessageID))
	}

	adminClient := cFactory.ServerAdminClient(c)
ShardIDLoop:
	for shardID := range getShards(c) {
		request := &types.MergeDLQMessagesRequest{
			Type:                  toQueueType(dlqType),
			SourceCluster:         sourceCluster,
			ShardID:               int32(shardID),
			InclusiveEndMessageID: lastMessageID,
			MaximumPageSize:       defaultPageSize,
		}

		for {
			ctx, cancel := newContext(c)
			response, err := adminClient.MergeDLQMessages(ctx, request)
			cancel()
			if err != nil {
				fmt.Printf("Failed to merge DLQ message in shard %v with error: %v.\n", shardID, err)
				continue ShardIDLoop
			}

			if len(response.NextPageToken) == 0 {
				break
			}

			request.NextPageToken = response.NextPageToken
		}
		fmt.Printf("Successfully merged all messages in shard %v.\n", shardID)
	}
}

func getShards(c *cli.Context) chan int {
	// Check if we have stdin available
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return readShardsFromStdin()
	}

	return generateShardRangeFromFlags(c)
}

func generateShardRangeFromFlags(c *cli.Context) chan int {
	shards := make(chan int)
	go func() {
		shardRange, err := parseIntMultiRange(c.String(FlagShards))
		if err != nil {
			fmt.Printf("failed to parse shard range: %q\n", c.String(FlagShards))
		} else {
			for _, shard := range shardRange {
				shards <- shard
			}
		}
		close(shards)
	}()
	return shards
}

func readShardsFromStdin() chan int {
	shards := make(chan int)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("Unable to read from stdin: %v", err)
				continue
			}
			shard, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
			if err != nil {
				fmt.Printf("Failed to parse shard id: %q\n", line)
				continue
			}
			shards <- int(shard)
		}
		close(shards)
	}()
	return shards
}

func toQueueType(dlqType string) *types.DLQType {
	switch dlqType {
	case "domain":
		return types.DLQTypeDomain.Ptr()
	case "history":
		return types.DLQTypeReplication.Ptr()
	default:
		ErrorAndExit("The queue type is not supported.", fmt.Errorf("the queue type is not supported. Type: %v", dlqType))
	}
	return nil
}

func deserializeBatchEvents(blob *types.DataBlob) []*types.HistoryEvent {
	if blob == nil {
		return nil
	}
	serializer := persistence.NewPayloadSerializer()
	events, err := serializer.DeserializeBatchEvents(persistence.NewDataBlobFromInternal(blob))
	if err != nil {
		ErrorAndExit("Failed to decode DLQ history replication events", err)
	}
	return events
}

func collectEventIDs(events []*types.HistoryEvent) []int64 {
	ids := make([]int64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	return ids
}
