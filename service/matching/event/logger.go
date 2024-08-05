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

package event

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

var enabled = false

func init() {
	enabled = os.Getenv("MATCHING_LOG_EVENTS") == "true"
}

type E struct {
	persistence.TaskInfo
	TaskListName string
	TaskListKind *types.TaskListKind
	TaskListType int // persistence.TaskListTypeDecision or persistence.TaskListTypeActivity

	EventTime time.Time

	// EventName describes the event. It is used to query events in simulations so don't change existing event names.
	EventName string
	Host      string
	Payload   map[string]any
}

func Log(events ...E) {
	if !enabled {
		return
	}
	for _, e := range events {
		e.EventTime = time.Now()
		data, err := json.Marshal(e)
		if err != nil {
			fmt.Printf("failed to marshal event: %v", err)
		}

		fmt.Printf("Matching New Event: %s\n", data)
	}
}
