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

package handler

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/service/matching/config"
)

func TestMembershipSubscriptionShutdown(t *testing.T) {
	assert.NotPanics(t, func() {
		ctrl := gomock.NewController(t)
		// this is nil for the memebership resolver, and it'll panic
		r := resource.NewTest(t, ctrl, 0)
		e := matchingEngineImpl{
			config: &config.Config{
				EnableTasklistOwnershipGuard: func(opts ...dynamicconfig.FilterOption) bool { return true },
			},
			shutdown: make(chan struct{}),
		}

		r.MembershipResolver.EXPECT().Subscribe(service.Matching, "matching-engine", gomock.Any()).Times(1)
		r.MembershipResolver.EXPECT().WhoAmI().Times(1)

		go func() {
			time.Sleep(time.Second)
			close(e.shutdown)
		}()
		e.subscribeToMembershipChanges()
	})
}

func TestMembershipSubscriptionPanicHandling(t *testing.T) {
	assert.NotPanics(t, func() {
		ctrl := gomock.NewController(t)
		// this is nil for the memebership resolver, and it'll panic

		r := resource.NewTest(t, ctrl, 0)
		r.MembershipResolver.EXPECT().Subscribe(service.Matching, "matching-engine", gomock.Any()).DoAndReturn(func(_, _, _ any) {
			panic("a panic has occurred")
		})

		e := matchingEngineImpl{
			membershipResolver: r.MembershipResolver,
			config: &config.Config{
				EnableTasklistOwnershipGuard: func(opts ...dynamicconfig.FilterOption) bool { return true },
			},
			shutdown: make(chan struct{}),
		}

		e.subscribeToMembershipChanges()
	})
}
