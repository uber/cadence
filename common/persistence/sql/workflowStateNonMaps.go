package sql
import(
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/hmgle/sqlx"
	"fmt"
)


const (
	deleteSignalsRequestedSetSQLQuery = `DELETE FROM signals_requested_sets
WHERE
shard_id = :shard_id AND
domain_id = :domain_id AND
workflow_id = :workflow_id AND
run_id = :run_id
`

	addToSignalsRequestedSetSQLQuery = `INSERT IGNORE INTO signals_requested_sets
(shard_id, domain_id, workflow_id, run_id, signal_id) VALUES
(:shard_id, :domain_id, :workflow_id, :run_id, :signal_id)`

	removeFromSignalsRequestedSetSQLQuery = `DELETE FROM signals_requested_sets
WHERE 
shard_id = :shard_id AND
domain_id = :domain_id AND
workflow_id = :workflow_id AND
run_id = :run_id AND
signal_id = :signal_id`

	getSignalsRequestedSetSQLQuery = `SELECT signal_id FROM signals_requested_sets WHERE
shard_id = ? AND
domain_id = ? AND
workflow_id = ? AND
run_id = ?`
)

type (
	signalsRequestedSetsRow struct {
		ShardID int64 `db:"shard_id"`
		DomainID string `db:"domain_id"`
		WorkflowID string `db:"workflow_id"`
		RunID string `db:"run_id"`
		SignalID string `db:"signal_id"`
	}
)

func updateSignalsRequested(tx *sqlx.Tx,
	signalRequestedIDs []string,
		deleteSignalRequestID string,
	shardID int,
			domainID, workflowID, runID string) error {
				if len(signalRequestedIDs) > 0 {
					signalsRequestedSetsRows := make([]signalsRequestedSetsRow, len(signalRequestedIDs))
					for i, v := range signalRequestedIDs {
						signalsRequestedSetsRows[i] = signalsRequestedSetsRow{
							ShardID: int64(shardID),
							DomainID: domainID,
							WorkflowID: workflowID,
							RunID: runID,
							SignalID: v,
						}
					}

					query, args, err := tx.BindNamed(addToSignalsRequestedSetSQLQuery, signalsRequestedSetsRows)
					if err != nil {
						return &workflow.InternalServiceError{
							Message: fmt.Sprintf("Failed to update signals requested. Failed to bind query. Error: %v", err),
						}
					}

					if _, err := tx.Exec(query, args...); err != nil {
						return &workflow.InternalServiceError{
							Message: fmt.Sprintf("Failed to update signals requested. Failed to execute update query. Error: %v", err),
						}
					}
				}

				if deleteSignalRequestID != "" {
					if _, err := tx.NamedExec(removeFromSignalsRequestedSetSQLQuery, &signalsRequestedSetsRow{
						ShardID: int64(shardID),
						DomainID: domainID,
						WorkflowID: workflowID,
						RunID: runID,
						SignalID: deleteSignalRequestID,
					}); err != nil {
						return &workflow.InternalServiceError{
							Message: fmt.Sprintf("Failed to update signals requested. Failed to execute delete query. Error: %v", err),
						}
					}
				}

			return nil
}

func getSignalsRequested(tx *sqlx.Tx,
	shardID int,
	domainID,
	workflowID,
	runID string) (map[string]struct{}, error) {
		var signals []string
		if err := tx.Select(&signals, getSignalsRequestedSetSQLQuery,
			shardID,
			domainID,
			workflowID,
			runID); err != nil {
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("Failed to get signals requested. Error: %v", err),
			}
		}

		var ret = make(map[string]struct{})
		for _, s := range signals {
			ret[s] = struct{}{}
		}
		return ret, nil
}

func deleteSignalsRequestedSet(tx *sqlx.Tx, shardID int, domainID, workflowID, runID string) error {
	if _, err := tx.NamedExec(deleteSignalsRequestedSetSQLQuery, &signalsRequestedSetsRow{
		ShardID: int64(shardID),
		DomainID: domainID,
		WorkflowID: workflowID,
		RunID: runID,
	}); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to delete signals requested set. Error: %v", err),
		}
	}
	return nil
}