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

package serialization

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
)

func TestDomain_empty_struct(t *testing.T) {
	var d *DomainInfo

	assert.Equal(t, "", d.GetName())
	assert.Equal(t, "", d.GetDescription())
	assert.Equal(t, "", d.GetOwner())
	assert.Equal(t, int32(0), d.GetStatus())
	assert.Equal(t, time.Duration(0), d.GetRetention())
	assert.Equal(t, false, d.GetEmitMetric())
	assert.Equal(t, "", d.GetArchivalBucket())
	assert.Equal(t, int16(0), d.GetArchivalStatus())
	assert.Equal(t, int64(0), d.GetConfigVersion())
	assert.Equal(t, int64(0), d.GetFailoverVersion())
	assert.Equal(t, int64(0), d.GetNotificationVersion())
	assert.Equal(t, int64(0), d.GetFailoverNotificationVersion())
	assert.Equal(t, int64(0), d.GetPreviousFailoverVersion())
	assert.Equal(t, "", d.GetActiveClusterName())
	assert.Equal(t, emptySlice[string](), d.GetClusters())
	assert.Equal(t, emptyMap[string, string](), d.GetData())
	assert.Equal(t, emptySlice[uint8](), d.GetBadBinaries())
	assert.Equal(t, "", d.GetBadBinariesEncoding())
	assert.Equal(t, int16(0), d.GetHistoryArchivalStatus())
	assert.Equal(t, "", d.GetHistoryArchivalURI())
	assert.Equal(t, int16(0), d.GetVisibilityArchivalStatus())
	assert.Equal(t, "", d.GetVisibilityArchivalURI())
	assert.Equal(t, time.Unix(0, 0), d.GetFailoverEndTimestamp())
	assert.Equal(t, time.Unix(0, 0), d.GetLastUpdatedTimestamp())
}

func TestDomain_non_empty(t *testing.T) {
	now := time.Now()

	d := DomainInfo{
		Name:                        "test_name",
		Description:                 "test_description",
		Owner:                       "test_owner",
		Status:                      1,
		Retention:                   2,
		EmitMetric:                  true,
		ArchivalBucket:              "test_bucket",
		ArchivalStatus:              3,
		ConfigVersion:               4,
		NotificationVersion:         5,
		FailoverNotificationVersion: 6,
		FailoverVersion:             7,
		ActiveClusterName:           "test_cluster_name",
		Clusters:                    []string{"cluster1", "cluster2"},
		Data:                        map[string]string{"key1": "val1", "key2": "val2"},
		BadBinaries:                 []byte{1, 2, 3},
		BadBinariesEncoding:         "test_encoding",
		HistoryArchivalStatus:       8,
		HistoryArchivalURI:          "test_archival_uri",
		VisibilityArchivalStatus:    9,
		VisibilityArchivalURI:       "test_visibility_uri",
		FailoverEndTimestamp:        common.TimePtr(now),
		PreviousFailoverVersion:     10,
		LastUpdatedTimestamp:        now.Add(-1 * time.Second),
		IsolationGroups:             []byte{1, 2, 3},
		IsolationGroupsEncoding:     "test_isolation_encoding",
	}
	assert.Equal(t, d.Name, d.GetName())
	assert.Equal(t, d.Description, d.GetDescription())
	assert.Equal(t, d.Owner, d.GetOwner())
	assert.Equal(t, d.Status, d.GetStatus())
	assert.Equal(t, d.Retention, d.GetRetention())
	assert.Equal(t, d.EmitMetric, d.GetEmitMetric())
	assert.Equal(t, d.ArchivalBucket, d.GetArchivalBucket())
	assert.Equal(t, d.ArchivalStatus, d.GetArchivalStatus())
	assert.Equal(t, d.ConfigVersion, d.GetConfigVersion())
	assert.Equal(t, d.FailoverVersion, d.GetFailoverVersion())
	assert.Equal(t, d.NotificationVersion, d.GetNotificationVersion())
	assert.Equal(t, d.FailoverNotificationVersion, d.GetFailoverNotificationVersion())
	assert.Equal(t, d.PreviousFailoverVersion, d.GetPreviousFailoverVersion())
	assert.Equal(t, d.ActiveClusterName, d.GetActiveClusterName())
	assert.Equal(t, d.Clusters, d.GetClusters())
	assert.Equal(t, d.Data, d.GetData())
	assert.Equal(t, d.BadBinaries, d.GetBadBinaries())
	assert.Equal(t, d.BadBinariesEncoding, d.GetBadBinariesEncoding())
	assert.Equal(t, d.HistoryArchivalStatus, d.GetHistoryArchivalStatus())
	assert.Equal(t, d.HistoryArchivalURI, d.GetHistoryArchivalURI())
	assert.Equal(t, d.VisibilityArchivalStatus, d.GetVisibilityArchivalStatus())
	assert.Equal(t, d.VisibilityArchivalURI, d.GetVisibilityArchivalURI())
	assert.Equal(t, *d.FailoverEndTimestamp, d.GetFailoverEndTimestamp())
	assert.Equal(t, d.LastUpdatedTimestamp, d.GetLastUpdatedTimestamp())
}
