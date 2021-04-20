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

package thrift

import (
	"github.com/uber/cadence/common/types"

	"github.com/uber/cadence/.gen/go/admin"
)

// FromAddSearchAttributeRequest converts internal AddSearchAttributeRequest type to thrift
func FromAddSearchAttributeRequest(t *types.AddSearchAttributeRequest) *admin.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &admin.AddSearchAttributeRequest{
		SearchAttribute: FromIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   &t.SecurityToken,
	}
}

// ToAddSearchAttributeRequest converts thrift AddSearchAttributeRequest type to internal
func ToAddSearchAttributeRequest(t *admin.AddSearchAttributeRequest) *types.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &types.AddSearchAttributeRequest{
		SearchAttribute: ToIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   t.GetSecurityToken(),
	}
}

// FromDescribeClusterResponse converts internal DescribeClusterResponse type to thrift
func FromDescribeClusterResponse(t *types.DescribeClusterResponse) *admin.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &admin.DescribeClusterResponse{
		SupportedClientVersions: FromSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          FromMembershipInfo(t.MembershipInfo),
		PersistenceInfo:         FromPersistenceInfoMap(t.PersistenceInfo),
	}
}

// ToDescribeClusterResponse converts thrift DescribeClusterResponse type to internal
func ToDescribeClusterResponse(t *admin.DescribeClusterResponse) *types.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeClusterResponse{
		SupportedClientVersions: ToSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          ToMembershipInfo(t.MembershipInfo),
		PersistenceInfo:         ToPersistenceInfoMap(t.PersistenceInfo),
	}
}

// FromAdminDescribeWorkflowExecutionRequest converts internal DescribeWorkflowExecutionRequest type to thrift
func FromAdminDescribeWorkflowExecutionRequest(t *types.AdminDescribeWorkflowExecutionRequest) *admin.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &admin.DescribeWorkflowExecutionRequest{
		Domain:    &t.Domain,
		Execution: FromWorkflowExecution(t.Execution),
	}
}

// ToAdminDescribeWorkflowExecutionRequest converts thrift DescribeWorkflowExecutionRequest type to internal
func ToAdminDescribeWorkflowExecutionRequest(t *admin.DescribeWorkflowExecutionRequest) *types.AdminDescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionRequest{
		Domain:    t.GetDomain(),
		Execution: ToWorkflowExecution(t.Execution),
	}
}

// FromAdminDescribeWorkflowExecutionResponse converts internal DescribeWorkflowExecutionResponse type to thrift
func FromAdminDescribeWorkflowExecutionResponse(t *types.AdminDescribeWorkflowExecutionResponse) *admin.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &admin.DescribeWorkflowExecutionResponse{
		ShardId:                &t.ShardID,
		HistoryAddr:            &t.HistoryAddr,
		MutableStateInCache:    &t.MutableStateInCache,
		MutableStateInDatabase: &t.MutableStateInDatabase,
	}
}

// ToAdminDescribeWorkflowExecutionResponse converts thrift DescribeWorkflowExecutionResponse type to internal
func ToAdminDescribeWorkflowExecutionResponse(t *admin.DescribeWorkflowExecutionResponse) *types.AdminDescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionResponse{
		ShardID:                t.GetShardId(),
		HistoryAddr:            t.GetHistoryAddr(),
		MutableStateInCache:    t.GetMutableStateInCache(),
		MutableStateInDatabase: t.GetMutableStateInDatabase(),
	}
}

// FromGetWorkflowExecutionRawHistoryV2Request converts internal GetWorkflowExecutionRawHistoryV2Request type to thrift
func FromGetWorkflowExecutionRawHistoryV2Request(t *types.GetWorkflowExecutionRawHistoryV2Request) *admin.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &admin.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            &t.Domain,
		Execution:         FromWorkflowExecution(t.Execution),
		StartEventId:      t.StartEventID,
		StartEventVersion: t.StartEventVersion,
		EndEventId:        t.EndEventID,
		EndEventVersion:   t.EndEventVersion,
		MaximumPageSize:   &t.MaximumPageSize,
		NextPageToken:     t.NextPageToken,
	}
}

// ToGetWorkflowExecutionRawHistoryV2Request converts thrift GetWorkflowExecutionRawHistoryV2Request type to internal
func ToGetWorkflowExecutionRawHistoryV2Request(t *admin.GetWorkflowExecutionRawHistoryV2Request) *types.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            t.GetDomain(),
		Execution:         ToWorkflowExecution(t.Execution),
		StartEventID:      t.StartEventId,
		StartEventVersion: t.StartEventVersion,
		EndEventID:        t.EndEventId,
		EndEventVersion:   t.EndEventVersion,
		MaximumPageSize:   t.GetMaximumPageSize(),
		NextPageToken:     t.NextPageToken,
	}
}

// FromGetWorkflowExecutionRawHistoryV2Response converts internal GetWorkflowExecutionRawHistoryV2Response type to thrift
func FromGetWorkflowExecutionRawHistoryV2Response(t *types.GetWorkflowExecutionRawHistoryV2Response) *admin.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &admin.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: FromDataBlobArray(t.HistoryBatches),
		VersionHistory: FromVersionHistory(t.VersionHistory),
	}
}

// ToGetWorkflowExecutionRawHistoryV2Response converts thrift GetWorkflowExecutionRawHistoryV2Response type to internal
func ToGetWorkflowExecutionRawHistoryV2Response(t *admin.GetWorkflowExecutionRawHistoryV2Response) *types.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: ToDataBlobArray(t.HistoryBatches),
		VersionHistory: ToVersionHistory(t.VersionHistory),
	}
}

// FromHostInfo converts internal HostInfo type to thrift
func FromHostInfo(t *types.HostInfo) *admin.HostInfo {
	if t == nil {
		return nil
	}
	return &admin.HostInfo{
		Identity: &t.Identity,
	}
}

// ToHostInfo converts thrift HostInfo type to internal
func ToHostInfo(t *admin.HostInfo) *types.HostInfo {
	if t == nil {
		return nil
	}
	return &types.HostInfo{
		Identity: t.GetIdentity(),
	}
}

// FromMembershipInfo converts internal MembershipInfo type to thrift
func FromMembershipInfo(t *types.MembershipInfo) *admin.MembershipInfo {
	if t == nil {
		return nil
	}
	return &admin.MembershipInfo{
		CurrentHost:      FromHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            FromRingInfoArray(t.Rings),
	}
}

// ToMembershipInfo converts thrift MembershipInfo type to internal
func ToMembershipInfo(t *admin.MembershipInfo) *types.MembershipInfo {
	if t == nil {
		return nil
	}
	return &types.MembershipInfo{
		CurrentHost:      ToHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            ToRingInfoArray(t.Rings),
	}
}

// FromPersistenceInfoMap converts internal map[string]*types.PersistenceInfo type to thrift
func FromPersistenceInfoMap(t map[string]*types.PersistenceInfo) map[string]*admin.PersistenceInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*admin.PersistenceInfo, len(t))
	for key := range t {
		v[key] = FromPersistenceInfo(t[key])
	}
	return v
}

// FromPersistenceInfo converts internal PersistenceInfo type to thrift
func FromPersistenceInfo(t *types.PersistenceInfo) *admin.PersistenceInfo {
	if t == nil {
		return nil
	}
	return &admin.PersistenceInfo{
		Backend:  &t.Backend,
		Settings: FromPersistenceSettings(t.Settings),
		Features: FromPersistenceFeatures(t.Features),
	}
}

// FromPersistenceSettings converts internal []*types.PersistenceSetting type to thrift
func FromPersistenceSettings(t []*types.PersistenceSetting) []*admin.PersistenceSetting {
	if t == nil {
		return nil
	}
	v := make([]*admin.PersistenceSetting, len(t))
	for i := range t {
		v[i] = FromPersistenceSetting(t[i])
	}
	return v
}

// FromPersistenceSetting converts internal PersistenceSetting type to thrift
func FromPersistenceSetting(t *types.PersistenceSetting) *admin.PersistenceSetting {
	if t == nil {
		return nil
	}
	return &admin.PersistenceSetting{
		Key:   &t.Key,
		Value: &t.Value,
	}
}

// FromPersistenceFeatures converts internal []*types.PersistenceFeature type to thrift
func FromPersistenceFeatures(t []*types.PersistenceFeature) []*admin.PersistenceFeature {
	if t == nil {
		return nil
	}
	v := make([]*admin.PersistenceFeature, len(t))
	for i := range t {
		v[i] = FromPersistenceFeature(t[i])
	}
	return v
}

// FromPersistenceFeature converts internal PersistenceFeature type to thrift
func FromPersistenceFeature(t *types.PersistenceFeature) *admin.PersistenceFeature {
	if t == nil {
		return nil
	}
	return &admin.PersistenceFeature{
		Key:     &t.Key,
		Enabled: &t.Enabled,
	}
}

// ToPersistenceInfoMap converts thrift to internal map[string]*types.PersistenceInfo type
func ToPersistenceInfoMap(t map[string]*admin.PersistenceInfo) map[string]*types.PersistenceInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*types.PersistenceInfo, len(t))
	for key := range t {
		v[key] = ToPersistenceInfo(t[key])
	}
	return v
}

// ToPersistenceInfo converts thrift to internal PersistenceInfo type
func ToPersistenceInfo(t *admin.PersistenceInfo) *types.PersistenceInfo {
	if t == nil {
		return nil
	}
	return &types.PersistenceInfo{
		Backend:  t.GetBackend(),
		Settings: ToPersistenceSettings(t.Settings),
		Features: ToPersistenceFeatures(t.Features),
	}
}

// ToPersistenceSettings converts thrift to internal []*types.PersistenceSetting type
func ToPersistenceSettings(t []*admin.PersistenceSetting) []*types.PersistenceSetting {
	if t == nil {
		return nil
	}
	v := make([]*types.PersistenceSetting, len(t))
	for i := range t {
		v[i] = ToPersistenceSetting(t[i])
	}
	return v
}

// ToPersistenceSetting converts from thrift to internal PersistenceSetting type
func ToPersistenceSetting(t *admin.PersistenceSetting) *types.PersistenceSetting {
	if t == nil {
		return nil
	}
	return &types.PersistenceSetting{
		Key:   t.GetKey(),
		Value: t.GetValue(),
	}
}

// ToPersistenceFeatures converts from thrift to internal []*types.PersistenceFeature type
func ToPersistenceFeatures(t []*admin.PersistenceFeature) []*types.PersistenceFeature {
	if t == nil {
		return nil
	}
	v := make([]*types.PersistenceFeature, len(t))
	for i := range t {
		v[i] = ToPersistenceFeature(t[i])
	}
	return v
}

// ToPersistenceFeature converts from thrift to internal PersistenceFeature type
func ToPersistenceFeature(t *admin.PersistenceFeature) *types.PersistenceFeature {
	if t == nil {
		return nil
	}
	return &types.PersistenceFeature{
		Key:     t.GetKey(),
		Enabled: t.GetEnabled(),
	}
}

// FromResendReplicationTasksRequest converts internal ResendReplicationTasksRequest type to thrift
func FromResendReplicationTasksRequest(t *types.ResendReplicationTasksRequest) *admin.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &admin.ResendReplicationTasksRequest{
		DomainID:      &t.DomainID,
		WorkflowID:    &t.WorkflowID,
		RunID:         &t.RunID,
		RemoteCluster: &t.RemoteCluster,
		StartEventID:  t.StartEventID,
		StartVersion:  t.StartVersion,
		EndEventID:    t.EndEventID,
		EndVersion:    t.EndVersion,
	}
}

// ToResendReplicationTasksRequest converts thrift ResendReplicationTasksRequest type to internal
func ToResendReplicationTasksRequest(t *admin.ResendReplicationTasksRequest) *types.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &types.ResendReplicationTasksRequest{
		DomainID:      t.GetDomainID(),
		WorkflowID:    t.GetWorkflowID(),
		RunID:         t.GetRunID(),
		RemoteCluster: t.GetRemoteCluster(),
		StartEventID:  t.StartEventID,
		StartVersion:  t.StartVersion,
		EndEventID:    t.EndEventID,
		EndVersion:    t.EndVersion,
	}
}

// FromRingInfo converts internal RingInfo type to thrift
func FromRingInfo(t *types.RingInfo) *admin.RingInfo {
	if t == nil {
		return nil
	}
	return &admin.RingInfo{
		Role:        &t.Role,
		MemberCount: &t.MemberCount,
		Members:     FromHostInfoArray(t.Members),
	}
}

// ToRingInfo converts thrift RingInfo type to internal
func ToRingInfo(t *admin.RingInfo) *types.RingInfo {
	if t == nil {
		return nil
	}
	return &types.RingInfo{
		Role:        t.GetRole(),
		MemberCount: t.GetMemberCount(),
		Members:     ToHostInfoArray(t.Members),
	}
}

// FromRingInfoArray converts internal RingInfo type array to thrift
func FromRingInfoArray(t []*types.RingInfo) []*admin.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*admin.RingInfo, len(t))
	for i := range t {
		v[i] = FromRingInfo(t[i])
	}
	return v
}

// ToRingInfoArray converts thrift RingInfo type array to internal
func ToRingInfoArray(t []*admin.RingInfo) []*types.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.RingInfo, len(t))
	for i := range t {
		v[i] = ToRingInfo(t[i])
	}
	return v
}

// FromHostInfoArray converts internal HostInfo type array to thrift
func FromHostInfoArray(t []*types.HostInfo) []*admin.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*admin.HostInfo, len(t))
	for i := range t {
		v[i] = FromHostInfo(t[i])
	}
	return v
}

// ToHostInfoArray converts thrift HostInfo type array to internal
func ToHostInfoArray(t []*admin.HostInfo) []*types.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.HostInfo, len(t))
	for i := range t {
		v[i] = ToHostInfo(t[i])
	}
	return v
}
