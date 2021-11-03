// Copyright (c) 2018 Uber Technologies, Inc.
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

package sql

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/uber/cadence/common/persistence/serialization"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
)

// TODO: Rename all SQL Managers to Stores
type sqlStore struct {
	db     sqlplugin.DB
	logger log.Logger
	parser serialization.Parser
}

func (m *sqlStore) GetName() string {
	return m.db.PluginName()
}

func (m *sqlStore) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *sqlStore) txExecute(ctx context.Context, dbShardID int, operation string, f func(tx sqlplugin.Tx) error) error {
	tx, err := m.db.BeginTx(dbShardID, ctx)
	if err != nil {
		return convertCommonErrors(m.db, operation, "Failed to start transaction.", err)
	}
	err = f(tx)
	if err != nil {
		rollBackErr := tx.Rollback()
		if rollBackErr != nil {
			m.logger.Error("transaction rollback error", tag.Error(rollBackErr))
		}
		return convertCommonErrors(m.db, operation, "", err)
	}
	if err := tx.Commit(); err != nil {
		return convertCommonErrors(m.db, operation, "Failed to commit transaction.", err)
	}
	return nil
}

func gobSerialize(x interface{}) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(x)
	if err != nil {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("Error in serialization: %v", err),
		}
	}
	return b.Bytes(), nil
}

func gobDeserialize(a []byte, x interface{}) error {
	b := bytes.NewBuffer(a)
	d := gob.NewDecoder(b)
	err := d.Decode(x)

	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Error in deserialization: %v", err),
		}
	}
	return nil
}

func serializePageToken(offset int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(offset))
	return b
}

func deserializePageToken(payload []byte) (int64, error) {
	if len(payload) != 8 {
		return 0, fmt.Errorf("Invalid token of %v length", len(payload))
	}
	return int64(binary.LittleEndian.Uint64(payload)), nil
}

func convertCommonErrors(
	errChecker sqlplugin.ErrorChecker,
	operation, message string,
	err error,
) error {
	switch err.(type) {
	case *persistence.ConditionFailedError,
		*persistence.CurrentWorkflowConditionFailedError,
		*persistence.WorkflowExecutionAlreadyStartedError,
		*persistence.ShardOwnershipLostError,
		*persistence.TimeoutError,
		*types.DomainAlreadyExistsError,
		*types.EntityNotExistsError,
		*types.ServiceBusyError,
		*types.InternalServiceError:
		return err
	}
	if errChecker.IsNotFoundError(err) {
		return &types.EntityNotExistsError{
			Message: fmt.Sprintf("%v failed. %s Error: %v ", operation, message, err),
		}
	}

	if errChecker.IsTimeoutError(err) {
		return &persistence.TimeoutError{Msg: fmt.Sprintf("%v timed out. %s Error: %v", operation, message, err)}
	}

	if errChecker.IsThrottlingError(err) {
		return &types.ServiceBusyError{
			Message: fmt.Sprintf("%v operation failed. %s Error: %v", operation, message, err),
		}
	}

	return &types.InternalServiceError{
		Message: fmt.Sprintf("%v operation failed. %s Error: %v", operation, message, err),
	}
}
