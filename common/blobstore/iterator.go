// The MIT License (MIT)
// 
// Copyright (c) 2020 Uber Technologies, Inc.
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

package blobstore

import (
	"bytes"
	"errors"
	"sync"
)

var (
	// ErrIteratorFinished indicates that next was called on an iterator
	// which has already reached the end of its input.
	ErrIteratorFinished = errors.New("iterator has reached end")
)
type (
	// Iterator is used to iterate over entities (represented as byte slices) in a collection of blobs.
	// Iterator is thread safe.
	// Iterator does not make deep copies of in or out data.
	Iterator interface {
		Next() ([]byte, bool, error)
		HasNext() bool
		Tags() map[string]string
	}

	// GetFn fetches blob with given key or error on failure.
	GetFn func(string) (Blob, error)

	iterator struct {
		sync.Mutex

		keys []string
		keysIndex int
		page [][]byte
		pageIndex int
		nextResult []byte
		nextError error
		tags map[string]string
		newTags bool

		separatorToken []byte
		getFn GetFn
	}
)

// NewIterator constructs a new iterator.
func NewIterator(
	keys []string,
	getFn GetFn,
	separatorToken []byte,
) Iterator {
	itr := &iterator{
		keys: keys,
		keysIndex: -1,

		separatorToken: separatorToken,
		getFn: getFn,
	}
	itr.advance()
	return itr
}

// Next returns the next element in the iterator. Returns an error if no elements exist
// or if a non-retryable error occurred. Additionally returns true if a new blob was successfully fetched.
// When a new blob is successfully fetched it means Tags will be updated to reflect tags for new blob.
func (i *iterator) Next() ([]byte, bool, error) {
	i.Lock()
	defer i.Unlock()

	result := i.nextResult
	error := i.nextError
	newTags := i.newTags
	if i.hasNext() {
		i.advance()
	}
	return result, newTags, error
}

// HasNext returns true if there is a next element. If HasNext returns true
// it is guaranteed that Next will return a nil error and non-nil byte slice.
func (i *iterator) HasNext() bool {
	i.Lock()
	defer i.Unlock()

	return i.hasNext()
}

// Tags returns the tags for the current blob which is being iterated over.
// Tags will be updated after any call to Next which returns true.
func (i *iterator) Tags() map[string]string {
	i.Lock()
	defer i.Unlock()

	return i.tags
}

func (i *iterator) advance() {
	// if current page has unconsumed elements than first consume from current page
	if i.pageIndex < len(i.page) {
		i.nextResult = i.page[i.pageIndex]
		i.nextError = nil
		i.pageIndex = i.pageIndex + 1
		i.newTags = false
		return
	}

	// if current page is finished and list of keys is
	// finished then set iterator to finished state
	if i.keysIndex >= len(i.keys) {
		i.nextResult = nil
		i.nextError = ErrIteratorFinished
		i.newTags = false
		return
	}

	// otherwise try to fetch next blob
	i.keysIndex = i.keysIndex + 1
	blob, err := i.getFn(i.keys[i.keysIndex])
	if err != nil {
		i.nextResult = nil
		i.nextError = err
		i.newTags = false
		return
	}

	// if current blob is empty and there are more blobs which can
	// be fetched, continue to fetch until non-empty blob is fetched
	if len(blob.Body) == 0 {
		i.keysIndex = i.keysIndex + 1
	}
	for len(blob.Body) == 0 && i.keysIndex < len(i.keys) {
		blob, err = i.getFn(i.keys[i.keysIndex])
		if err != nil {
			i.nextResult = nil
			i.nextError = err
			i.newTags = false
			return
		}
		if len(blob.Body) != 0 {
			break
		}
		i.keysIndex = i.keysIndex + 1
	}

	if len(blob.Body) == 0 {
		i.nextError = nil
		i.nextError = ErrIteratorFinished
		i.newTags = false
		return
	}

	// at this point a new blob has been fetched
	i.pageIndex = 0
	i.tags = blob.Tags
	i.newTags = true
	i.page = bytes.Split(blob.Body, i.separatorToken)
	i.nextResult = i.page[i.pageIndex]
	i.nextError = nil
	i.pageIndex = i.pageIndex + 1
}

func (i *iterator) hasNext() bool {
	return i.nextResult != nil && i.nextError == nil
}