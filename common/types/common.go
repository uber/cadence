// The MIT License (MIT)
//
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

import (
	"time"
)

// WorkflowExecution identifies workflow execution
type WorkflowExecution struct {
	WorkflowId string
	RunId      string
}

// WorkflowType identifies workflow type
type WorkflowType struct {
	Name string
}

// ActivityType identifies activity type
type ActivityType struct {
	Name string
}

// Header is used for context propagation
type Header struct {
	Fields map[string][]byte
}

// Memo allows adding arbitrary key-value pairs to a workflow
type Memo struct {
	Fields map[string][]byte
}

// SearchAttributes allows adding indexed key-values pairs that can be searched by
type SearchAttributes struct {
	IndexedFields map[string][]byte
}

// RetryPolicy is a retry policy
type RetryPolicy struct {
	InitialInterval          time.Duration
	BackoffCoefficient       float64
	MaximumInterval          time.Duration
	MaximumAttempts          int32
	NonRetriableErrorReasons []string
	ExpirationInterval       time.Duration
}
