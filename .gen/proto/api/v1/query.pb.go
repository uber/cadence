// Copyright (c) 2020 Uber Technologies, Inc.
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: uber/cadence/api/v1/query.proto

package apiv1

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type QueryResultType int32

const (
	QueryResultType_QUERY_RESULT_TYPE_INVALID  QueryResultType = 0
	QueryResultType_QUERY_RESULT_TYPE_ANSWERED QueryResultType = 1
	QueryResultType_QUERY_RESULT_TYPE_FAILED   QueryResultType = 2
)

// Enum value maps for QueryResultType.
var (
	QueryResultType_name = map[int32]string{
		0: "QUERY_RESULT_TYPE_INVALID",
		1: "QUERY_RESULT_TYPE_ANSWERED",
		2: "QUERY_RESULT_TYPE_FAILED",
	}
	QueryResultType_value = map[string]int32{
		"QUERY_RESULT_TYPE_INVALID":  0,
		"QUERY_RESULT_TYPE_ANSWERED": 1,
		"QUERY_RESULT_TYPE_FAILED":   2,
	}
)

func (x QueryResultType) Enum() *QueryResultType {
	p := new(QueryResultType)
	*p = x
	return p
}

func (x QueryResultType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryResultType) Descriptor() protoreflect.EnumDescriptor {
	return file_uber_cadence_api_v1_query_proto_enumTypes[0].Descriptor()
}

func (QueryResultType) Type() protoreflect.EnumType {
	return &file_uber_cadence_api_v1_query_proto_enumTypes[0]
}

func (x QueryResultType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryResultType.Descriptor instead.
func (QueryResultType) EnumDescriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{0}
}

type QueryRejectCondition int32

const (
	QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID QueryRejectCondition = 0
	// QUERY_REJECT_CONDITION_NOT_OPEN indicates that query should be rejected if workflow is not open.
	QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN QueryRejectCondition = 1
	// QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY indicates that query should be rejected if workflow did not complete cleanly.
	QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY QueryRejectCondition = 2
)

// Enum value maps for QueryRejectCondition.
var (
	QueryRejectCondition_name = map[int32]string{
		0: "QUERY_REJECT_CONDITION_INVALID",
		1: "QUERY_REJECT_CONDITION_NOT_OPEN",
		2: "QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY",
	}
	QueryRejectCondition_value = map[string]int32{
		"QUERY_REJECT_CONDITION_INVALID":               0,
		"QUERY_REJECT_CONDITION_NOT_OPEN":              1,
		"QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY": 2,
	}
)

func (x QueryRejectCondition) Enum() *QueryRejectCondition {
	p := new(QueryRejectCondition)
	*p = x
	return p
}

func (x QueryRejectCondition) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryRejectCondition) Descriptor() protoreflect.EnumDescriptor {
	return file_uber_cadence_api_v1_query_proto_enumTypes[1].Descriptor()
}

func (QueryRejectCondition) Type() protoreflect.EnumType {
	return &file_uber_cadence_api_v1_query_proto_enumTypes[1]
}

func (x QueryRejectCondition) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryRejectCondition.Descriptor instead.
func (QueryRejectCondition) EnumDescriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{1}
}

type QueryConsistencyLevel int32

const (
	QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID QueryConsistencyLevel = 0
	// EVENTUAL indicates that query should be eventually consistent.
	QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL QueryConsistencyLevel = 1
	// STRONG indicates that any events that came before query should be reflected in workflow state before running query.
	QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG QueryConsistencyLevel = 2
)

// Enum value maps for QueryConsistencyLevel.
var (
	QueryConsistencyLevel_name = map[int32]string{
		0: "QUERY_CONSISTENCY_LEVEL_INVALID",
		1: "QUERY_CONSISTENCY_LEVEL_EVENTUAL",
		2: "QUERY_CONSISTENCY_LEVEL_STRONG",
	}
	QueryConsistencyLevel_value = map[string]int32{
		"QUERY_CONSISTENCY_LEVEL_INVALID":  0,
		"QUERY_CONSISTENCY_LEVEL_EVENTUAL": 1,
		"QUERY_CONSISTENCY_LEVEL_STRONG":   2,
	}
)

func (x QueryConsistencyLevel) Enum() *QueryConsistencyLevel {
	p := new(QueryConsistencyLevel)
	*p = x
	return p
}

func (x QueryConsistencyLevel) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryConsistencyLevel) Descriptor() protoreflect.EnumDescriptor {
	return file_uber_cadence_api_v1_query_proto_enumTypes[2].Descriptor()
}

func (QueryConsistencyLevel) Type() protoreflect.EnumType {
	return &file_uber_cadence_api_v1_query_proto_enumTypes[2]
}

func (x QueryConsistencyLevel) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryConsistencyLevel.Descriptor instead.
func (QueryConsistencyLevel) EnumDescriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{2}
}

type WorkflowQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueryType string   `protobuf:"bytes,1,opt,name=query_type,json=queryType,proto3" json:"query_type,omitempty"`
	QueryArgs *Payload `protobuf:"bytes,2,opt,name=query_args,json=queryArgs,proto3" json:"query_args,omitempty"`
}

func (x *WorkflowQuery) Reset() {
	*x = WorkflowQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_uber_cadence_api_v1_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowQuery) ProtoMessage() {}

func (x *WorkflowQuery) ProtoReflect() protoreflect.Message {
	mi := &file_uber_cadence_api_v1_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowQuery.ProtoReflect.Descriptor instead.
func (*WorkflowQuery) Descriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{0}
}

func (x *WorkflowQuery) GetQueryType() string {
	if x != nil {
		return x.QueryType
	}
	return ""
}

func (x *WorkflowQuery) GetQueryArgs() *Payload {
	if x != nil {
		return x.QueryArgs
	}
	return nil
}

type WorkflowQueryResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResultType   QueryResultType `protobuf:"varint,1,opt,name=result_type,json=resultType,proto3,enum=uber.cadence.api.v1.QueryResultType" json:"result_type,omitempty"`
	Answer       *Payload        `protobuf:"bytes,2,opt,name=answer,proto3" json:"answer,omitempty"`
	ErrorMessage string          `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *WorkflowQueryResult) Reset() {
	*x = WorkflowQueryResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_uber_cadence_api_v1_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowQueryResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowQueryResult) ProtoMessage() {}

func (x *WorkflowQueryResult) ProtoReflect() protoreflect.Message {
	mi := &file_uber_cadence_api_v1_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowQueryResult.ProtoReflect.Descriptor instead.
func (*WorkflowQueryResult) Descriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{1}
}

func (x *WorkflowQueryResult) GetResultType() QueryResultType {
	if x != nil {
		return x.ResultType
	}
	return QueryResultType_QUERY_RESULT_TYPE_INVALID
}

func (x *WorkflowQueryResult) GetAnswer() *Payload {
	if x != nil {
		return x.Answer
	}
	return nil
}

func (x *WorkflowQueryResult) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

type QueryRejected struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CloseStatus WorkflowExecutionCloseStatus `protobuf:"varint,1,opt,name=close_status,json=closeStatus,proto3,enum=uber.cadence.api.v1.WorkflowExecutionCloseStatus" json:"close_status,omitempty"`
}

func (x *QueryRejected) Reset() {
	*x = QueryRejected{}
	if protoimpl.UnsafeEnabled {
		mi := &file_uber_cadence_api_v1_query_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRejected) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRejected) ProtoMessage() {}

func (x *QueryRejected) ProtoReflect() protoreflect.Message {
	mi := &file_uber_cadence_api_v1_query_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRejected.ProtoReflect.Descriptor instead.
func (*QueryRejected) Descriptor() ([]byte, []int) {
	return file_uber_cadence_api_v1_query_proto_rawDescGZIP(), []int{2}
}

func (x *QueryRejected) GetCloseStatus() WorkflowExecutionCloseStatus {
	if x != nil {
		return x.CloseStatus
	}
	return WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID
}

var File_uber_cadence_api_v1_query_proto protoreflect.FileDescriptor

var file_uber_cadence_api_v1_query_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x75, 0x62, 0x65, 0x72, 0x2f, 0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x75, 0x62, 0x65, 0x72, 0x2e, 0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x75, 0x62, 0x65, 0x72, 0x2f, 0x63, 0x61, 0x64,
	0x65, 0x6e, 0x63, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x75, 0x62, 0x65, 0x72, 0x2f, 0x63,
	0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6b, 0x0a, 0x0d,
	0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1d, 0x0a,
	0x0a, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x71, 0x75, 0x65, 0x72, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3b, 0x0a, 0x0a,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x61, 0x72, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x75, 0x62, 0x65, 0x72, 0x2e, 0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x09,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x41, 0x72, 0x67, 0x73, 0x22, 0xb7, 0x01, 0x0a, 0x13, 0x57, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x45, 0x0a, 0x0b, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x75, 0x62, 0x65, 0x72, 0x2e, 0x63, 0x61,
	0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x72, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x34, 0x0a, 0x06, 0x61, 0x6e, 0x73, 0x77,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x75, 0x62, 0x65, 0x72, 0x2e,
	0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x12, 0x23,
	0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0x65, 0x0a, 0x0d, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x6a, 0x65,
	0x63, 0x74, 0x65, 0x64, 0x12, 0x54, 0x0a, 0x0c, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x5f, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x31, 0x2e, 0x75, 0x62, 0x65,
	0x72, 0x2e, 0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0b, 0x63,
	0x6c, 0x6f, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0x6e, 0x0a, 0x0f, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x0a,
	0x19, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x55, 0x4c, 0x54, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x1e, 0x0a, 0x1a,
	0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x55, 0x4c, 0x54, 0x5f, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x41, 0x4e, 0x53, 0x57, 0x45, 0x52, 0x45, 0x44, 0x10, 0x01, 0x12, 0x1c, 0x0a, 0x18,
	0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x55, 0x4c, 0x54, 0x5f, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x02, 0x2a, 0x91, 0x01, 0x0a, 0x14, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x1e, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x4a,
	0x45, 0x43, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x44, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x49, 0x4e,
	0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x23, 0x0a, 0x1f, 0x51, 0x55, 0x45, 0x52, 0x59,
	0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x44, 0x49, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x10, 0x01, 0x12, 0x30, 0x0a, 0x2c,
	0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x43, 0x4f, 0x4e,
	0x44, 0x49, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x43, 0x4f, 0x4d, 0x50, 0x4c,
	0x45, 0x54, 0x45, 0x44, 0x5f, 0x43, 0x4c, 0x45, 0x41, 0x4e, 0x4c, 0x59, 0x10, 0x02, 0x2a, 0x86,
	0x01, 0x0a, 0x15, 0x51, 0x75, 0x65, 0x72, 0x79, 0x43, 0x6f, 0x6e, 0x73, 0x69, 0x73, 0x74, 0x65,
	0x6e, 0x63, 0x79, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x23, 0x0a, 0x1f, 0x51, 0x55, 0x45, 0x52,
	0x59, 0x5f, 0x43, 0x4f, 0x4e, 0x53, 0x49, 0x53, 0x54, 0x45, 0x4e, 0x43, 0x59, 0x5f, 0x4c, 0x45,
	0x56, 0x45, 0x4c, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x24, 0x0a,
	0x20, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x43, 0x4f, 0x4e, 0x53, 0x49, 0x53, 0x54, 0x45, 0x4e,
	0x43, 0x59, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x55, 0x41,
	0x4c, 0x10, 0x01, 0x12, 0x22, 0x0a, 0x1e, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x43, 0x4f, 0x4e,
	0x53, 0x49, 0x53, 0x54, 0x45, 0x4e, 0x43, 0x59, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x53,
	0x54, 0x52, 0x4f, 0x4e, 0x47, 0x10, 0x02, 0x42, 0x56, 0x0a, 0x17, 0x63, 0x6f, 0x6d, 0x2e, 0x75,
	0x62, 0x65, 0x72, 0x2e, 0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x42, 0x08, 0x41, 0x70, 0x69, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2f,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x75, 0x62, 0x65, 0x72, 0x2f,
	0x63, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x2f, 0x2e, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x70, 0x69, 0x76, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_uber_cadence_api_v1_query_proto_rawDescOnce sync.Once
	file_uber_cadence_api_v1_query_proto_rawDescData = file_uber_cadence_api_v1_query_proto_rawDesc
)

func file_uber_cadence_api_v1_query_proto_rawDescGZIP() []byte {
	file_uber_cadence_api_v1_query_proto_rawDescOnce.Do(func() {
		file_uber_cadence_api_v1_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_uber_cadence_api_v1_query_proto_rawDescData)
	})
	return file_uber_cadence_api_v1_query_proto_rawDescData
}

var file_uber_cadence_api_v1_query_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_uber_cadence_api_v1_query_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_uber_cadence_api_v1_query_proto_goTypes = []interface{}{
	(QueryResultType)(0),              // 0: uber.cadence.api.v1.QueryResultType
	(QueryRejectCondition)(0),         // 1: uber.cadence.api.v1.QueryRejectCondition
	(QueryConsistencyLevel)(0),        // 2: uber.cadence.api.v1.QueryConsistencyLevel
	(*WorkflowQuery)(nil),             // 3: uber.cadence.api.v1.WorkflowQuery
	(*WorkflowQueryResult)(nil),       // 4: uber.cadence.api.v1.WorkflowQueryResult
	(*QueryRejected)(nil),             // 5: uber.cadence.api.v1.QueryRejected
	(*Payload)(nil),                   // 6: uber.cadence.api.v1.Payload
	(WorkflowExecutionCloseStatus)(0), // 7: uber.cadence.api.v1.WorkflowExecutionCloseStatus
}
var file_uber_cadence_api_v1_query_proto_depIdxs = []int32{
	6, // 0: uber.cadence.api.v1.WorkflowQuery.query_args:type_name -> uber.cadence.api.v1.Payload
	0, // 1: uber.cadence.api.v1.WorkflowQueryResult.result_type:type_name -> uber.cadence.api.v1.QueryResultType
	6, // 2: uber.cadence.api.v1.WorkflowQueryResult.answer:type_name -> uber.cadence.api.v1.Payload
	7, // 3: uber.cadence.api.v1.QueryRejected.close_status:type_name -> uber.cadence.api.v1.WorkflowExecutionCloseStatus
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_uber_cadence_api_v1_query_proto_init() }
func file_uber_cadence_api_v1_query_proto_init() {
	if File_uber_cadence_api_v1_query_proto != nil {
		return
	}
	file_uber_cadence_api_v1_common_proto_init()
	file_uber_cadence_api_v1_workflow_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_uber_cadence_api_v1_query_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_uber_cadence_api_v1_query_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowQueryResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_uber_cadence_api_v1_query_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryRejected); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_uber_cadence_api_v1_query_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_uber_cadence_api_v1_query_proto_goTypes,
		DependencyIndexes: file_uber_cadence_api_v1_query_proto_depIdxs,
		EnumInfos:         file_uber_cadence_api_v1_query_proto_enumTypes,
		MessageInfos:      file_uber_cadence_api_v1_query_proto_msgTypes,
	}.Build()
	File_uber_cadence_api_v1_query_proto = out.File
	file_uber_cadence_api_v1_query_proto_rawDesc = nil
	file_uber_cadence_api_v1_query_proto_goTypes = nil
	file_uber_cadence_api_v1_query_proto_depIdxs = nil
}
