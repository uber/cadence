package persistence

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/uber-common/bark"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const (
	// Fixed domain values for now
	domainID        = "73278331-596f-4033-9ce2-ac6451989400"
	domainPartition = 0
)

const (
	templateCreateWorkflowExecutionStarted = `INSERT INTO open_executions (` +
		`domain_id, domain_partition, workflow_id, run_id, start_time, workflow_type_name) ` +
		`VALUES (?, ?, ?, ?, ?, ?)`

	templateDeleteWorkflowExecutionStarted = `DELETE FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? `

	templateCreateWorkflowExecutionClosed = `INSERT INTO closed_executions (` +
		`domain_id, domain_partition, workflow_id, run_id, start_time, close_time, workflow_type_name) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?) using TTL ?`
)

type (
	cassandraVisibilityPersistence struct {
		session      *gocql.Session
		lowConslevel gocql.Consistency
		logger       bark.Logger
	}
)

// NewCassandraVisibilityPersistence is used to create an instance of VisibilityManager implementation
func NewCassandraVisibilityPersistence(
	hosts string, dc string, keyspace string, logger bark.Logger) (VisibilityManager, error) {
	cluster := common.NewCassandraCluster(hosts, dc)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = cassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = defaultSessionTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &cassandraVisibilityPersistence{session: session, lowConslevel: gocql.One, logger: logger}, nil
}

func (v *cassandraVisibilityPersistence) RecordWorkflowExecutionStarted(
	request *RecordWorkflowExecutionStartedRequest) error {
	query := v.session.Query(templateCreateWorkflowExecutionStarted,
		domainID,
		domainPartition,
		request.Execution.GetWorkflowId(),
		request.Execution.GetRunId(),
		request.StartTimestamp,
		request.WorkflowTypeName,
	)
	err := query.Exec()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("RecordWorkflowExecutionStarted operation failed. Error: %v", err),
		}
	}

	return nil
}

func (v *cassandraVisibilityPersistence) RecordWorkflowExecutionClosed(
	request *RecordWorkflowExecutionClosedRequest) error {
	// First, remove execution from the open table
	query := v.session.Query(templateDeleteWorkflowExecutionStarted,
		domainID,
		domainPartition,
		request.Execution.GetWorkflowId(),
		request.Execution.GetRunId(),
	)
	err := query.Exec()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("RecordWorkflowExecutionClosed operation failed. Error: %v", err),
		}
	}

	// Next, add a row in the closed table. This row is kepy for defaultDeleteTTLSeconds
	query = v.session.Query(templateCreateWorkflowExecutionClosed,
		domainID,
		domainPartition,
		request.Execution.GetWorkflowId(),
		request.Execution.GetRunId(),
		request.StartTimestamp,
		request.CloseTimestamp,
		request.WorkflowTypeName,
		defaultDeleteTTLSeconds,
	)
	err = query.Exec()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("RecordWorkflowExecutionClosed operation failed. Error: %v", err),
		}
	}
	return nil
}
