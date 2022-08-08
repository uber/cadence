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

package cache

import "errors"

type noOpCache struct{}

func (c *noOpCache) Get(key interface{}) interface{} {
	return nil
}

// Put adds an element to the cache, returning the previous element
func (c *noOpCache) Put(key interface{}, value interface{}) interface{} {
	return nil
}

func (c *noOpCache) PutIfNotExist(key interface{}, value interface{}) (interface{}, error) {
	return nil, errors.New("Not supported for noOpCache")
}

func (c *noOpCache) Delete(key interface{}) {}

func (c *noOpCache) Release(key interface{}) {}

func (c *noOpCache) Iterator() Iterator {
	return nil
}

func (c *noOpCache) Size() int {
	return 0
}
