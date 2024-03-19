// Copyright (c) 2021 Uber Technologies Inc.
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

package proto

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/yarpc/yarpcerrors"

	"github.com/uber/cadence/common/types/testdata"
)

func TestErrors(t *testing.T) {
	for _, err := range []error{
		&testdata.AccessDeniedError,
		&testdata.BadRequestError,
		&testdata.CancellationAlreadyRequestedError,
		&testdata.ClientVersionNotSupportedError,
		&testdata.CurrentBranchChangedError,
		&testdata.DomainAlreadyExistsError,
		&testdata.DomainNotActiveError,
		&testdata.EntityNotExistsError,
		&testdata.WorkflowExecutionAlreadyCompletedError,
		&testdata.EventAlreadyStartedError,
		&testdata.InternalDataInconsistencyError,
		&testdata.InternalServiceError,
		&testdata.LimitExceededError,
		&testdata.QueryFailedError,
		&testdata.RemoteSyncMatchedError,
		&testdata.RetryTaskV2Error,
		&testdata.ServiceBusyError,
		&testdata.ShardOwnershipLostError,
		&testdata.WorkflowExecutionAlreadyStartedError,
		&testdata.StickyWorkerUnavailableError,
	} {
		name := reflect.TypeOf(err).Elem().Name()
		t.Run(name, func(t *testing.T) {
			// Check that all the fields are set in the test error
			assert.True(t, checkAllIsSet(err))

			// Test that the mappings does not lose information
			assert.Equal(t, err, ToError(FromError(err)))
		})
	}
}

func checkAllIsSet(err error) bool {
	// All the errors are pointers, so we get the value with .Elem
	errValue := reflect.ValueOf(err).Elem()

	for i := 0; i < errValue.NumField(); i++ {
		field := errValue.Field(i)

		// IsZero checks if the value is the default value (e.g. nil, "", 0 etc)
		if field.IsZero() {
			return false
		}
	}

	return true
}

func TestNilMapping(t *testing.T) {
	protobufNoError := FromError(nil)
	assert.Equal(t, yarpcerrors.CodeOK, yarpcerrors.FromError(protobufNoError).Code())
	assert.Nil(t, ToError(protobufNoError))
}

func TestUnknownError(t *testing.T) {
	err := errors.New("unknown error")
	protobufErr := FromError(err)
	assert.True(t, yarpcerrors.IsUnknown(protobufErr))
	assert.Equal(t, err, ToError(protobufErr))
}

func TestTimeoutError(t *testing.T) {
	// For unknown errors we return the raw yarpc error
	timeout := yarpcerrors.DeadlineExceededErrorf("timeout")
	assert.Equal(t, timeout, ToError(timeout))
}
