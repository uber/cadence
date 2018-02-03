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

// Client allows fetching values from a dynamic configuration system NOTE: This does not have async
// options right now. In the interest of keeping it minimal, we can add when requirement arises.
type Client interface {
	GetValue(name Key) (interface{}, error)
	GetValueWithConstraints(name Key, constraints map[ConstraintKey]interface{}) (interface{}, error)
}

// Key represents a key/property stored in dynamic config
type Key int

func (k Key) String() string {
	keys := []string{
		"unknownKey",
		"taskListActivitiesPerSecond",
	}
	if k <= unknownKey || k > TaskListActivitiesPerSecond {
		return keys[unknownKey]
	}
	return keys[k]
}

const (
	// Matching keys

	unknownKey Key = iota
	// TaskListActivitiesPerSecond represents number of activities allowed per tasklist
	TaskListActivitiesPerSecond
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

// NewCollection creates a new collection
func NewCollection(client Client) *Collection {
	return &Collection{client}
}

// Collection of values that are in dynamic config
type Collection struct {
	client Client
}

// GetProperty represents
func (c *Collection) GetProperty(key Key) func(float64) float64 {
	return func(defaultVal float64) float64 {
		val, err := c.client.GetValue(key)
		if err != nil {
			return defaultVal
		}
		return val.(float64)
	}
}
