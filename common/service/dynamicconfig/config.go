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

import (
	"time"
)

// NewCollection creates a new collection
func NewCollection(client Client) *Collection {
	return &Collection{client}
}

// Collection wraps dynamic config client with a closure so that across the code, the config values
// can be directly accessed by calling the function without propagating the client everywhere in
// code
type Collection struct {
	client Client
}

// GetProperty gets a eface property and returns defaultValue if property is not found
func (c *Collection) GetProperty(key Key, defaultValue interface{}) func() interface{} {
	return func() interface{} {
		val, _ := c.client.GetValue(key, defaultValue)
		return val
	}
}

func getFilterMap(opts ...FilterOption) map[Filter]interface{} {
	l := len(opts)
	m := make(map[Filter]interface{}, l)
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// GetIntProperty gets property and asserts that it's an integer
func (c *Collection) GetIntProperty(key Key, defaultValue int) func(opts ...FilterOption) int {
	return func(opts ...FilterOption) int {
		val, _ := c.client.GetIntValue(key, getFilterMap(opts...), defaultValue)
		return val
	}
}

// GetFloat64Property gets property and asserts that it's a float64
func (c *Collection) GetFloat64Property(key Key, defaultValue float64) func(opts ...FilterOption) float64 {
	return func(opts ...FilterOption) float64 {
		val, _ := c.client.GetFloatValue(key, getFilterMap(opts...), defaultValue)
		return val
	}
}

// GetDurationProperty gets property and asserts that it's a duration
func (c *Collection) GetDurationProperty(key Key, defaultValue time.Duration) func(opts ...FilterOption) time.Duration {
	return func(opts ...FilterOption) time.Duration {
		val, _ := c.client.GetDurationValue(key, getFilterMap(opts...), defaultValue)
		return val
	}
}

// GetBoolProperty gets property and asserts that it's an bool
func (c *Collection) GetBoolProperty(key Key, defaultValue bool) func(opts ...FilterOption) bool {
	return func(opts ...FilterOption) bool {
		val, _ := c.client.GetBoolValue(key, getFilterMap(opts...), defaultValue)
		return val
	}
}
