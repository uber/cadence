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

import "errors"

// Client allows fetching values from a dynamic configuration system NOTE: This does not have async
// options right now. In the interest of keeping it minimal, we can add when requirement arises.
type Client interface {
	GetValue(name Key) (interface{}, error)
	GetValueWithConstraints(name Key, constraints map[ConstraintKey]interface{}) (interface{}, error)
}

// NewNopClient creates a new noop client
func NewNopClient() Client {
	return &nopClient{}
}

type nopClient struct{}

func (mc *nopClient) GetValue(name Key) (interface{}, error) {
	return nil, errors.New("unable to find key")
}

func (mc *nopClient) GetValueWithConstraints(
	name Key, constraints map[ConstraintKey]interface{},
) (interface{}, error) {
	return nil, errors.New("unable to find key")
}

// NewCollection creates a new collection
func NewCollection(client Client) *Collection {
	return &Collection{client}
}

// Collection of values that are in dynamic config
type Collection struct {
	client Client
}

// GetIntProperty represents
func (c *Collection) GetIntProperty(key Key, defaultVal int) func() int {
	return func() int {
		val, err := c.client.GetValue(key)
		if err != nil {
			return defaultVal
		}
		return val.(int)
	}
}

// GetFloat64Property represents
func (c *Collection) GetFloat64Property(key Key, defaultVal float64) func() float64 {
	return func() float64 {
		val, err := c.client.GetValue(key)
		if err != nil {
			return defaultVal
		}
		return val.(float64)
	}
}

// GetBoolProperty represents
func (c *Collection) GetBoolProperty(key Key, defaultVal bool) func() bool {
	return func() bool {
		val, err := c.client.GetValue(key)
		if err != nil {
			return defaultVal
		}
		return val.(bool)
	}
}
