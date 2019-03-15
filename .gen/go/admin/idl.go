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

// Code generated by thriftrw v1.13.1. DO NOT EDIT.
// @generated

package admin

import (
	"github.com/uber/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/thriftreflect"
)

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "admin",
	Package:  "github.com/uber/cadence/.gen/go/admin",
	FilePath: "admin.thrift",
	SHA1:     "e66f95274a536a0894835e1fa0935ccc4ab994db",
	Includes: []*thriftreflect.ThriftModule{
		shared.ThriftModule,
	},
	Raw: rawIDL,
}

const rawIDL = "// Copyright (c) 2017 Uber Technologies, Inc.\n//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\nnamespace java com.uber.cadence.admin\n\ninclude \"shared.thrift\"\n\n/**\n* AdminService provides advanced APIs for debugging and analysis with admin privillege\n**/\nservice AdminService {\n  /**\n  * DescribeWorkflowExecution returns information about the internal states of workflow execution.\n  **/\n  DescribeWorkflowExecutionResponse DescribeWorkflowExecution(1: DescribeWorkflowExecutionRequest request)\n    throws (\n      1: shared.BadRequestError         badRequestError,\n      2: shared.InternalServiceError    internalServiceError,\n      3: shared.EntityNotExistsError    entityNotExistError,\n      4: shared.AccessDeniedError       accessDeniedError,\n    )\n\n  /**\n  * DescribeHistoryHost returns information about the internal states of a history host\n  **/\n  shared.DescribeHistoryHostResponse DescribeHistoryHost(1: shared.DescribeHistoryHostRequest request)\n    throws (\n      1: shared.BadRequestError       badRequestError,\n      2: shared.InternalServiceError  internalServiceError,\n      3: shared.AccessDeniedError     accessDeniedError,\n    )\n\n  /**\n  * Returns the raw history of specified workflow execution.  It fails with 'EntityNotExistError' if speficied workflow\n  * execution in unknown to the service.\n  **/\n  GetWorkflowExecutionRawHistoryResponse GetWorkflowExecutionRawHistory(1: GetWorkflowExecutionRawHistoryRequest getRequest)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.ServiceBusyError serviceBusyError,\n    )\n\n  /**\n  * DescribeTaskList returns information about the target tasklist, right now this API returns the\n  * pollers which polled this tasklist in last few minutes and status of tasklist's ackManager\n  * (readLevel, ackLevel, backlogCountHint and taskIDBlock).\n  **/\n  DescribeTaskListResponse DescribeTaskList(1: shared.DescribeTaskListRequest request)\n    throws (\n      1: shared.BadRequestError badRequestError,\n      2: shared.InternalServiceError internalServiceError,\n      3: shared.EntityNotExistsError entityNotExistError,\n      4: shared.LimitExceededError limitExceededError,\n      5: shared.ServiceBusyError serviceBusyError,\n    )\n}\n\nstruct DescribeWorkflowExecutionRequest {\n  10: optional string                       domain\n  20: optional shared.WorkflowExecution     execution\n}\n\nstruct DescribeWorkflowExecutionResponse{\n  10: optional string shardId\n  20: optional string historyAddr\n  40: optional string mutableStateInCache\n  50: optional string mutableStateInDatabase\n}\n\nstruct GetWorkflowExecutionRawHistoryRequest {\n  10: optional string domain\n  20: optional shared.WorkflowExecution execution\n  30: optional i64 (js.type = \"Long\") firstEventId\n  40: optional i64 (js.type = \"Long\") nextEventId\n  50: optional i32 maximumPageSize\n  60: optional binary nextPageToken\n}\n\nstruct GetWorkflowExecutionRawHistoryResponse {\n  10: optional binary nextPageToken\n  20: optional list<shared.DataBlob> historyBatches\n  30: optional map<string, shared.ReplicationInfo> replicationInfo\n  40: optional i32 eventStoreVersion\n}\n\nstruct TaskIDBlock {\n  10: optional i64 (js.type = \"Long\")  startID\n  20: optional i64 (js.type = \"Long\")  endID\n}\n\nstruct DescribeTaskListResponse {\n  10: optional list<shared.PollerInfo> pollers\n  20: optional i64 (js.type = \"Long\") backlogCountHint\n  30: optional i64 (js.type = \"Long\") readLevel\n  40: optional i64 (js.type = \"Long\") ackLevel\n  50: optional TaskIDBlock taskIDBlock\n}\n"
