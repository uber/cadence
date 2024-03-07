package gocql

import (
	"context"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	testBatch := gocql.Batch{}
	b := newBatch(&testBatch)
	assert.NotNil(t, b)
}

func Test_Batch_WithContext(t *testing.T) {
	testCtx := context.Background()
	testBatch := gocql.Batch{}
	b := newBatch(&testBatch)
	assert.NotNil(t, b.WithContext(testCtx))
}

func Test_Batch_WithTimestamp(t *testing.T) {
	timeNow := time.Now().Unix()
	testBatch := gocql.Batch{}
	b := newBatch(&testBatch)
	withTimeStamp := b.WithTimestamp(timeNow)
	assert.NotNil(t, withTimeStamp)
}

func Test_Batch_Consistency(t *testing.T) {
	testBatch := gocql.Batch{}
	b := newBatch(&testBatch)
	consistency := b.Consistency(One)
	assert.NotNil(t, consistency)
}

func Test_MustConvertBatchType(t *testing.T) {
	batch := []BatchType{LoggedBatch, UnloggedBatch, CounterBatch}
	for _, b := range batch {
		mustConvertBatchType(b)
	}
}

func Test_MustConvertBatchType_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	mustConvertBatchType(4)
}
