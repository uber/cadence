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

package types

const (
	ZoneDrainStatusInvalid ZoneStatus = iota
	ZoneDrainStatusHealthy
	ZoneDrainStatusDrained
)

// A ZoneName is a subdivision of a 'region', such as a subset of racks in a datacentre or a division of
// traffic which there is a need for logical separation for resilience, but these subdivisions still operate within
// the databases' ability to operate consistently.
type ZoneName string
type ZoneStatus int

// PartitionConfig is a key/value based set of configuration for partitioning traffic. Intended to be a key/value pair
// of data encoded in JSON or whatever encoding suits. This is intentionally opaque and to be passed blindly
// to the partitioner of choice as it may contain business-specific types.
//
// Example of the intent:
// partitionCfg := []byte(`{"wf-start-zone": "zone123", "userid: "1234", "weighting": 0.5}`)
// which, for example, may allow the partitioner to choose to split traffic based on where the workflow started, or
// the user, or any arbitrary other configuration
type PartitionConfig string

type ZonePartition struct {
	Name   ZoneName
	Status ZoneStatus
}
