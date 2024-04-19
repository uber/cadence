// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nosql

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testQueueType = persistence.DomainReplicationQueueType

type queueStoreTestData struct {
	mockDB *nosqlplugin.MockDB
}

func newQueueStoreTestData(t *testing.T) *queueStoreTestData {
	var testData queueStoreTestData
	ctrl := gomock.NewController(t)

	testData.mockDB = nosqlplugin.NewMockDB(ctrl)

	mockPlugin := nosqlplugin.NewMockPlugin(ctrl)
	mockPlugin.EXPECT().CreateDB(gomock.Any(), gomock.Any(), gomock.Any()).Return(testData.mockDB, nil).AnyTimes()
	RegisterPluginForTest(t, "cassandra", mockPlugin)
	return &testData
}

func (td *queueStoreTestData) newQueueStore() (persistence.Queue, error) {
	cfg := getValidShardedNoSQLConfig()
	return newNoSQLQueueStore(cfg, log.NewNoop(), testQueueType, nil)
}

func (td *queueStoreTestData) mockErrCheck(err error) {
	td.mockDB.EXPECT().IsNotFoundError(err).Return(false)
	td.mockDB.EXPECT().IsTimeoutError(err).Return(false)
	td.mockDB.EXPECT().IsDBUnavailableError(err).Return(false)
	td.mockDB.EXPECT().IsThrottlingError(err).Return(false)
}

func TestNewNoSQLQueueStore_NoErrors(t *testing.T) {
	const initialQueueVersion = int64(0)
	const dlqQueueType = -testQueueType

	td := newQueueStoreTestData(t)
	// return no queue metadata which should force to create a new one
	mainQueueCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), testQueueType).Return(nil, nil)
	mainMetadataInsert := td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), testQueueType, initialQueueVersion).
		Return(nil)

	// now the corresponding DLQ metadata should be created
	dlqCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), dlqQueueType).Return(nil, nil)
	dlqMetadataInsert := td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), dlqQueueType, initialQueueVersion).
		Return(nil)

	gomock.InOrder(
		mainQueueCheckExists,
		mainMetadataInsert,
		dlqCheckExists,
		dlqMetadataInsert,
	)

	store, err := td.newQueueStore()
	require.NoError(t, err)
	assert.NotNil(t, store)
}

func TestNewNoSQLQueueStore_FailsIfCantReadMetadata(t *testing.T) {
	selectErr := errors.New("select main-queue metadata failed")
	td := newQueueStoreTestData(t)

	td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), testQueueType).Return(nil, selectErr)
	td.mockDB.EXPECT().IsNotFoundError(selectErr).Return(false)

	td.mockErrCheck(selectErr)

	store, err := td.newQueueStore()
	assert.ErrorContains(t, err, selectErr.Error())
	assert.Nil(t, store)
}

func TestNewNoSQLQueueStore_FailsIfCantInsertMetadata(t *testing.T) {
	insertErr := errors.New("insert main-queue metadata failed")
	td := newQueueStoreTestData(t)

	td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), testQueueType).Return(nil, nil)
	td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), testQueueType, gomock.Any()).Return(insertErr)
	td.mockErrCheck(insertErr)

	store, err := td.newQueueStore()
	assert.ErrorContains(t, err, insertErr.Error())
	assert.Nil(t, store)
}

func TestNewNoSQLQueueStore_FailsIfCantReadDLQMetadata(t *testing.T) {
	const dlqQueueType = -testQueueType
	errSelect := errors.New("select dlq metadata failed")

	td := newQueueStoreTestData(t)
	mainQueueCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), testQueueType).Return(nil, nil)
	mainMetadataInsert := td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), testQueueType, gomock.Any()).
		Return(nil)

	dlqCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), dlqQueueType).Return(nil, errSelect)
	td.mockDB.EXPECT().IsNotFoundError(errSelect).Return(false)

	gomock.InOrder(
		mainQueueCheckExists,
		mainMetadataInsert,
		dlqCheckExists,
	)

	td.mockErrCheck(errSelect)

	store, err := td.newQueueStore()
	assert.ErrorContains(t, err, errSelect.Error())
	assert.Nil(t, store)
}

func TestNewNoSQLQueueStore_FailsIfCantInsertDLQMetadata(t *testing.T) {
	const dlqQueueType = -testQueueType
	errInsert := errors.New("insert dlq metadata failed")

	td := newQueueStoreTestData(t)
	mainQueueCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), testQueueType).Return(nil, nil)
	mainMetadataInsert := td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), testQueueType, gomock.Any()).
		Return(nil)

	dlqCheckExists := td.mockDB.EXPECT().SelectQueueMetadata(gomock.Any(), dlqQueueType).Return(nil, nil)
	dlqMetadataInsert := td.mockDB.EXPECT().InsertQueueMetadata(gomock.Any(), dlqQueueType, gomock.Any()).Return(errInsert)

	gomock.InOrder(
		mainQueueCheckExists,
		mainMetadataInsert,
		dlqCheckExists,
		dlqMetadataInsert,
	)

	td.mockErrCheck(errInsert)

	store, err := td.newQueueStore()
	assert.ErrorContains(t, err, errInsert.Error())
	assert.Nil(t, store)
}

func TestGetNextID(t *testing.T) {
	tests := map[string]struct {
		acks     map[string]int64
		lastID   int64
		expected int64
	}{
		"expected case - last ID is equal to ack-levels": {
			acks:     map[string]int64{"a": 3},
			lastID:   3,
			expected: 4,
		},
		"expected case - last ID is equal to ack-levels haven't caught up": {
			acks:     map[string]int64{"a": 2},
			lastID:   3,
			expected: 4,
		},
		"error case - ack-levels are ahead for some reason": {
			acks:     map[string]int64{"a": 3},
			lastID:   2,
			expected: 4,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, getNextID(td.acks, td.lastID))
		})
	}
}
