// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package error_injectors

import (
	"math/rand"

	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/persistence"
)

// _randomStubFunc is a stub randomized function that could be overriden in tests.
// It introduces randomness to timeout and unhandled errors, to mimic retriable db isues.
var _randomStubFunc = func() bool {
	// forward the call with 50% chance
	return rand.Intn(2) == 0
}

func shouldForwardCallToPersistence(
	err error,
) bool {
	if err == nil {
		return true
	}

	if err == ErrFakeTimeout || err == errors.ErrFakeUnhandled {
		return _randomStubFunc()
	}

	return false
}

func generateFakeError(
	errorRate float64,
) error {
	randFl := rand.Float64()
	if randFl < errorRate {
		return fakeErrors[rand.Intn(len(fakeErrors))]
	}

	return nil
}

var (
	// ErrFakeTimeout is a fake persistence timeout error.
	ErrFakeTimeout = &persistence.TimeoutError{Msg: "Fake Persistence Timeout Error."}
)

var (
	fakeErrors = []error{
		errors.ErrFakeServiceBusy,
		errors.ErrFakeInternalService,
		ErrFakeTimeout,
		errors.ErrFakeUnhandled,
	}
)

func isFakeError(err error) bool {
	for _, fakeErr := range fakeErrors {
		if err == fakeErr {
			return true
		}
	}

	return false
}

const (
	msgInjectedFakeErr = "Injected fake persistence error"
)
