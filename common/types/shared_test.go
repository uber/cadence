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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignalWithStartWorkflowExecutionRequestSerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		input               *SignalWithStartWorkflowExecutionRequest
		expectedOutput      string
		expectedErrorOutput error
	}{
		"complete request without error": {
			input:               createNewSignalWithStartWorkflowExecutionRequest(),
			expectedOutput:      "{\"domain\":\"testDomain\",\"workflowId\":\"testWorkflowID\",\"workflowType\":{\"name\":\"testWorkflowType\"},\"taskList\":{\"name\":\"testTaskList\",\"kind\":\"STICKY\"},\"executionStartToCloseTimeoutSeconds\":1,\"taskStartToCloseTimeoutSeconds\":1,\"identity\":\"testIdentity\",\"requestId\":\"DF66E35D-A5B0-425D-8731-6AAC4A4B6368\",\"workflowIdReusePolicy\":\"AllowDuplicate\",\"signalName\":\"testRequest\",\"signalInput\":\"dGVzdFNpZ25hbElucHV0\",\"control\":\"dGVzdENvbnRyb2w=\",\"retryPolicy\":{\"initialIntervalInSeconds\":1,\"backoffCoefficient\":1,\"maximumIntervalInSeconds\":1,\"maximumAttempts\":1,\"nonRetriableErrorReasons\":[\"testArray\"],\"expirationIntervalInSeconds\":1},\"cronSchedule\":\"testSchedule\",\"header\":{},\"delayStartSeconds\":1,\"jitterStartSeconds\":1}",
			expectedErrorOutput: nil,
		},

		"empty request without error": {
			input:               &SignalWithStartWorkflowExecutionRequest{},
			expectedOutput:      "{}",
			expectedErrorOutput: nil,
		},

		"nil request without error": {
			input:               nil,
			expectedOutput:      "",
			expectedErrorOutput: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output, err := test.input.SerializeForLogging()
				assert.Equal(t, test.expectedOutput, output)
				assert.Equal(t, test.expectedErrorOutput, err)
				assert.NotContains(t, output, "PII")
			})
		})
	}
}

func TestPiiSampleRequestSerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		input               *PiiSampleRequest
		expectedOutput      string
		expectedErrorOutput error
	}{
		"complete request without error": {
			input:               createPiiSampleRequest(),
			expectedOutput:      "{}",
			expectedErrorOutput: nil,
		},

		"empty request without error": {
			input:               &PiiSampleRequest{},
			expectedOutput:      "{}",
			expectedErrorOutput: nil,
		},

		"nil request without error": {
			input:               nil,
			expectedOutput:      "",
			expectedErrorOutput: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output, err := test.input.SerializeForLogging()
				assert.Equal(t, test.expectedOutput, output)
				assert.Equal(t, test.expectedErrorOutput, err)
				assert.NotContains(t, output, "PII")
			})
		})
	}
}

func TestSignalWorkflowExecutionRequestSerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		input               *SignalWorkflowExecutionRequest
		expectedOutput      string
		expectedErrorOutput error
	}{
		"complete request without error": {
			input: &SignalWorkflowExecutionRequest{
				Domain:     "testDomain",
				Input:      []byte("testInputPII"),
				Identity:   "testIdentity",
				RequestID:  "DF66E35D-A5B0-425D-8731-6AAC4A4B6368",
				SignalName: "testRequest",
				Control:    []byte("testControl"),
			},
			expectedOutput:      "{\"domain\":\"testDomain\",\"signalName\":\"testRequest\",\"identity\":\"testIdentity\",\"requestId\":\"DF66E35D-A5B0-425D-8731-6AAC4A4B6368\",\"control\":\"dGVzdENvbnRyb2w=\"}",
			expectedErrorOutput: nil,
		},

		"empty request without error": {
			input:               &SignalWorkflowExecutionRequest{},
			expectedOutput:      "{}",
			expectedErrorOutput: nil,
		},

		"nil request without error": {
			input:               nil,
			expectedOutput:      "",
			expectedErrorOutput: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output, err := test.input.SerializeForLogging()
				assert.Equal(t, test.expectedOutput, output)
				assert.Equal(t, test.expectedErrorOutput, err)
				assert.NotContains(t, output, "PII")
			})
		})
	}
}

func TestStartWorkflowExecutionRequestRequestSerializeForLogging(t *testing.T) {
	testTasklistKind := TaskListKind(1)
	testExecutionStartToCloseTimeoutSeconds := int32(1)
	testTaskStartToCloseTimeoutSeconds := int32(1)
	testWorkflowIDReusePolicy := WorkflowIDReusePolicy(1)
	testDelayStartSeconds := int32(1)
	testJitterStartSeconds := int32(1)

	tests := map[string]struct {
		input               *StartWorkflowExecutionRequest
		expectedOutput      string
		expectedErrorOutput error
	}{
		"complete request without error": {
			input: &StartWorkflowExecutionRequest{
				Domain:       "testDomain",
				WorkflowID:   "testWorkflowID",
				WorkflowType: &WorkflowType{Name: "testWorkflowType"},
				TaskList: &TaskList{
					Name: "testTaskList",
					Kind: &testTasklistKind,
				},
				Input:                               []byte("testInputPII"),
				ExecutionStartToCloseTimeoutSeconds: &testExecutionStartToCloseTimeoutSeconds,
				TaskStartToCloseTimeoutSeconds:      &testTaskStartToCloseTimeoutSeconds,
				Identity:                            "testIdentity",
				RequestID:                           "DF66E35D-A5B0-425D-8731-6AAC4A4B6368",
				WorkflowIDReusePolicy:               &testWorkflowIDReusePolicy,
				RetryPolicy: &RetryPolicy{
					InitialIntervalInSeconds:    1,
					BackoffCoefficient:          1,
					MaximumIntervalInSeconds:    1,
					MaximumAttempts:             1,
					NonRetriableErrorReasons:    []string{"testArray"},
					ExpirationIntervalInSeconds: 1,
				},
				CronSchedule:       "testSchedule",
				Memo:               &Memo{Fields: map[string][]byte{}},
				SearchAttributes:   &SearchAttributes{IndexedFields: map[string][]byte{}},
				Header:             &Header{Fields: map[string][]byte{}},
				DelayStartSeconds:  &testDelayStartSeconds,
				JitterStartSeconds: &testJitterStartSeconds,
			},
			expectedOutput:      "{\"domain\":\"testDomain\",\"workflowId\":\"testWorkflowID\",\"workflowType\":{\"name\":\"testWorkflowType\"},\"taskList\":{\"name\":\"testTaskList\",\"kind\":\"STICKY\"},\"executionStartToCloseTimeoutSeconds\":1,\"taskStartToCloseTimeoutSeconds\":1,\"identity\":\"testIdentity\",\"requestId\":\"DF66E35D-A5B0-425D-8731-6AAC4A4B6368\",\"workflowIdReusePolicy\":\"AllowDuplicate\",\"retryPolicy\":{\"initialIntervalInSeconds\":1,\"backoffCoefficient\":1,\"maximumIntervalInSeconds\":1,\"maximumAttempts\":1,\"nonRetriableErrorReasons\":[\"testArray\"],\"expirationIntervalInSeconds\":1},\"cronSchedule\":\"testSchedule\",\"header\":{},\"delayStartSeconds\":1,\"jitterStartSeconds\":1}",
			expectedErrorOutput: nil,
		},

		"empty request without error": {
			input:               &StartWorkflowExecutionRequest{},
			expectedOutput:      "{}",
			expectedErrorOutput: nil,
		},

		"nil request without error": {
			input:               nil,
			expectedOutput:      "",
			expectedErrorOutput: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output, err := test.input.SerializeForLogging()
				assert.Equal(t, test.expectedOutput, output)
				assert.Equal(t, test.expectedErrorOutput, err)
				assert.NotContains(t, output, "PII")
			})
		})
	}
}

func TestSerializeRequest(t *testing.T) {
	// test serializing a normal request
	testReq := createNewSignalWithStartWorkflowExecutionRequest()
	serializeRes, err := SerializeRequest(testReq)

	expectRes := "{\"domain\":\"testDomain\",\"workflowId\":\"testWorkflowID\",\"workflowType\":{\"name\":\"testWorkflowType\"},\"taskList\":{\"name\":\"testTaskList\",\"kind\":\"STICKY\"},\"executionStartToCloseTimeoutSeconds\":1,\"taskStartToCloseTimeoutSeconds\":1,\"identity\":\"testIdentity\",\"requestId\":\"DF66E35D-A5B0-425D-8731-6AAC4A4B6368\",\"workflowIdReusePolicy\":\"AllowDuplicate\",\"signalName\":\"testRequest\",\"control\":\"dGVzdENvbnRyb2w=\",\"retryPolicy\":{\"initialIntervalInSeconds\":1,\"backoffCoefficient\":1,\"maximumIntervalInSeconds\":1,\"maximumAttempts\":1,\"nonRetriableErrorReasons\":[\"testArray\"],\"expirationIntervalInSeconds\":1},\"cronSchedule\":\"testSchedule\",\"header\":{},\"delayStartSeconds\":1,\"jitterStartSeconds\":1}"
	expectErr := error(nil)

	assert.Equal(t, expectRes, serializeRes)
	assert.Equal(t, expectErr, err)

	// test serializing a request that only contains PII
	testPiiReq := createPiiSampleRequest()
	serializePiiRes, piiErr := SerializeRequest(testPiiReq)

	expectPiiRes := "{}"
	expectPiiErr := error(nil)

	assert.Equal(t, expectPiiRes, serializePiiRes)
	assert.Equal(t, expectPiiErr, piiErr)

	assert.NotPanics(t, func() {
		SerializeRequest(nil)
	})

}

func createNewSignalWithStartWorkflowExecutionRequest() *SignalWithStartWorkflowExecutionRequest {
	testTasklistKind := TaskListKind(1)
	testExecutionStartToCloseTimeoutSeconds := int32(1)
	testTaskStartToCloseTimeoutSeconds := int32(1)
	testWorkflowIDReusePolicy := WorkflowIDReusePolicy(1)
	testDelayStartSeconds := int32(1)
	testJitterStartSeconds := int32(1)
	testReq := &SignalWithStartWorkflowExecutionRequest{
		Domain:       "testDomain",
		WorkflowID:   "testWorkflowID",
		WorkflowType: &WorkflowType{Name: "testWorkflowType"},
		TaskList: &TaskList{
			Name: "testTaskList",
			Kind: &testTasklistKind,
		},
		Input:                               []byte("testInputPII"),
		ExecutionStartToCloseTimeoutSeconds: &testExecutionStartToCloseTimeoutSeconds,
		TaskStartToCloseTimeoutSeconds:      &testTaskStartToCloseTimeoutSeconds,
		Identity:                            "testIdentity",
		RequestID:                           "DF66E35D-A5B0-425D-8731-6AAC4A4B6368",
		WorkflowIDReusePolicy:               &testWorkflowIDReusePolicy,
		SignalName:                          "testRequest",
		SignalInput:                         []byte("testSignalInputPII"),
		Control:                             []byte("testControl"),
		RetryPolicy: &RetryPolicy{
			InitialIntervalInSeconds:    1,
			BackoffCoefficient:          1,
			MaximumIntervalInSeconds:    1,
			MaximumAttempts:             1,
			NonRetriableErrorReasons:    []string{"testArray"},
			ExpirationIntervalInSeconds: 1,
		},
		CronSchedule:       "testSchedule",
		Memo:               &Memo{Fields: map[string][]byte{}},
		SearchAttributes:   &SearchAttributes{IndexedFields: map[string][]byte{}},
		Header:             &Header{Fields: map[string][]byte{}},
		DelayStartSeconds:  &testDelayStartSeconds,
		JitterStartSeconds: &testJitterStartSeconds,
	}
	return testReq
}

type PiiSampleRequest struct {
	Input            []byte            `json:"-"` // Filtering PII
	Memo             *Memo             `json:"-"` // Filtering PII
	SearchAttributes *SearchAttributes `json:"-"` // Filtering PII
	SignalInput      []byte            `json:"-"` // Filtering PII
}

func (v *PiiSampleRequest) SerializeForLogging() (string, error) {
	if v == nil {
		return "", nil
	}
	return SerializeRequest(v)
}

func createPiiSampleRequest() *PiiSampleRequest {
	piiTestArray := []byte("testInputPII")
	testMap := make(map[string][]byte)
	testMap["PII"] = piiTestArray

	testReq := &PiiSampleRequest{
		Input:            piiTestArray,
		Memo:             &Memo{Fields: testMap},
		SearchAttributes: &SearchAttributes{IndexedFields: testMap},
		SignalInput:      piiTestArray,
	}

	return testReq
}
