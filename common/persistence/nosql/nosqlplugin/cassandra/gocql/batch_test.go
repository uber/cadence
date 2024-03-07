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

package gocql

import (
	"context"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
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
