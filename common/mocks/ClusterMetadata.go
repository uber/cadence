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

package mocks

import mock "github.com/stretchr/testify/mock"

// ClusterMetadata is an autogenerated mock type for the Metadata type
type ClusterMetadata struct {
	mock.Mock
}

// ClusterNameForFailoverVersion provides a mock function with given fields:
func (_m *ClusterMetadata) ClusterNameForFailoverVersion(failoverVersion int64) string {
	ret := _m.Called(failoverVersion)

	var r0 string
	if rf, ok := ret.Get(0).(func(int64) string); ok {
		r0 = rf(failoverVersion)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetAllClusterFailoverVersions provides a mock function with given fields:
func (_m *ClusterMetadata) GetAllClusterFailoverVersions() map[string]int64 {
	ret := _m.Called()

	var r0 map[string]int64
	if rf, ok := ret.Get(0).(func() map[string]int64); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]int64)
		}
	}

	return r0
}

// GetCurrentClusterName provides a mock function with given fields:
func (_m *ClusterMetadata) GetCurrentClusterName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetMasterClusterName provides a mock function with given fields:
func (_m *ClusterMetadata) GetMasterClusterName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetNextFailoverVersion provides a mock function with given fields: _a0
func (_m *ClusterMetadata) GetNextFailoverVersion(_a0 string, _a1 int64) int64 {
	ret := _m.Called(_a0, _a1)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, int64) int64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// IsGlobalDomainEnabled provides a mock function with given fields:
func (_m *ClusterMetadata) IsGlobalDomainEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsMasterCluster provides a mock function with given fields:
func (_m *ClusterMetadata) IsMasterCluster() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
