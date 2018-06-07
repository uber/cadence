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

namespace java com.uber.cadence.admin

include "shared.thrift"

/**
* AdminService provides advanced APIs for debugging and analysis with admin privillege
**/
service AdminService {
  /**
  * DescribeWorkflowExecution returns information about the internal states of workflow execution.
  **/
  DescribeWorkflowExecutionResponse DescribeWorkflowExecution(1: DescribeWorkflowExecutionRequest request)
    throws (
      1: shared.BadRequestError badRequestError,
      2: shared.InternalServiceError internalServiceError,
      3: shared.EntityNotExistsError entityNotExistError,
      4: shared.AccessDeniedError accessDeniedError,
    )

  /**
    * DescribeHistoryHost returns information about the internal states of a history host
    **/
    DescribeHistoryHostResponse DescribeHistoryHost(1: DescribeHistoryHostRequest request)
      throws (
        1: shared.BadRequestError badRequestError,
        2: shared.InternalServiceError internalServiceError,
        3: shared.AccessDeniedError accessDeniedError,
      )
}

struct DescribeWorkflowExecutionRequest {
  10: optional string domain
  20: optional shared.WorkflowExecution execution
}

struct DescribeWorkflowExecutionResponse{
  10: optional string shardId
  20: optional string historyAddr
  40: optional string mutableStateInCache
  50: optional string mutableStateInDatabase
}

//At least one of the parameters needs to be provided
struct DescribeHistoryHostRequest {
  10: optional string hostAddress //ip:port
  20: optional i32 shardIdForHost
  30: optional shared.WorkflowExecution executionForHost
}

struct DescribeHistoryHostResponse{
  10: optional i32 numberOfShards
  20: optional list<i32> shardIDs
  30: optional DomainCache domainCache
  40: optional string shardControllerStatus
  50: optional string address
}

struct DomainCache{
  10: optional i64 numOfItemsInCacheByID
  20: optional i64 numOfItemsInCacheByName
}