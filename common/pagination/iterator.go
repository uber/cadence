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

type (
	iterator struct {
		entities       []Entity
		entityIndex    int
		pageTokens     []PageToken
		pageTokenIndex int

		nextEntity Entity
		nextError  error

		fetchFn FetchFn
	}
)

// NewIterator constructs a new Iterator
func NewIterator(
	pageTokens []PageToken,
	fetchFn FetchFn,
) Iterator {
	itr := &iterator{
		entities:       nil,
		entityIndex:    0,
		pageTokens:     pageTokens,
		pageTokenIndex: -1,
		fetchFn:        fetchFn,
	}
	itr.advance(true)
	return itr
}

// Next returns the next Entity or error.
// Returning nil, nil is valid if that is what the provided fetch function provided.
func (i *iterator) Next() (Entity, error) {
	entity := i.nextEntity
	error := i.nextError
	i.advance(false)
	return entity, error
}

// HasNext returns true if there is a next element. There is considered to be a next element
// As long as a fatal error has not occurred and the iterator has not reached the end.
func (i *iterator) HasNext() bool {
	return i.nextError == nil
}

func (i *iterator) advance(initialization bool) {
	if !i.HasNext() && !initialization {
		return
	}
	if i.entityIndex < len(i.entities) {
		i.consume()
	} else {
		i.pageTokenIndex++
		if err := i.advanceToNonEmptyPage(); err != nil {
			i.terminate(err)
		} else {
			i.consume()
		}
	}
}

func (i *iterator) advanceToNonEmptyPage() error {
	if i.pageTokenIndex >= len(i.pageTokens) {
		return ErrIteratorFinished
	}
	entities, err := i.fetchFn(i.pageTokens[i.pageTokenIndex])
	if err != nil {
		return err
	}
	if len(entities) != 0 {
		i.entities = entities
		i.entityIndex = 0
		return nil
	}
	i.pageTokenIndex++
	return i.advanceToNonEmptyPage()
}

func (i *iterator) consume() {
	i.nextEntity = i.entities[i.entityIndex]
	i.nextError = nil
	i.entityIndex++
}

func (i *iterator) terminate(err error) {
	i.nextEntity = nil
	i.nextError = err
}
