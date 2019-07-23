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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var crontests = []struct {
	cron      string
	startTime string
	endTime   string
	result    time.Duration
}{
	{"0 10 * * *", "2018-12-17T08:00:00-08:00", "", time.Hour * 18},
	{"0 10 * * *", "2018-12-18T02:00:00-08:00", "", time.Hour * 24},
	{"0 * * * *", "2018-12-17T08:08:00+00:00", "", time.Minute * 52},
	{"0 * * * *", "2018-12-17T09:00:00+00:00", "", time.Hour},
	{"* * * * *", "2018-12-17T08:08:18+00:00", "", time.Second * 42},
	{"0 * * * *", "2018-12-17T09:00:00+00:00", "", time.Minute * 60},
	{"0 10 * * *", "2018-12-17T08:00:00+00:00", "2018-12-20T00:00:00+00:00", time.Hour * 10},
	{"0 10 * * *", "2018-12-17T08:00:00+00:00", "2018-12-17T09:00:00+00:00", time.Hour},
	{"*/10 * * * *", "2018-12-17T00:04:00+00:00", "2018-12-17T01:02:00+00:00", time.Minute * 8},
	{"invalid-cron-spec", "2018-12-17T00:04:00+00:00", "2018-12-17T01:02:00+00:00", NoBackoff},
	{"@every 5h", "2018-12-17T08:00:00+00:00", "2018-12-17T09:00:00+00:00", time.Hour * 4},
	{"@every 5h", "2018-12-17T08:00:00+00:00", "2018-12-18T00:00:00+00:00", time.Hour * 4},
}

func TestCron(t *testing.T) {
	for idx, tt := range crontests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			start, _ := time.Parse(time.RFC3339, tt.startTime)
			end := start
			if tt.endTime != "" {
				end, _ = time.Parse(time.RFC3339, tt.endTime)
			}
			backoff := GetBackoffForNextSchedule(tt.cron, start, end)
			assert.Equal(t, tt.result, backoff)
		})
	}
}
