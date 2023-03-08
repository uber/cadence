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

type Partitioner interface {
	// GetTaskZoneByDomainID gets where the task workflow should be executing. Largely used by Matching
	// when determining which zone to place the tasks in
	GetTaskZoneByDomainID(ctx context.Context, DomainID string, partitionKey types.PartitionConfig) (*types.ZoneName, error)
	// IsDrained answers the question - "is this particular zone drained?". Used by startWorkflow calls
	// and similar sync frontend calls to make routing decisions
	IsDrained(ctx context.Context, Domain string, Zone types.ZoneName) (bool, error)
	IsDrainedByDomainID(ctx context.Context, DomainID string, Zone types.ZoneName) (bool, error)
}

type ZoneState interface {
	// ListAll lists the status of all zones with respect to the particular domainID
	ListAll(ctx context.Context, domainID string) ([]types.ZonePartition, error)
	// Get returns the state of a particular zone from the point of view of the domain
	// responsible for merging both global configuration and zone specific configuration, with domain
	// configuration overriding the global configuration
	Get(ctx context.Context, domain string, zone types.ZoneName) (*types.ZonePartition, error)
	GetByDomainID(ctx context.Context, domainID string, zone types.ZoneName) (*types.ZonePartition, error)
}

// Config is the base configuration for the partitioning library
type Config struct {
	zonalPartitioningEnabled dynamicconfig.BoolPropertyFnWithDomainFilter
	allZones                 []types.ZoneName
}
