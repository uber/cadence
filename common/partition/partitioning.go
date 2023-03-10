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

package partition

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination partitioning_mock.go -self_package github.com/uber/cadence/common/partition

import (
	"context"
	"github.com/uber/cadence/common/dynamicconfig"

	"github.com/uber/cadence/common/types"
)

// Config is the base configuration for the partitioning library
type Config struct {
	zonalPartitioningEnabledGlobally  dynamicconfig.BoolPropertyFnWithDomainIDFilter
	zonalPartitioningEnabledForDomain dynamicconfig.BoolPropertyFnWithDomainFilter
	allIsolationGroups                []types.IsolationGroupName
}

// State is a convenience return type of a collection of IsolationGroup configurations
type State struct {
	Global types.IsolationGroupConfiguration
	Domain types.IsolationGroupConfiguration
}

type Partitioner interface {
	// GetTaskIsolationGroupByDomainID gets where the task workflow should be executing. Largely used by Matching
	// when determining which isolationGroup to place the tasks in.
	// Implementations ought to return (nil, nil) for when the feature is not enabled.
	GetTaskIsolationGroupByDomainID(ctx context.Context, DomainID string, partitionKey types.PartitionConfig) (*types.IsolationGroupName, error)
	// IsDrained answers the question - "is this particular isolationGroup drained?". Used by startWorkflow calls
	// and similar sync frontend calls to make routing decisions
	IsDrained(ctx context.Context, Domain string, IsolationGroup types.IsolationGroupName) (bool, error)
	IsDrainedByDomainID(ctx context.Context, DomainID string, IsolationGroup types.IsolationGroupName) (bool, error)
}

// IsolationGroupState is a heavily cached in-memory library for returning the state of what zones are healthy or
// drained presently. It may return an inclusive (allow-list based) or a exclusive (deny-list based) set of IsolationGroups
// depending on the implementation.
type IsolationGroupState interface {
	Get(ctx context.Context, domain string) (*State, error)
	GetByDomainID(ctx context.Context, domainID string) (*State, error)
}
