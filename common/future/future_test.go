// Copyright (c) 2021 Uber Technologies, Inc.
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

package future

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type (
	futureSuite struct {
		suite.Suite
		*require.Assertions
	}

	testType struct {
		intField    int
		stringField string
		timeField   time.Time
	}
)

func TestFutureSuite(t *testing.T) {
	s := new(futureSuite)
	suite.Run(t, s)
}

func (s *futureSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *futureSuite) TestFutureSetGet() {
	var intVal int
	var listVal []string
	var mapVal map[string]struct{}
	var testTypeVal testType

	testCases := []struct {
		futureValue interface{}
		futureErr   error
		valuePtr    interface{}
		expectErr   bool
	}{
		{
			futureValue: testType{
				intField:    101,
				stringField: "a",
				timeField:   time.Now(),
			},
			valuePtr: &testTypeVal,
		},
		{
			futureValue: int(101),
			valuePtr:    &intVal,
		},
		{
			futureValue: []string{"a", "b", "c"},
			valuePtr:    &listVal,
		},
		{
			futureValue: nil,
			valuePtr:    &listVal,
		},
		{
			futureValue: map[string]struct{}{"a": {}, "b": {}},
			valuePtr:    &mapVal,
		},
		{
			futureValue: map[string]struct{}{"a": {}, "b": {}},
			valuePtr:    &intVal,
			expectErr:   true,
		},
		{
			futureValue: int(101),
			valuePtr:    101,
			expectErr:   true,
		},
		{
			futureValue: nil,
			futureErr:   errors.New("some random value"),
			valuePtr:    &intVal,
			expectErr:   true,
		},
		{
			futureValue: time.Now(),
			valuePtr:    nil,
		},
	}

	for _, tc := range testCases {
		future, settable := NewFuture()
		settable.Set(tc.futureValue, tc.futureErr)

		err := future.Get(context.Background(), tc.valuePtr)
		if tc.expectErr {
			if tc.futureErr != nil {
				s.Equal(tc.futureErr, err)
			}
			s.Error(err)
			continue
		}

		s.NoError(err)
		if tc.valuePtr != nil {
			switch tc.futureValue.(type) {
			case int:
				s.Equal(tc.futureValue, intVal)
			case []string:
				s.Equal(tc.futureValue, listVal)
			case map[string]struct{}:
				s.Equal(tc.futureValue, mapVal)
			case testType:
				s.Equal(tc.futureValue, testTypeVal)
			}
		}
	}
}

func (s *futureSuite) TestFutureGet_ContextErr() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	future, settable := NewFuture()
	err := future.Get(ctx, nil)
	s.Error(err)

	settable.Set("some random value", nil)
	err = future.Get(ctx, nil)
	s.Error(err)
}

func (s *futureSuite) TestFutureIsReady() {
	future, settable := NewFuture()
	s.False(future.IsReady())

	settable.Set(nil, errors.New("some random error"))
	s.True(future.IsReady())
}

func (s *futureSuite) TestFutureDoubleSet() {
	_, settable := NewFuture()
	settable.Set("some random value", nil)

	s.Panics(func() {
		settable.Set(nil, errors.New("some random error"))
	})
}
