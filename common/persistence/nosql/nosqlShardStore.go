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

package nosql

import (
	"context"
	"fmt"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

type (
	// Implements ShardStore
	nosqlShardStore struct {
		nosqlStore
		currentClusterName string
	}
)

var _ p.ShardStore = (*nosqlShardStore)(nil)

// newNoSQLShardStore is used to create an instance of ShardStore implementation
func newNoSQLShardStore(
	cfg config.NoSQL,
	clusterName string,
	logger log.Logger,
) (p.ShardStore, error) {
	db, err := NewNoSQLDB(&cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlShardStore{
		nosqlStore: nosqlStore{
			db:     db,
			logger: logger,
		},
		currentClusterName: clusterName,
	}, nil
}

// NewNoSQLShardStoreFromSession is used to create an instance of ShardStore implementation
// It is being used by some admin toolings
func NewNoSQLShardStoreFromSession(
	db nosqlplugin.DB,
	clusterName string,
	logger log.Logger,
) p.ShardStore {
	return &nosqlShardStore{
		nosqlStore: nosqlStore{
			db:     db,
			logger: logger,
		},
		currentClusterName: clusterName,
	}
}

func (sh *nosqlShardStore) CreateShard(
	ctx context.Context,
	request *p.InternalCreateShardRequest,
) error {

	err := sh.db.InsertShard(ctx, request.ShardInfo)
	if err != nil {
		conditionFailure, ok := err.(*nosqlplugin.ShardOperationConditionFailure)
		if ok {
			return &p.ShardAlreadyExistError{
				Msg: fmt.Sprintf("Shard already exists in executions table.  ShardId: %v, request_range_id: %v, actual_range_id : %v, columns: (%v)",
					request.ShardInfo.ShardID, request.ShardInfo.RangeID, conditionFailure.RangeID, conditionFailure.Details),
			}
		}
		return convertCommonErrors(sh.db, "CreateShard", err)
	}

	return nil
}

func (sh *nosqlShardStore) GetShard(
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

func (sh *nosqlShardStore) updateRangeID(
	ctx context.Context,
	shardID int,
	rangeID int64,
	previousRangeID int64,
) error {

	err := sh.db.UpdateRangeID(ctx, shardID, rangeID, previousRangeID)
	if err != nil {
		conditionFailure, ok := err.(*nosqlplugin.ShardOperationConditionFailure)
		if ok {
			return &p.ShardOwnershipLostError{
				ShardID: shardID,
				Msg: fmt.Sprintf("Failed to update shard rangeID.  request_range_id: %v, actual_range_id : %v, columns: (%v)",
					previousRangeID, conditionFailure.RangeID, conditionFailure.Details),
			}
		}
		return convertCommonErrors(sh.db, "UpdateRangeID", err)
	}

	return nil
}

func (sh *nosqlShardStore) UpdateShard(
	ctx context.Context,
	request *p.InternalUpdateShardRequest,
) error {
	err := sh.db.UpdateShard(ctx, request.ShardInfo, request.PreviousRangeID)
	if err != nil {
		conditionFailure, ok := err.(*nosqlplugin.ShardOperationConditionFailure)
		if ok {
			return &p.ShardOwnershipLostError{
				ShardID: request.ShardInfo.ShardID,
				Msg: fmt.Sprintf("Failed to update shard rangeID.  request_range_id: %v, actual_range_id : %v, columns: (%v)",
					request.PreviousRangeID, conditionFailure.RangeID, conditionFailure.Details),
			}
		}
		return convertCommonErrors(sh.db, "UpdateShard", err)
	}

	return nil
}
