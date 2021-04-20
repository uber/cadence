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

package common

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/yarpc/yarpcerrors"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

func TestIsServiceTransientError_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(100 * time.Millisecond)

	require.False(t, IsServiceTransientError(ctx.Err()))
}

func TestIsServiceTransientError_YARPCDeadlineExceeded(t *testing.T) {
	yarpcErr := yarpcerrors.DeadlineExceededErrorf("yarpc deadline exceeded")
	require.False(t, IsServiceTransientError(yarpcErr))
}

func TestIsServiceTransientError_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.False(t, IsServiceTransientError(ctx.Err()))
}

func TestIsContextTimeoutError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(50 * time.Millisecond)
	require.True(t, IsContextTimeoutError(ctx.Err()))
	require.True(t, IsContextTimeoutError(&types.InternalServiceError{Message: ctx.Err().Error()}))

	yarpcErr := yarpcerrors.DeadlineExceededErrorf("yarpc deadline exceeded")
	require.True(t, IsContextTimeoutError(yarpcErr))

	require.False(t, IsContextTimeoutError(errors.New("some random error")))

	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	require.False(t, IsContextTimeoutError(ctx.Err()))
}

func TestConvertDynamicConfigMapPropertyToIntMap(t *testing.T) {
	dcValue := make(map[string]interface{})
	for idx, value := range []interface{}{int(0), int32(1), int64(2), float64(3.0)} {
		dcValue[strconv.Itoa(idx)] = value
	}

	intMap, err := ConvertDynamicConfigMapPropertyToIntMap(dcValue)
	require.NoError(t, err)
	require.Len(t, intMap, 4)
	for i := 0; i != 4; i++ {
		require.Equal(t, i, intMap[i])
	}
}

func TestCreateHistoryStartWorkflowRequest_ExpirationTimeWithCron(t *testing.T) {
	domainID := uuid.New()
	request := &types.StartWorkflowExecutionRequest{
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    60,
			ExpirationIntervalInSeconds: 60,
		},
		CronSchedule: "@every 300s",
	}
	now := time.Now()
	startRequest, _ := CreateHistoryStartWorkflowRequest(domainID, request, now)
	require.NotNil(t, startRequest)

	expirationTime := startRequest.GetExpirationTimestamp()
	require.NotNil(t, expirationTime)
	require.True(t, time.Unix(0, expirationTime).Sub(now) > 60*time.Second)
}

func TestCreateHistoryStartWorkflowRequest_DelayStart(t *testing.T) {
	domainID := uuid.New()
	request := &types.StartWorkflowExecutionRequest{
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    60,
			ExpirationIntervalInSeconds: 60,
		},
		DelayStartSeconds: Int32Ptr(100),
	}
	now := time.Now()
	startRequest, _ := CreateHistoryStartWorkflowRequest(domainID, request, now)
	require.NotNil(t, startRequest)

	expirationTime := startRequest.GetExpirationTimestamp()
	require.NotNil(t, expirationTime)

	// Since we assign the expiration time after we create the workflow request,
	// There's a chance that the test thread might sleep or get deprioritized and
	// expirationTime - now may not be equal to DelayStartSeconds. Adding 2 seconds
	// buffer to avoid this test being flaky
	require.True(
		t,
		time.Unix(0, expirationTime).Sub(now) >= (100+58)*time.Second,
		"Integration test took too short: %f seconds vs %f seconds",
		time.Duration(time.Unix(0, expirationTime).Sub(now)).Round(time.Millisecond).Seconds(),
		time.Duration((100+58)*time.Second).Round(time.Millisecond).Seconds(),
	)
	require.True(
		t,
		time.Unix(0, expirationTime).Sub(now) < (100+68)*time.Second,
		"Integration test took too long: %f seconds vs %f seconds",
		time.Duration(time.Unix(0, expirationTime).Sub(now)).Round(time.Millisecond).Seconds(),
		time.Duration((100+68)*time.Second).Round(time.Millisecond).Seconds(),
	)
}

func TestCreateHistoryStartWorkflowRequest_ExpirationTimeWithoutCron(t *testing.T) {
	domainID := uuid.New()
	request := &types.StartWorkflowExecutionRequest{
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    60,
			ExpirationIntervalInSeconds: 60,
		},
	}
	now := time.Now()
	startRequest, _ := CreateHistoryStartWorkflowRequest(domainID, request, now)
	require.NotNil(t, startRequest)

	expirationTime := startRequest.GetExpirationTimestamp()
	require.NotNil(t, expirationTime)
	delta := time.Unix(0, expirationTime).Sub(now)
	require.True(t, delta > 58*time.Second)
	require.True(t, delta < 62*time.Second)
}

func TestConvertIndexedValueTypeToThriftType(t *testing.T) {
	expected := workflow.IndexedValueType_Values()
	for i := 0; i < len(expected); i++ {
		require.Equal(t, expected[i], ConvertIndexedValueTypeToThriftType(i, nil))
		require.Equal(t, expected[i], ConvertIndexedValueTypeToThriftType(float64(i), nil))
	}
}

func TestValidateDomainUUID(t *testing.T) {
	testCases := []struct {
		msg        string
		domainUUID string
		valid      bool
	}{
		{
			msg:        "empty",
			domainUUID: "",
			valid:      false,
		},
		{
			msg:        "invalid",
			domainUUID: "some random uuid",
			valid:      false,
		},
		{
			msg:        "valid",
			domainUUID: uuid.New(),
			valid:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := ValidateDomainUUID(tc.domainUUID)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestVersionGreaterThanOrEqualTo(t *testing.T) {
	greater := false
	err := error(nil)

	greater, err = VersionGreaterThanOrEqualTo("1.7.0", "1.6.0")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.7.0", "1.6.1231")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.7.1213", "1.6.13123")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.7.2", "1.7.1")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("2.7.2", "2.7.1")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.7.0", "1.7.0")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.7.2", "1.7.2")
	require.True(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("2.7.2", "2.7.2")
	require.True(t, greater)
	require.Nil(t, err)

	greater, err = VersionGreaterThanOrEqualTo("1.6.0", "1.7.0")
	require.False(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.6.12313", "1.7.0")
	require.False(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.6.12312", "1.7.12313")
	require.False(t, greater)
	require.Nil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.9999.0", "2.0.0")
	require.False(t, greater)
	require.Nil(t, err)

	greater, err = VersionGreaterThanOrEqualTo("", "2.0.0")
	require.False(t, greater)
	require.NotNil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.2.0", "")
	require.False(t, greater)
	require.NotNil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("", "")
	require.False(t, greater)
	require.NotNil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("1.2.0", "absd")
	require.False(t, greater)
	require.NotNil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("abasd", "2.0.0")
	require.False(t, greater)
	require.NotNil(t, err)
	greater, err = VersionGreaterThanOrEqualTo("asdas", "adsfasd")
	require.False(t, greater)
	require.NotNil(t, err)
}
