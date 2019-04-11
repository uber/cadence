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

package log

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/service/dynamicconfig"

	"code.uber.internal/devexp/cadence-tools/go-build/.go/src/gb2/src/github.com/magiconair/properties/assert"

	"github.com/uber/cadence/common/log/tag"

	"go.uber.org/zap"
)

func TestDefaultLogger(t *testing.T) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)
	// copy the output in a separate goroutine so logging can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	var zapLogger *zap.Logger
	zapLogger = zap.NewExample()

	logger := NewLogger(zapLogger)
	logger.Info("test info", tag.WorkflowAction(tag.ValueActionActivityTaskCanceled))

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC
	assert.Equal(t, out, `{"level":"info","msg":"test info","logging-call-at":"logger_test.go:56","wf-action":"add-activitytask-canceled-event"}`+"\n")
}

func TestThrottleLogger(t *testing.T) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)
	// copy the output in a separate goroutine so logging can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	var zapLogger *zap.Logger
	zapLogger = zap.NewExample()

	dc := dynamicconfig.NewNopClient()
	cln := dynamicconfig.NewCollection(dc, bark.NewLoggerFromLogrus(logrus.New()))
	logger := NewThrottledLogger(zapLogger, cln.GetIntProperty(dynamicconfig.FrontendRPS, 1))
	logger.Info("test info", tag.WorkflowAction(tag.ValueActionActivityTaskCanceled))

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC
	assert.Equal(t, out, `{"level":"info","msg":"test info","logging-call-at":"logger_test.go:83","wf-action":"add-activitytask-canceled-event"}`+"\n")
}
