package shard

import (
	"fmt"
	"github.com/pborman/uuid"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"github.com/uber/cadence/service/worker/scanner/executions/invariants"
)

type (
	scanner struct {
		shardID int
		itr common.ExecutionIterator
		failedWriter common.ExecutionWriter
		corruptedWriter common.ExecutionWriter
		invariantManager common.InvariantManager
	}
)

// NewScanner constructs a new scanner
func NewScanner(
	shardID int,
	pr common.PersistenceRetryer,
	persistencePageSize int,
	blobstoreClient blobstore.Client,
	blobstoreFlushThreshold int,
	invariantCollections []common.InvariantCollection,
) common.Scanner {
	id := uuid.New()
	return &scanner{
		shardID: shardID,
		itr: common.NewPersistenceIterator(pr, persistencePageSize, shardID),
		failedWriter: common.NewBlobstoreWriter(id, common.FailedExtension, blobstoreClient, blobstoreFlushThreshold),
		corruptedWriter: common.NewBlobstoreWriter(id, common.CorruptedExtension, blobstoreClient, blobstoreFlushThreshold),
		invariantManager: invariants.NewInvariantManager(invariantCollections, pr),
	}
}

// Scan scans over all executions in shard and runs invariant checks per execution.
func (s *scanner) Scan() common.ShardScanReport {
	result := common.ShardScanReport{
		ShardID: s.shardID,
		Stats: common.ShardScanStats{
			CorruptionByType: make(map[common.InvariantType]int64),
		},
	}

	for s.itr.HasNext() {
		exec, err := s.itr.Next()
		if err != nil {
			result.Result.ControlFlowFailure = &common.ControlFlowFailure{
				Info: "persistence iterator returned error",
				InfoDetails: err.Error(),
			}
			return result
		}
		checkResult := s.invariantManager.RunChecks(*exec)
		result.Stats.ExecutionsCount++
		switch checkResult.CheckResultType {
		case common.CheckResultTypeHealthy:
			// do nothing if execution is healthy
		case common.CheckResultTypeCorrupted:
			if err := s.corruptedWriter.Add(common.ScanOutputEntity{
				Execution: *exec,
				Result: checkResult,
			}); err != nil {
				result.Result.ControlFlowFailure = &common.ControlFlowFailure{
					Info: "blobstore add failed for corrupted execution check",
					InfoDetails: err.Error(),
				}
				return result
			}
			result.Stats.CorruptedCount++
			lastInvariant := checkResult.CheckResults[len(checkResult.CheckResults) - 1]
			result.Stats.CorruptionByType[lastInvariant.InvariantType]++
			if common.Open(exec.State) {
				result.Stats.CorruptedOpenExecutionCount++
			}
		case common.CheckResultTypeFailed:
			if err := s.failedWriter.Add(common.ScanOutputEntity{
				Execution: *exec,
				Result: checkResult,
			}); err != nil {
				result.Result.ControlFlowFailure = &common.ControlFlowFailure{
					Info: "blobstore add failed for failed execution check",
					InfoDetails: err.Error(),
				}
				return result
			}
			result.Stats.CheckFailedCount++
		default:
			panic(fmt.Sprintf("unknown CheckResultType: %v", checkResult.CheckResultType))
		}
	}

	if err := s.failedWriter.Flush(); err != nil {
		result.Result.ControlFlowFailure = &common.ControlFlowFailure{
			Info: "failed to flush for failed execution checks",
			InfoDetails: err.Error(),
		}
		return result
	}
	if err := s.corruptedWriter.Flush(); err != nil {
		result.Result.ControlFlowFailure = &common.ControlFlowFailure{
			Info: "failed to flush for corrupted execution checks",
			InfoDetails: err.Error(),
		}
		return result
	}

	result.Result.ShardScanKeys = &common.ShardScanKeys{
		Corrupt: s.corruptedWriter.FlushedKeys(),
		Failed: s.failedWriter.FlushedKeys(),
	}
	return result
}
