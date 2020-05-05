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

package pagination

import "errors"

// ErrIteratorFinished indicates that Next was called on a finished iterator
var ErrIteratorFinished = errors.New("iterator has reached end")

type (
	// Page contains a PageToken and a list of Entity
	Page struct {
		PageToken PageToken
		Entities  []Entity
	}
	// Entity is a generic type which can be operated on by Iterator and Writer
	Entity interface{}
	// PageToken identifies a page
	PageToken interface{}
)

type (
	// WriteFn writes given page to underlying sink.
	// Returns next PageToken or error
	WriteFn func(Page) (PageToken, error)
	// ShouldFlushFn returns true if given page should be flushed false otherwise.
	ShouldFlushFn func(Page) bool
	// FetchFn fetches the entities and next PageToken for the provided PageToken or error on failure.
	FetchFn func(PageToken) ([]Entity, PageToken, error)
)

type (
	// Iterator is used to get entities from a collection of pages.
	// When HasNext returns true it is guaranteed that Next will not return an error.
	// Once iterator returns an error it will never make progress again and will always return that same error.
	// Iterator is not thread safe and does not make defensive in or out copies.
	Iterator interface {
		Next() (Entity, error)
		HasNext() bool
	}
	// Writer is used to buffer and write entities to underlying store.
	Writer interface {
		Add(Entity) error
		Flush() error
		FlushedPages() []PageToken
		FirstFlushedPage() PageToken
		LastFlushedPage() PageToken
	}
)
