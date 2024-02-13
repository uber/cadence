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

package consumer

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type fakeMessageConsumer struct {
	// input
	failToStart bool
	ch          chan messaging.Message

	// output
	stopped bool
}

func (f *fakeMessageConsumer) Start() error {
	if f.failToStart {
		return errors.New("failed to start")
	}

	return nil
}

func (f *fakeMessageConsumer) Stop() {
	f.stopped = true
}

func (f *fakeMessageConsumer) Messages() <-chan messaging.Message {
	return f.ch
}

type fakeMessage struct {
	// input
	val     []byte
	wantAck bool

	// output
	acked  bool
	nacked bool
}

func (m *fakeMessage) Value() []byte {
	return m.val
}

func (m *fakeMessage) Partition() int32 {
	return 0
}

func (m *fakeMessage) Offset() int64 {
	return 0
}

func (m *fakeMessage) Ack() error {
	m.acked = true
	return nil
}

func (m *fakeMessage) Nack() error {
	m.nacked = true
	return nil
}

func TestDefaultConsumer(t *testing.T) {
	tests := []struct {
		name                     string
		innerConsumerFailToStart bool
		frontendFails            bool
		closeChanBeforeStop      bool
		msgs                     []*fakeMessage
	}{
		{
			name:                     "failed to start",
			innerConsumerFailToStart: true,
		},
		{
			name: "invalid messages",
			msgs: []*fakeMessage{
				{val: []byte("invalid payload"), wantAck: false},
				{val: []byte("invalid payload 2"), wantAck: false},
			},
		},
		{
			name:          "unsupported request type",
			frontendFails: true,
			msgs: []*fakeMessage{
				{val: mustGenerateUnsupportedRequestMsg(t), wantAck: false},
			},
		},
		{
			name:          "startworkflow request with invalid payload content",
			frontendFails: true,
			msgs: []*fakeMessage{
				{val: mustGenerateStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, false), wantAck: false},
			},
		},
		{
			name:          "startworkflowfrontend fails to respond",
			frontendFails: true,
			msgs: []*fakeMessage{
				{val: mustGenerateStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, true), wantAck: false},
			},
		},
		{
			name: "startworkflow unsupported encoding type",
			msgs: []*fakeMessage{
				{val: mustGenerateStartWorkflowExecutionRequestMsg(t, common.EncodingTypeProto, true), wantAck: false},
			},
		},
		{
			name: "startworkflow ok",
			msgs: []*fakeMessage{
				{val: mustGenerateStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, true), wantAck: true},
			},
		},
		{
			name:                "startworkflow ok with chan closed before stopping",
			closeChanBeforeStop: true,
			msgs: []*fakeMessage{
				{val: mustGenerateStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, true), wantAck: true},
			},
		},
		// signal with start test cases
		{
			name:          "signalwithstartworkflow request with invalid payload content",
			frontendFails: true,
			msgs: []*fakeMessage{
				{val: mustGenerateSignalWithStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, false), wantAck: false},
			},
		},
		{
			name:          "signalwithstartworkflow frontend fails to respond",
			frontendFails: true,
			msgs: []*fakeMessage{
				{val: mustGenerateSignalWithStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, true), wantAck: false},
			},
		},
		{
			name: "signalwithstartworkflow unsupported encoding type",
			msgs: []*fakeMessage{
				{val: mustGenerateSignalWithStartWorkflowExecutionRequestMsg(t, common.EncodingTypeProto, true), wantAck: false},
			},
		},
		{
			name: "signalwithstartworkflow ok",
			msgs: []*fakeMessage{
				{val: mustGenerateSignalWithStartWorkflowExecutionRequestMsg(t, common.EncodingTypeJSON, true), wantAck: true},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeConsumer := &fakeMessageConsumer{
				ch:          make(chan messaging.Message),
				failToStart: tc.innerConsumerFailToStart,
			}

			mockFrontend := frontend.NewMockClient(gomock.NewController(t))
			// we fake 2 headers and pass them manually to the mock because "..." extension doesn't work with mocked interface
			opts := getYARPCOptions(fakeHeaders())
			if tc.frontendFails {
				mockFrontend.EXPECT().
					StartWorkflowExecution(gomock.Any(), gomock.Any(), opts[0], opts[1]).
					Return(nil, errors.New("failed")).AnyTimes()
				mockFrontend.EXPECT().
					SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), opts[0], opts[1]).
					Return(nil, errors.New("failed")).AnyTimes()
			} else {
				resp := &types.StartWorkflowExecutionResponse{RunID: "test-run-id"}
				mockFrontend.EXPECT().
					StartWorkflowExecution(gomock.Any(), gomock.Any(), opts[0], opts[1]).
					Return(resp, nil).AnyTimes()
				mockFrontend.EXPECT().
					SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), opts[0], opts[1]).
					Return(resp, nil).AnyTimes()
			}

			c := New("queueid1", fakeConsumer, testlogger.New(t), metrics.NewNoopMetricsClient(), mockFrontend, WithConcurrency(2))
			err := c.Start()
			if tc.innerConsumerFailToStart != (err != nil) {
				t.Errorf("Start() err: %v, wantErr: %v", err, tc.innerConsumerFailToStart)
			}
			if err != nil {
				return
			}

			for _, msg := range tc.msgs {
				fakeConsumer.ch <- msg
			}

			if tc.closeChanBeforeStop {
				close(fakeConsumer.ch)
			}

			c.Stop()
			if !fakeConsumer.stopped {
				t.Error("innerConsumer.Stop() not called")
			}

			for i, msg := range tc.msgs {
				if msg.wantAck && !msg.acked {
					t.Errorf("message %d not acked", i)
				}
				if !msg.wantAck && !msg.nacked {
					t.Errorf("message %d not nacked", i)
				}
			}
		})
	}
}

func mustGenerateStartWorkflowExecutionRequestMsg(t *testing.T, encodingType common.EncodingType, validPayload bool) []byte {
	startRequest := &types.StartWorkflowExecutionAsyncRequest{
		StartWorkflowExecutionRequest: &types.StartWorkflowExecutionRequest{
			Domain:       "test-domain",
			WorkflowID:   "test-workflow-id",
			WorkflowType: &types.WorkflowType{Name: "test-workflow-type"},
			Input:        []byte("test-input"),
		},
	}

	payload, err := json.Marshal(startRequest)
	if err != nil {
		t.Fatal(err)
	}

	if !validPayload {
		payload = []byte("invalid payload")
	}

	msg := &sqlblobs.AsyncRequestMessage{
		Type:     sqlblobs.AsyncRequestTypeStartWorkflowExecutionAsyncRequest.Ptr(),
		Header:   fakeHeaders(),
		Encoding: common.StringPtr(string(encodingType)),
		Payload:  payload,
	}

	res, err := codec.NewThriftRWEncoder().Encode(msg)
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func mustGenerateSignalWithStartWorkflowExecutionRequestMsg(t *testing.T, encodingType common.EncodingType, validPayload bool) []byte {
	signalWithStartRequest := &types.SignalWithStartWorkflowExecutionAsyncRequest{
		SignalWithStartWorkflowExecutionRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:       "test-domain",
			WorkflowID:   "test-workflow-id",
			WorkflowType: &types.WorkflowType{Name: "test-workflow-type"},
			Input:        []byte("test-input"),
			SignalName:   "test-signal-name",
		},
	}

	payload, err := json.Marshal(signalWithStartRequest)
	if err != nil {
		t.Fatal(err)
	}

	if !validPayload {
		payload = []byte("invalid payload")
	}

	msg := &sqlblobs.AsyncRequestMessage{
		Type:     sqlblobs.AsyncRequestTypeSignalWithStartWorkflowExecutionAsyncRequest.Ptr(),
		Header:   fakeHeaders(),
		Encoding: common.StringPtr(string(encodingType)),
		Payload:  payload,
	}

	res, err := codec.NewThriftRWEncoder().Encode(msg)
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func mustGenerateUnsupportedRequestMsg(t *testing.T) []byte {
	startRequest := &types.StartWorkflowExecutionAsyncRequest{
		StartWorkflowExecutionRequest: &types.StartWorkflowExecutionRequest{
			Domain:       "test-domain",
			WorkflowID:   "test-workflow-id",
			WorkflowType: &types.WorkflowType{Name: "test-workflow-type"},
			Input:        []byte("test-input"),
		},
	}

	payload, err := json.Marshal(startRequest)
	if err != nil {
		t.Fatal(err)
	}

	tp := sqlblobs.AsyncRequestType(-1)
	msg := &sqlblobs.AsyncRequestMessage{
		Type:     &tp,
		Header:   &shared.Header{},
		Encoding: common.StringPtr(string(common.EncodingTypeJSON)),
		Payload:  payload,
	}

	res, err := codec.NewThriftRWEncoder().Encode(msg)
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func fakeHeaders() *shared.Header {
	return &shared.Header{
		Fields: map[string][]byte{
			"key1": []byte("val1"),
			"key2": []byte("val2"),
		},
	}
}
