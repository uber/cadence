// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
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

//go:generate mockgen -copyright_file ../../../LICENSE -package $GOPACKAGE -source $GOFILE -destination mocks.go -self_package github.com/uber/cadence/common/reconciliation/common

package common

type (

	// ExecutionIterator gets Executions from underlying store.
	ExecutionIterator interface {
		// Next returns the next execution found. Any error reading from underlying store
		// or converting store entry to Execution will result in an error after which iterator cannot be used.
		Next() (interface{}, error)
		// HasNext indicates if the iterator has a next element. If HasNext is true
		// it is guaranteed that Next will return a nil error and a non-nil Execution.
		HasNext() bool
	}

	// ScanOutputIterator gets ScanOutputEntities from underlying store
	ScanOutputIterator interface {
		// Next returns the next ScanOutputEntity found. Any error reading from underlying store
		// or converting store entry to ScanOutputEntity will result in an error after which iterator cannot be used.
		Next() (*ScanOutputEntity, error)
		// HasNext indicates if the iterator has a next element. If HasNext is true it is
		// guaranteed that Next will return a nil error and non-nil ScanOutputEntity.
		HasNext() bool
	}

	// ExecutionWriter is used to write entities (FixOutputEntity or ScanOutputEntity) to blobstore
	ExecutionWriter interface {
		Add(interface{}) error
		Flush() error
		FlushedKeys() *Keys
	}

	// Scanner is used to scan over all executions in a shard. It is responsible for three things:
	// 1. Checking invariants for each execution.
	// 2. Recording corruption and failures to durable store.
	// 3. Producing a ShardScanReport
	Scanner interface {
		Scan() ShardScanReport
	}

	// Fixer is used to fix all executions in a shard. It is responsible for three things:
	// 1. Confirming that each execution it scans is corrupted.
	// 2. Attempting to fix any confirmed corrupted executions.
	// 3. Recording skipped executions, failed to fix executions and successfully fix executions to durable store.
	// 4. Producing a ShardFixReport
	Fixer interface {
		Fix() ShardFixReport
	}
)
