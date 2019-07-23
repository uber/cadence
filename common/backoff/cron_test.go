// Copyright (c) 2019 Uber Technologies, Inc.
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

package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NextCronSchedule(t *testing.T) {
	a := assert.New(t)

	// every day cron
	now, _ := time.Parse(time.RFC3339, "2018-12-17T08:00:00-08:00") // UTC: 2018-12-17 16:00:00 +0000 UTC
	cronSpec := "0 10 * * *"
	backoff := GetBackoffForNextSchedule(cronSpec, now, now)
	a.Equal(time.Hour*18, backoff)
	nextNow := now.Add(backoff)
	backoff = GetBackoffForNextSchedule(cronSpec, nextNow, nextNow)
	a.Equal(time.Hour*24, backoff)

	// every hour cron
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:08:00+00:00")
	cronSpec = "0 * * * *"
	backoff = GetBackoffForNextSchedule(cronSpec, now, now)
	a.Equal(time.Minute*52, backoff)
	nextNow = now.Add(backoff)
	backoff = GetBackoffForNextSchedule(cronSpec, nextNow, nextNow)
	a.Equal(time.Hour, backoff)

	// every minute cron
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:08:18+00:00")
	cronSpec = "* * * * *"
	backoff = GetBackoffForNextSchedule(cronSpec, now, now)
	a.Equal(time.Second*42, backoff)
	nextNow = now.Add(backoff)
	backoff = GetBackoffForNextSchedule(cronSpec, nextNow, nextNow)
	a.Equal(time.Minute, backoff)

	// close time before next schedule time
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:00:00+00:00")
	end, _ := time.Parse(time.RFC3339, "2018-12-17T09:00:00+00:00")
	cronSpec = "0 10 * * *"
	backoff = GetBackoffForNextSchedule(cronSpec, now, end)
	a.Equal(time.Hour*1, backoff)

	// close time after next schedule time
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:00:00+00:00")
	end, _ = time.Parse(time.RFC3339, "2018-12-20T00:00:00+00:00")
	cronSpec = "0 10 * * *"
	backoff = GetBackoffForNextSchedule(cronSpec, now, end)
	a.Equal(time.Hour*10, backoff)

	now, _ = time.Parse(time.RFC3339, "2018-12-17T00:04:00+00:00")
	end, _ = time.Parse(time.RFC3339, "2018-12-17T01:02:00+00:00")
	cronSpec = "*/10 * * * *"
	backoff = GetBackoffForNextSchedule(cronSpec, now, end)
	assert.Equal(t, time.Minute*8, backoff)

	// invalid cron spec
	cronSpec = "invalid-cron-spec"
	backoff = GetBackoffForNextSchedule(cronSpec, now, now)
	a.Equal(NoBackoff, backoff)
}

func Test_Every(t *testing.T) {
	cronSpec1 := "*/1 * * * *"
	cronSpec2 := "@every 1m"
	now, _ := time.Parse(time.RFC3339, "2018-12-17T08:08:30+00:00")
	duration1 := GetBackoffForNextSchedule(cronSpec1, now, now)
	duration2 := GetBackoffForNextSchedule(cronSpec2, now, now)
	assert.Equal(t, time.Second*30, duration1)
	assert.Equal(t, time.Second*60, duration2)

	// close time before next schedule time
	cronSpec := "@every 5h"
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:00:00+00:00")
	end, _ := time.Parse(time.RFC3339, "2018-12-17T09:00:00+00:00")
	backoff := GetBackoffForNextSchedule(cronSpec, now, end)
	assert.Equal(t, time.Hour*4, backoff)

	// close time after next schedule time
	now, _ = time.Parse(time.RFC3339, "2018-12-17T08:00:00+00:00")
	end, _ = time.Parse(time.RFC3339, "2018-12-18T00:00:00+00:00")
	backoff = GetBackoffForNextSchedule(cronSpec, now, end)
	assert.Equal(t, time.Hour*4, backoff)
}
