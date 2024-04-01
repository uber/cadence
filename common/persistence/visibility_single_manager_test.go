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

package persistence

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log"
)

func TestNewVisibilityManagerImpl(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			assert.NotPanics(t, func() {
				NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			})
		})
	}
}

func TestClose(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			mockVisibilityStore.EXPECT().Close().Return().Times(1)

			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			assert.NotPanics(t, func() {
				visibilityManager.Close()
			})
		})
	}
}

func TestGetName(t *testing.T) {
	tests := map[string]struct {
		name string
	}{
		"Case1: success case": {
			name: "TestNewVisibilityManagerImpl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockVisibilityStore := NewMockVisibilityStore(ctrl)
			mockVisibilityStore.EXPECT().GetName().Return(testTableName).Times(1)

			NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())
			visibilityManager := NewVisibilityManagerImpl(mockVisibilityStore, log.NewNoop())

			assert.NotPanics(t, func() {
				visibilityManager.GetName()
			})
		})
	}
}
