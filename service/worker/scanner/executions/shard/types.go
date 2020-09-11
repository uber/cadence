// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package shard

import (
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/reconciliation/types"
)

type (
	// ScanType is the enum for representing different entity types to scan
	ScanType int
)

const (
	// ConcreteExecutionType concrete execution entity
	ConcreteExecutionType ScanType = iota
	// CurrentExecutionType current execution entity
	CurrentExecutionType
)

// ToBlobstoreEntity picks struct depending on Scanner type
func (st ScanType) ToBlobstoreEntity() types.BlobstoreEntity {
	switch st {
	case ConcreteExecutionType:
		return &types.ConcreteExecution{}
	case CurrentExecutionType:
		return &types.CurrentExecution{}
	}
	panic("unknown scan type")
}

// ToInvariants returns list of invariants to be checked
func (st ScanType) ToInvariants(collections []common.InvariantCollection) []func(retryer persistence.Retryer) invariant.Invariant {
	var fns []func(retryer persistence.Retryer) invariant.Invariant
	switch st {
	case ConcreteExecutionType:
		for _, collection := range collections {
			switch collection {
			case common.InvariantCollectionHistory:
				fns = append(fns, invariant.NewHistoryExists)
			case common.InvariantCollectionMutableState:
				fns = append(fns, invariant.NewOpenCurrentExecution)
			}
		}
		return fns
	case CurrentExecutionType:
		for _, collection := range collections {
			switch collection {
			case common.InvariantCollectionMutableState:
				fns = append(fns, invariant.NewConcreteExecutionExists)
			}
		}
		return fns
	default:
		panic("unknown scan type")
	}
}

// ToScanner returns function to be used
func (st ScanType) ToScanner() func(ScannerParams) common.Scanner {
	switch st {
	case ConcreteExecutionType:
		return NewConcreteExecutionScanner
	case CurrentExecutionType:
		return NewCurrentExecutionScanner
	default:
		panic("unknown scan type")
	}
}

// ScannerParams holds list of arguments used when creating a Scanner
type ScannerParams struct {
	Retryer                 persistence.Retryer
	PersistencePageSize     int
	BlobstoreClient         blobstore.Client
	BlobstoreFlushThreshold int
	Invariants              []invariant.Invariant
	ProgressReportFn        func()
}
