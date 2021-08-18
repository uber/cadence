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
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	defaultPageSize = 1000
)

// AdminGetDLQMessages gets DLQ metadata
func AdminGetDLQMessages(c *cli.Context) {
	ctx, cancel := newContext(c)
	defer cancel()

	adminClient := cFactory.ServerAdminClient(c)
	dlqType := getRequiredOption(c, FlagDLQType)
	sourceCluster := getRequiredOption(c, FlagSourceCluster)
	shardID := getRequiredIntOption(c, FlagShardID)
	serializer := persistence.NewPayloadSerializer()
	outputFile := getOutputFile(c.String(FlagOutputFilename))
	defer outputFile.Close()

	showRawTask := c.Bool(FlagDLQRawTask)
	var rawTasksInfo []*types.ReplicationTaskInfo
	remainingMessageCount := common.EndMessageID
	if c.IsSet(FlagMaxMessageCount) {
		remainingMessageCount = c.Int64(FlagMaxMessageCount)
	}
	lastMessageID := common.EndMessageID
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = c.Int64(FlagLastMessageID)
	}

	paginationFunc := func(paginationToken []byte) ([]interface{}, []byte, error) {
		resp, err := adminClient.ReadDLQMessages(ctx, &types.ReadDLQMessagesRequest{
			Type:                  toQueueType(dlqType),
			SourceCluster:         sourceCluster,
			ShardID:               int32(shardID),
			InclusiveEndMessageID: common.Int64Ptr(lastMessageID),
			MaximumPageSize:       defaultPageSize,
			NextPageToken:         paginationToken,
		})
		if err != nil {
			return nil, nil, err
		}
		var paginateItems []interface{}
		for _, item := range resp.GetReplicationTasks() {
			paginateItems = append(paginateItems, item)
		}
		if showRawTask {
			rawTasksInfo = append(rawTasksInfo, resp.GetReplicationTasksInfo()...)
		}

		return paginateItems, resp.GetNextPageToken(), err
	}

	iterator := collection.NewPagingIterator(paginationFunc)
	var lastReadMessageID int
	for iterator.HasNext() && remainingMessageCount > 0 {
		item, err := iterator.Next()
		if err != nil {
			ErrorAndExit(fmt.Sprintf("fail to read dlq message. Last read message id: %v", lastReadMessageID), err)
		}

		task := item.(*types.ReplicationTask)
		taskStr, err := decodeReplicationTask(task, serializer)
		if err != nil {
			ErrorAndExit(fmt.Sprintf("fail to encode dlq message. Last read message id: %v", lastReadMessageID), err)
		}

		lastReadMessageID = int(task.SourceTaskID)
		remainingMessageCount--
		_, err = outputFile.WriteString(fmt.Sprintf("%v\n", string(taskStr)))
		if err != nil {
			ErrorAndExit("fail to print dlq messages.", err)
		}
	}

	if showRawTask {
		_, err := outputFile.WriteString(fmt.Sprintf("#### REPLICATION DLQ RAW TASKS INFO ####\n"))
		if err != nil {
			ErrorAndExit("fail to print dlq raw tasks.", err)
		}
		for _, info := range rawTasksInfo {
			str, err := json.Marshal(info)
			if err != nil {
				ErrorAndExit("fail to encode dlq raw tasks.", err)
			}

			if _, err = outputFile.WriteString(fmt.Sprintf("%v\n", string(str))); err != nil {
				ErrorAndExit("fail to print dlq raw tasks.", err)
			}
		}
	} else {
		if lastReadMessageID == 0 && len(rawTasksInfo) > 0 {
			if _, err := outputFile.WriteString(
				fmt.Sprintf("WARN: Received empty replication task but metadata is not empty. Please use %v to show metadata task.\n", FlagDLQRawTask),
			); err != nil {
				ErrorAndExit("fail to print warning message.", err)
			}
		}
	}
}

// AdminPurgeDLQMessages deletes messages from DLQ
func AdminPurgeDLQMessages(c *cli.Context) {
	dlqType := getRequiredOption(c, FlagDLQType)
	sourceCluster := getRequiredOption(c, FlagSourceCluster)
	lowerShardBound := c.Int(FlagLowerShardBound)
	upperShardBound := c.Int(FlagUpperShardBound)
	var lastMessageID *int64
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = common.Int64Ptr(c.Int64(FlagLastMessageID))
	}

	adminClient := cFactory.ServerAdminClient(c)
	for shardID := lowerShardBound; shardID <= upperShardBound; shardID++ {
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
	lowerShardBound := c.Int(FlagLowerShardBound)
	upperShardBound := c.Int(FlagUpperShardBound)
	var lastMessageID *int64
	if c.IsSet(FlagLastMessageID) {
		lastMessageID = common.Int64Ptr(c.Int64(FlagLastMessageID))
	}

	adminClient := cFactory.ServerAdminClient(c)
ShardIDLoop:
	for shardID := lowerShardBound; shardID <= upperShardBound; shardID++ {
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
