// Copyright (c) 2017 Uber Technologies, Inc.
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

package metrics

import (
	"strconv"
)

const (
	revisionTag     = "revision"
	branchTag       = "branch"
	buildDateTag    = "build_date"
	buildVersionTag = "build_version"
	goVersionTag    = "go_version"

	instance       = "instance"
	domain         = "domain"
	targetCluster  = "target_cluster"
	activeCluster  = "active_cluster"
	taskList       = "tasklist"
	taskListType   = "tasklistType"
	workflowType   = "workflowType"
	activityType   = "activityType"
	decisionType   = "decisionType"
	invariantType  = "invariantType"
	kafkaPartition = "kafkaPartition"
	transport      = "transport"
	signalName     = "signalName"

	domainAllValue = "all"
	unknownValue   = "_unknown_"

	transportThrift = "thrift"
	transportGRPC   = "grpc"
)

// Tag is an interface to define metrics tags
type (
	Tag interface {
		Key() string
		Value() string
	}

	simpleMetric struct {
		key   string
		value string
	}
)

func (s simpleMetric) Key() string   { return s.key }
func (s simpleMetric) Value() string { return s.value }

func metricWithUnknown(key, value string) Tag {
	if len(value) == 0 {
		value = unknownValue
	}
	return simpleMetric{key: key, value: value}
}

// DomainTag returns a new domain tag. For timers, this also ensures that we
// dual emit the metric with the all tag. If a blank domain is provided then
// this converts that to an unknown domain.
func DomainTag(value string) Tag {
	return metricWithUnknown(domain, value)
}

// DomainUnknownTag returns a new domain:unknown tag-value
func DomainUnknownTag() Tag {
	return DomainTag("")
}

// InstanceTag returns a new instance tag
func InstanceTag(value string) Tag {
	return simpleMetric{key: instance, value: value}
}

// TargetClusterTag returns a new target cluster tag.
func TargetClusterTag(value string) Tag {
	return metricWithUnknown(targetCluster, value)
}

// ActiveClusterTag returns a new active cluster type tag.
func ActiveClusterTag(value string) Tag {
	return metricWithUnknown(activeCluster, value)
}

// TaskListTag returns a new task list tag.
func TaskListTag(value string) Tag {
	if len(value) == 0 {
		value = unknownValue
	}
	return simpleMetric{key: taskList, value: sanitizer.Value(value)}
}

// TaskListUnknownTag returns a new tasklist:unknown tag-value
func TaskListUnknownTag() Tag {
	return simpleMetric{key: taskList, value: unknownValue}
}

// TaskListTypeTag returns a new task list type tag.
func TaskListTypeTag(value string) Tag {
	return metricWithUnknown(taskListType, value)
}

// WorkflowTypeTag returns a new workflow type tag.
func WorkflowTypeTag(value string) Tag {
	return metricWithUnknown(workflowType, value)
}

// ActivityTypeTag returns a new activity type tag.
func ActivityTypeTag(value string) Tag {
	return metricWithUnknown(activityType, value)
}

// DecisionTypeTag returns a new decision type tag.
func DecisionTypeTag(value string) Tag {
	return metricWithUnknown(decisionType, value)
}

// InvariantTypeTag returns a new invariant type tag.
func InvariantTypeTag(value string) Tag {
	return metricWithUnknown(invariantType, value)
}

// KafkaPartitionTag returns a new KafkaPartition type tag.
func KafkaPartitionTag(value int32) Tag {
	return simpleMetric{key: kafkaPartition, value: strconv.Itoa(int(value))}
}

// ThriftTransportTag returns a new Thrift transport type tag.
func ThriftTransportTag() Tag {
	return simpleMetric{key: transport, value: transportThrift}
}

// GPRCTransportTag returns a new GRPC transport type tag.
func GPRCTransportTag() Tag {
	return simpleMetric{key: transport, value: transportGRPC}
}

// SignalNameTag returns a new SignalName tag
func SignalNameTag(value string) Tag {
	return metricWithUnknown(signalName, value)
}
