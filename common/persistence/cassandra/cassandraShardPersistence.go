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

package cassandra

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
)

type (
	// Implements ShardManager
	nosqlShardManager struct {
		db                 nosqlplugin.DB
		logger             log.Logger
		currentClusterName string
	}
)

var _ p.ShardStore = (*nosqlShardManager)(nil)

func (sh *nosqlShardManager) GetName() string {
	return sh.db.PluginName()
}

// Close releases the underlying resources held by this object
func (sh *nosqlShardManager) Close() {
	sh.db.Close()
}

// newShardPersistence is used to create an instance of ShardManager implementation
func newShardPersistence(
	cfg config.Cassandra,
	clusterName string,
	logger log.Logger,
) (p.ShardStore, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlShardManager{
		db:                 db,
		logger:             logger,
		currentClusterName: clusterName,
	}, nil
}

// NewShardPersistenceFromSession is used to create an instance of ShardManager implementation
// It is being used by some admin toolings
func NewShardPersistenceFromSession(
	client gocql.Client,
	session gocql.Session,
	clusterName string,
	logger log.Logger,
) p.ShardStore {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db := cassandra.NewCassandraDBFromSession(client, session, logger)

	return &nosqlShardManager{
		db:                 db,
		logger:             logger,
		currentClusterName: clusterName,
	}
}

func (sh *nosqlShardManager) CreateShard(
	ctx context.Context,
	request *p.InternalCreateShardRequest,
) error {

	previous, err := sh.db.InsertShard(ctx, request.ShardInfo)
	if err != nil {
		if sh.db.IsConditionFailedError(err) {
			return &p.ShardAlreadyExistError{
				Msg: fmt.Sprintf("Shard already exists in executions table.  ShardId: %v, RangeId: %v",
					previous.ShardID, previous.PreviousRangeID),
			}
		}
		return convertCommonErrors(sh.db, "CreateShard", err)
	}

	return nil
}

func (sh *nosqlShardManager) GetShard(
	ctx context.Context,
	request *p.InternalGetShardRequest,
) (*p.InternalGetShardResponse, error) {
	shardID := request.ShardID
	rangeID, shardInfo, err := sh.db.SelectShard(ctx, shardID, sh.currentClusterName)

	if err != nil {
		if sh.db.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Shard not found.  ShardId: %v", shardID),
			}
		}

		return nil, convertCommonErrors(sh.db, "GetShard", err)
	}

	shardInfoRangeID := shardInfo.RangeID

	// check if rangeID column and rangeID field in shard column matches, if not we need to pick the larger
	// rangeID.
	if shardInfoRangeID > rangeID {
		// In this case we need to fix the rangeID column before returning the result as:
		// 1. if we return shardInfoRangeID, then later shard CAS operation will fail
		// 2. if we still return rangeID, CAS will work but rangeID will move backward which
		// result in lost tasks, corrupted workflow history, etc.

		sh.logger.Warn("Corrupted shard rangeID", tag.ShardID(shardID), tag.ShardRangeID(shardInfoRangeID), tag.PreviousShardRangeID(rangeID))
		if err := sh.updateRangeID(ctx, shardID, shardInfoRangeID, rangeID); err != nil {
			return nil, err
		}
	} else {
		// return the actual rangeID
		shardInfo.RangeID = rangeID
		//
		// If shardInfoRangeID = rangeID, no corruption, so no action needed.
		//
		// If shardInfoRangeID < rangeID, we also don't need to do anything here as createShardInfo will ignore
		// shardInfoRangeID and return rangeID instead. Later when updating the shard, CAS can still succeed
		// as the value from rangeID columns is returned, shardInfoRangeID will also be updated to the correct value.
	}

	return &p.InternalGetShardResponse{ShardInfo: shardInfo}, nil
}

func (sh *nosqlShardManager) updateRangeID(
	ctx context.Context,
	shardID int,
	rangeID int64,
	previousRangeID int64,
) error {

	previous, err := sh.db.UpdateRangeID(ctx, shardID, rangeID, previousRangeID)
	if err != nil {
		if sh.db.IsConditionFailedError(err) {
			return &p.ShardOwnershipLostError{
				ShardID: shardID,
				Msg: fmt.Sprintf("Failed to update shard rangeID.  previous_range_id: %v, columns: (%v)",
					previousRangeID, previous.Details),
			}
		}
		return convertCommonErrors(sh.db, "UpdateRangeID", err)
	}

	return nil
}

func (sh *nosqlShardManager) UpdateShard(
	ctx context.Context,
	request *p.InternalUpdateShardRequest,
) error {
	previous, err := sh.db.UpdateShard(ctx, request.ShardInfo, request.PreviousRangeID)
	if err != nil {
		if sh.db.IsConditionFailedError(err) {
			return &p.ShardOwnershipLostError{
				ShardID: request.ShardInfo.ShardID,
				Msg: fmt.Sprintf("Failed to update shard.  previous_range_id: %v, columns: (%v)",
					request.PreviousRangeID, previous.Details),
			}
		}
		return convertCommonErrors(sh.db, "UpdateShard", err)
	}

	return nil
}
