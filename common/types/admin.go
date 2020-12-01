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

package types

// AddSearchAttributeRequest is an internal type (TBD...)
type AddSearchAttributeRequest struct {
	SearchAttribute map[string]IndexedValueType `json:"searchAttribute,omitempty"`
	SecurityToken   *string                     `json:"securityToken,omitempty"`
}

// GetSearchAttribute is an internal getter (TBD...)
func (v *AddSearchAttributeRequest) GetSearchAttribute() (o map[string]IndexedValueType) {
	if v != nil && v.SearchAttribute != nil {
		return v.SearchAttribute
	}
	return
}

// GetSecurityToken is an internal getter (TBD...)
func (v *AddSearchAttributeRequest) GetSecurityToken() (o string) {
	if v != nil && v.SecurityToken != nil {
		return *v.SecurityToken
	}
	return
}

// AdminServiceAddSearchAttributeArgs is an internal type (TBD...)
type AdminServiceAddSearchAttributeArgs struct {
	Request *AddSearchAttributeRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceAddSearchAttributeArgs) GetRequest() (o *AddSearchAttributeRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceAddSearchAttributeResult is an internal type (TBD...)
type AdminServiceAddSearchAttributeResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceAddSearchAttributeResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceAddSearchAttributeResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceAddSearchAttributeResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// AdminServiceCloseShardArgs is an internal type (TBD...)
type AdminServiceCloseShardArgs struct {
	Request *CloseShardRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceCloseShardArgs) GetRequest() (o *CloseShardRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceCloseShardResult is an internal type (TBD...)
type AdminServiceCloseShardResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	AccessDeniedError    *AccessDeniedError    `json:"accessDeniedError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceCloseShardResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceCloseShardResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceCloseShardResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// AdminServiceDescribeClusterArgs is an internal type (TBD...)
type AdminServiceDescribeClusterArgs struct {
}

// AdminServiceDescribeClusterResult is an internal type (TBD...)
type AdminServiceDescribeClusterResult struct {
	Success              *DescribeClusterResponse `json:"success,omitempty"`
	InternalServiceError *InternalServiceError    `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError        `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceDescribeClusterResult) GetSuccess() (o *DescribeClusterResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceDescribeClusterResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceDescribeClusterResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// AdminServiceDescribeHistoryHostArgs is an internal type (TBD...)
type AdminServiceDescribeHistoryHostArgs struct {
	Request *DescribeHistoryHostRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceDescribeHistoryHostArgs) GetRequest() (o *DescribeHistoryHostRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceDescribeHistoryHostResult is an internal type (TBD...)
type AdminServiceDescribeHistoryHostResult struct {
	Success              *DescribeHistoryHostResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError             `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError        `json:"internalServiceError,omitempty"`
	AccessDeniedError    *AccessDeniedError           `json:"accessDeniedError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceDescribeHistoryHostResult) GetSuccess() (o *DescribeHistoryHostResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceDescribeHistoryHostResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceDescribeHistoryHostResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceDescribeHistoryHostResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// AdminServiceDescribeQueueArgs is an internal type (TBD...)
type AdminServiceDescribeQueueArgs struct {
	Request *DescribeQueueRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceDescribeQueueArgs) GetRequest() (o *DescribeQueueRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceDescribeQueueResult is an internal type (TBD...)
type AdminServiceDescribeQueueResult struct {
	Success              *DescribeQueueResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError       `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError  `json:"internalServiceError,omitempty"`
	AccessDeniedError    *AccessDeniedError     `json:"accessDeniedError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceDescribeQueueResult) GetSuccess() (o *DescribeQueueResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceDescribeQueueResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceDescribeQueueResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceDescribeQueueResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// AdminServiceDescribeWorkflowExecutionArgs is an internal type (TBD...)
type AdminServiceDescribeWorkflowExecutionArgs struct {
	Request *AdminDescribeWorkflowExecutionRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionArgs) GetRequest() (o *AdminDescribeWorkflowExecutionRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceDescribeWorkflowExecutionResult is an internal type (TBD...)
type AdminServiceDescribeWorkflowExecutionResult struct {
	Success              *AdminDescribeWorkflowExecutionResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError                        `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError                   `json:"internalServiceError,omitempty"`
	EntityNotExistError  *EntityNotExistsError                   `json:"entityNotExistError,omitempty"`
	AccessDeniedError    *AccessDeniedError                      `json:"accessDeniedError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionResult) GetSuccess() (o *AdminDescribeWorkflowExecutionResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceDescribeWorkflowExecutionResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// AdminServiceGetDLQReplicationMessagesArgs is an internal type (TBD...)
type AdminServiceGetDLQReplicationMessagesArgs struct {
	Request *GetDLQReplicationMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceGetDLQReplicationMessagesArgs) GetRequest() (o *GetDLQReplicationMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceGetDLQReplicationMessagesResult is an internal type (TBD...)
type AdminServiceGetDLQReplicationMessagesResult struct {
	Success          *GetDLQReplicationMessagesResponse `json:"success,omitempty"`
	BadRequestError  *BadRequestError                   `json:"badRequestError,omitempty"`
	ServiceBusyError *ServiceBusyError                  `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceGetDLQReplicationMessagesResult) GetSuccess() (o *GetDLQReplicationMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceGetDLQReplicationMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceGetDLQReplicationMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// AdminServiceGetDomainReplicationMessagesArgs is an internal type (TBD...)
type AdminServiceGetDomainReplicationMessagesArgs struct {
	Request *GetDomainReplicationMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesArgs) GetRequest() (o *GetDomainReplicationMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceGetDomainReplicationMessagesResult is an internal type (TBD...)
type AdminServiceGetDomainReplicationMessagesResult struct {
	Success                        *GetDomainReplicationMessagesResponse `json:"success,omitempty"`
	BadRequestError                *BadRequestError                      `json:"badRequestError,omitempty"`
	LimitExceededError             *LimitExceededError                   `json:"limitExceededError,omitempty"`
	ServiceBusyError               *ServiceBusyError                     `json:"serviceBusyError,omitempty"`
	ClientVersionNotSupportedError *ClientVersionNotSupportedError       `json:"clientVersionNotSupportedError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesResult) GetSuccess() (o *GetDomainReplicationMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetClientVersionNotSupportedError is an internal getter (TBD...)
func (v *AdminServiceGetDomainReplicationMessagesResult) GetClientVersionNotSupportedError() (o *ClientVersionNotSupportedError) {
	if v != nil && v.ClientVersionNotSupportedError != nil {
		return v.ClientVersionNotSupportedError
	}
	return
}

// AdminServiceGetReplicationMessagesArgs is an internal type (TBD...)
type AdminServiceGetReplicationMessagesArgs struct {
	Request *GetReplicationMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesArgs) GetRequest() (o *GetReplicationMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceGetReplicationMessagesResult is an internal type (TBD...)
type AdminServiceGetReplicationMessagesResult struct {
	Success                        *GetReplicationMessagesResponse `json:"success,omitempty"`
	BadRequestError                *BadRequestError                `json:"badRequestError,omitempty"`
	LimitExceededError             *LimitExceededError             `json:"limitExceededError,omitempty"`
	ServiceBusyError               *ServiceBusyError               `json:"serviceBusyError,omitempty"`
	ClientVersionNotSupportedError *ClientVersionNotSupportedError `json:"clientVersionNotSupportedError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesResult) GetSuccess() (o *GetReplicationMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetClientVersionNotSupportedError is an internal getter (TBD...)
func (v *AdminServiceGetReplicationMessagesResult) GetClientVersionNotSupportedError() (o *ClientVersionNotSupportedError) {
	if v != nil && v.ClientVersionNotSupportedError != nil {
		return v.ClientVersionNotSupportedError
	}
	return
}

// AdminServiceGetWorkflowExecutionRawHistoryV2Args is an internal type (TBD...)
type AdminServiceGetWorkflowExecutionRawHistoryV2Args struct {
	GetRequest *GetWorkflowExecutionRawHistoryV2Request `json:"getRequest,omitempty"`
}

// GetGetRequest is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Args) GetGetRequest() (o *GetWorkflowExecutionRawHistoryV2Request) {
	if v != nil && v.GetRequest != nil {
		return v.GetRequest
	}
	return
}

// AdminServiceGetWorkflowExecutionRawHistoryV2Result is an internal type (TBD...)
type AdminServiceGetWorkflowExecutionRawHistoryV2Result struct {
	Success              *GetWorkflowExecutionRawHistoryV2Response `json:"success,omitempty"`
	BadRequestError      *BadRequestError                          `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError                     `json:"internalServiceError,omitempty"`
	EntityNotExistError  *EntityNotExistsError                     `json:"entityNotExistError,omitempty"`
	ServiceBusyError     *ServiceBusyError                         `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Result) GetSuccess() (o *GetWorkflowExecutionRawHistoryV2Response) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceGetWorkflowExecutionRawHistoryV2Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// AdminServiceMergeDLQMessagesArgs is an internal type (TBD...)
type AdminServiceMergeDLQMessagesArgs struct {
	Request *MergeDLQMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesArgs) GetRequest() (o *MergeDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceMergeDLQMessagesResult is an internal type (TBD...)
type AdminServiceMergeDLQMessagesResult struct {
	Success              *MergeDLQMessagesResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError          `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError     `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError         `json:"serviceBusyError,omitempty"`
	EntityNotExistError  *EntityNotExistsError     `json:"entityNotExistError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesResult) GetSuccess() (o *MergeDLQMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceMergeDLQMessagesResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServicePurgeDLQMessagesArgs is an internal type (TBD...)
type AdminServicePurgeDLQMessagesArgs struct {
	Request *PurgeDLQMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServicePurgeDLQMessagesArgs) GetRequest() (o *PurgeDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServicePurgeDLQMessagesResult is an internal type (TBD...)
type AdminServicePurgeDLQMessagesResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
	EntityNotExistError  *EntityNotExistsError `json:"entityNotExistError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServicePurgeDLQMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServicePurgeDLQMessagesResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServicePurgeDLQMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServicePurgeDLQMessagesResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServiceReadDLQMessagesArgs is an internal type (TBD...)
type AdminServiceReadDLQMessagesArgs struct {
	Request *ReadDLQMessagesRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesArgs) GetRequest() (o *ReadDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceReadDLQMessagesResult is an internal type (TBD...)
type AdminServiceReadDLQMessagesResult struct {
	Success              *ReadDLQMessagesResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError         `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError    `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError        `json:"serviceBusyError,omitempty"`
	EntityNotExistError  *EntityNotExistsError    `json:"entityNotExistError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesResult) GetSuccess() (o *ReadDLQMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceReadDLQMessagesResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServiceReapplyEventsArgs is an internal type (TBD...)
type AdminServiceReapplyEventsArgs struct {
	ReapplyEventsRequest *ReapplyEventsRequest `json:"reapplyEventsRequest,omitempty"`
}

// GetReapplyEventsRequest is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsArgs) GetReapplyEventsRequest() (o *ReapplyEventsRequest) {
	if v != nil && v.ReapplyEventsRequest != nil {
		return v.ReapplyEventsRequest
	}
	return
}

// AdminServiceReapplyEventsResult is an internal type (TBD...)
type AdminServiceReapplyEventsResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	DomainNotActiveError *DomainNotActiveError `json:"domainNotActiveError,omitempty"`
	LimitExceededError   *LimitExceededError   `json:"limitExceededError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
	EntityNotExistError  *EntityNotExistsError `json:"entityNotExistError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsResult) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceReapplyEventsResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServiceRefreshWorkflowTasksArgs is an internal type (TBD...)
type AdminServiceRefreshWorkflowTasksArgs struct {
	Request *RefreshWorkflowTasksRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceRefreshWorkflowTasksArgs) GetRequest() (o *RefreshWorkflowTasksRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceRefreshWorkflowTasksResult is an internal type (TBD...)
type AdminServiceRefreshWorkflowTasksResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	DomainNotActiveError *DomainNotActiveError `json:"domainNotActiveError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
	EntityNotExistError  *EntityNotExistsError `json:"entityNotExistError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceRefreshWorkflowTasksResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *AdminServiceRefreshWorkflowTasksResult) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceRefreshWorkflowTasksResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceRefreshWorkflowTasksResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServiceRemoveTaskArgs is an internal type (TBD...)
type AdminServiceRemoveTaskArgs struct {
	Request *RemoveTaskRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceRemoveTaskArgs) GetRequest() (o *RemoveTaskRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceRemoveTaskResult is an internal type (TBD...)
type AdminServiceRemoveTaskResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	AccessDeniedError    *AccessDeniedError    `json:"accessDeniedError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceRemoveTaskResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceRemoveTaskResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceRemoveTaskResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// AdminServiceResendReplicationTasksArgs is an internal type (TBD...)
type AdminServiceResendReplicationTasksArgs struct {
	Request *ResendReplicationTasksRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceResendReplicationTasksArgs) GetRequest() (o *ResendReplicationTasksRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceResendReplicationTasksResult is an internal type (TBD...)
type AdminServiceResendReplicationTasksResult struct {
	BadRequestError     *BadRequestError      `json:"badRequestError,omitempty"`
	ServiceBusyError    *ServiceBusyError     `json:"serviceBusyError,omitempty"`
	EntityNotExistError *EntityNotExistsError `json:"entityNotExistError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceResendReplicationTasksResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *AdminServiceResendReplicationTasksResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *AdminServiceResendReplicationTasksResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// AdminServiceResetQueueArgs is an internal type (TBD...)
type AdminServiceResetQueueArgs struct {
	Request *ResetQueueRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *AdminServiceResetQueueArgs) GetRequest() (o *ResetQueueRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// AdminServiceResetQueueResult is an internal type (TBD...)
type AdminServiceResetQueueResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	AccessDeniedError    *AccessDeniedError    `json:"accessDeniedError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *AdminServiceResetQueueResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *AdminServiceResetQueueResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *AdminServiceResetQueueResult) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// DescribeClusterResponse is an internal type (TBD...)
type DescribeClusterResponse struct {
	SupportedClientVersions *SupportedClientVersions `json:"supportedClientVersions,omitempty"`
	MembershipInfo          *MembershipInfo          `json:"membershipInfo,omitempty"`
}

// GetSupportedClientVersions is an internal getter (TBD...)
func (v *DescribeClusterResponse) GetSupportedClientVersions() (o *SupportedClientVersions) {
	if v != nil && v.SupportedClientVersions != nil {
		return v.SupportedClientVersions
	}
	return
}

// GetMembershipInfo is an internal getter (TBD...)
func (v *DescribeClusterResponse) GetMembershipInfo() (o *MembershipInfo) {
	if v != nil && v.MembershipInfo != nil {
		return v.MembershipInfo
	}
	return
}

// AdminDescribeWorkflowExecutionRequest is an internal type (TBD...)
type AdminDescribeWorkflowExecutionRequest struct {
	Domain    *string            `json:"domain,omitempty"`
	Execution *WorkflowExecution `json:"execution,omitempty"`
}

// GetDomain is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// AdminDescribeWorkflowExecutionResponse is an internal type (TBD...)
type AdminDescribeWorkflowExecutionResponse struct {
	ShardID                *string `json:"shardId,omitempty"`
	HistoryAddr            *string `json:"historyAddr,omitempty"`
	MutableStateInCache    *string `json:"mutableStateInCache,omitempty"`
	MutableStateInDatabase *string `json:"mutableStateInDatabase,omitempty"`
}

// GetShardID is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionResponse) GetShardID() (o string) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}

// GetHistoryAddr is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionResponse) GetHistoryAddr() (o string) {
	if v != nil && v.HistoryAddr != nil {
		return *v.HistoryAddr
	}
	return
}

// GetMutableStateInCache is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionResponse) GetMutableStateInCache() (o string) {
	if v != nil && v.MutableStateInCache != nil {
		return *v.MutableStateInCache
	}
	return
}

// GetMutableStateInDatabase is an internal getter (TBD...)
func (v *AdminDescribeWorkflowExecutionResponse) GetMutableStateInDatabase() (o string) {
	if v != nil && v.MutableStateInDatabase != nil {
		return *v.MutableStateInDatabase
	}
	return
}

// GetWorkflowExecutionRawHistoryV2Request is an internal type (TBD...)
type GetWorkflowExecutionRawHistoryV2Request struct {
	Domain            *string            `json:"domain,omitempty"`
	Execution         *WorkflowExecution `json:"execution,omitempty"`
	StartEventID      *int64             `json:"startEventId,omitempty"`
	StartEventVersion *int64             `json:"startEventVersion,omitempty"`
	EndEventID        *int64             `json:"endEventId,omitempty"`
	EndEventVersion   *int64             `json:"endEventVersion,omitempty"`
	MaximumPageSize   *int32             `json:"maximumPageSize,omitempty"`
	NextPageToken     []byte             `json:"nextPageToken,omitempty"`
}

// GetDomain is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetStartEventID is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetStartEventID() (o int64) {
	if v != nil && v.StartEventID != nil {
		return *v.StartEventID
	}
	return
}

// GetStartEventVersion is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetStartEventVersion() (o int64) {
	if v != nil && v.StartEventVersion != nil {
		return *v.StartEventVersion
	}
	return
}

// GetEndEventID is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetEndEventID() (o int64) {
	if v != nil && v.EndEventID != nil {
		return *v.EndEventID
	}
	return
}

// GetEndEventVersion is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetEndEventVersion() (o int64) {
	if v != nil && v.EndEventVersion != nil {
		return *v.EndEventVersion
	}
	return
}

// GetMaximumPageSize is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}

// GetNextPageToken is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Request) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// GetWorkflowExecutionRawHistoryV2Response is an internal type (TBD...)
type GetWorkflowExecutionRawHistoryV2Response struct {
	NextPageToken  []byte          `json:"nextPageToken,omitempty"`
	HistoryBatches []*DataBlob     `json:"historyBatches,omitempty"`
	VersionHistory *VersionHistory `json:"versionHistory,omitempty"`
}

// GetNextPageToken is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Response) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// GetHistoryBatches is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Response) GetHistoryBatches() (o []*DataBlob) {
	if v != nil && v.HistoryBatches != nil {
		return v.HistoryBatches
	}
	return
}

// GetVersionHistory is an internal getter (TBD...)
func (v *GetWorkflowExecutionRawHistoryV2Response) GetVersionHistory() (o *VersionHistory) {
	if v != nil && v.VersionHistory != nil {
		return v.VersionHistory
	}
	return
}

// HostInfo is an internal type (TBD...)
type HostInfo struct {
	Identity *string `json:"Identity,omitempty"`
}

// GetIdentity is an internal getter (TBD...)
func (v *HostInfo) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

// MembershipInfo is an internal type (TBD...)
type MembershipInfo struct {
	CurrentHost      *HostInfo   `json:"currentHost,omitempty"`
	ReachableMembers []string    `json:"reachableMembers,omitempty"`
	Rings            []*RingInfo `json:"rings,omitempty"`
}

// GetCurrentHost is an internal getter (TBD...)
func (v *MembershipInfo) GetCurrentHost() (o *HostInfo) {
	if v != nil && v.CurrentHost != nil {
		return v.CurrentHost
	}
	return
}

// GetReachableMembers is an internal getter (TBD...)
func (v *MembershipInfo) GetReachableMembers() (o []string) {
	if v != nil && v.ReachableMembers != nil {
		return v.ReachableMembers
	}
	return
}

// GetRings is an internal getter (TBD...)
func (v *MembershipInfo) GetRings() (o []*RingInfo) {
	if v != nil && v.Rings != nil {
		return v.Rings
	}
	return
}

// ResendReplicationTasksRequest is an internal type (TBD...)
type ResendReplicationTasksRequest struct {
	DomainID      *string `json:"domainID,omitempty"`
	WorkflowID    *string `json:"workflowID,omitempty"`
	RunID         *string `json:"runID,omitempty"`
	RemoteCluster *string `json:"remoteCluster,omitempty"`
	StartEventID  *int64  `json:"startEventID,omitempty"`
	StartVersion  *int64  `json:"startVersion,omitempty"`
	EndEventID    *int64  `json:"endEventID,omitempty"`
	EndVersion    *int64  `json:"endVersion,omitempty"`
}

// GetDomainID is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}

// GetWorkflowID is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

// GetRunID is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

// GetRemoteCluster is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetRemoteCluster() (o string) {
	if v != nil && v.RemoteCluster != nil {
		return *v.RemoteCluster
	}
	return
}

// GetStartEventID is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetStartEventID() (o int64) {
	if v != nil && v.StartEventID != nil {
		return *v.StartEventID
	}
	return
}

// GetStartVersion is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetStartVersion() (o int64) {
	if v != nil && v.StartVersion != nil {
		return *v.StartVersion
	}
	return
}

// GetEndEventID is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetEndEventID() (o int64) {
	if v != nil && v.EndEventID != nil {
		return *v.EndEventID
	}
	return
}

// GetEndVersion is an internal getter (TBD...)
func (v *ResendReplicationTasksRequest) GetEndVersion() (o int64) {
	if v != nil && v.EndVersion != nil {
		return *v.EndVersion
	}
	return
}

// RingInfo is an internal type (TBD...)
type RingInfo struct {
	Role        *string     `json:"role,omitempty"`
	MemberCount *int32      `json:"memberCount,omitempty"`
	Members     []*HostInfo `json:"members,omitempty"`
}

// GetRole is an internal getter (TBD...)
func (v *RingInfo) GetRole() (o string) {
	if v != nil && v.Role != nil {
		return *v.Role
	}
	return
}

// GetMemberCount is an internal getter (TBD...)
func (v *RingInfo) GetMemberCount() (o int32) {
	if v != nil && v.MemberCount != nil {
		return *v.MemberCount
	}
	return
}

// GetMembers is an internal getter (TBD...)
func (v *RingInfo) GetMembers() (o []*HostInfo) {
	if v != nil && v.Members != nil {
		return v.Members
	}
	return
}
