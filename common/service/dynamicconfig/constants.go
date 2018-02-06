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

package dynamicconfig

// Key represents a key/property stored in dynamic config
type Key int

func (k Key) String() string {
	keys := []string{
		"unknownKey",
		"minTaskThrottlingBurstSize",
	}
	if k <= unknownKey || k > MinTaskThrottlingBurstSize {
		return keys[unknownKey]
	}
	return keys[k]
}

const (
	// Matching keys

	unknownKey Key = iota
	// MinTaskThrottlingBurstSize represents the minimum burst size for task list throttling
	MinTaskThrottlingBurstSize
)

// ConstraintKey represents a key for a constraint on the dynamic config key
type ConstraintKey int

func (k ConstraintKey) String() string {
	keys := []string{
		"unknownConstraintKey",
		"domainName",
		"taskListName",
	}
	if k <= unknownConstraintKey || k > TaskListName {
		return keys[unknownConstraintKey]
	}
	return keys[k]
}

const (
	unknownConstraintKey ConstraintKey = iota
	// DomainName is the domain name
	DomainName
	// TaskListName is the tasklist name
	TaskListName
)
