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

package history

import (
	"github.com/uber/cadence/common/persistence"
)

func convertPendingActivityInfos(
	inputs map[int64]*persistence.ActivityInfo,
) []*persistence.ActivityInfo {

	outputs := []*persistence.ActivityInfo{}
	for _, item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertUpdateActivityInfos(
	inputs map[*persistence.ActivityInfo]struct{},
) []*persistence.ActivityInfo {

	outputs := []*persistence.ActivityInfo{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertDeleteActivityInfos(
	inputs map[int64]struct{},
) []int64 {

	outputs := []int64{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertSyncActivityInfos(
	activityInfos map[int64]*persistence.ActivityInfo,
	inputs map[int64]struct{},
) []persistence.Task {
	outputs := []persistence.Task{}
	for item := range inputs {
		activityInfo, ok := activityInfos[item]
		if ok {
			outputs = append(outputs, &persistence.SyncActivityTask{
				Version:     activityInfo.Version,
				ScheduledID: activityInfo.ScheduleID,
			})
		}
	}
	return outputs
}

func convertPendingTimerInfos(
	inputs map[string]*persistence.TimerInfo,
) []*persistence.TimerInfo {

	outputs := []*persistence.TimerInfo{}
	for _, item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertUpdateTimerInfos(
	inputs map[*persistence.TimerInfo]struct{},
) []*persistence.TimerInfo {

	outputs := []*persistence.TimerInfo{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertDeleteTimerInfos(
	inputs map[string]struct{},
) []string {

	outputs := []string{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertPendingChildExecutionInfos(
	inputs map[int64]*persistence.ChildExecutionInfo,
) []*persistence.ChildExecutionInfo {

	outputs := []*persistence.ChildExecutionInfo{}
	for _, item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertUpdateChildExecutionInfos(
	inputs map[*persistence.ChildExecutionInfo]struct{},
) []*persistence.ChildExecutionInfo {

	outputs := []*persistence.ChildExecutionInfo{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertPendingRequestCancelInfos(
	inputs map[int64]*persistence.RequestCancelInfo,
) []*persistence.RequestCancelInfo {

	outputs := []*persistence.RequestCancelInfo{}
	for _, item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertUpdateRequestCancelInfos(
	inputs map[*persistence.RequestCancelInfo]struct{},
) []*persistence.RequestCancelInfo {

	outputs := []*persistence.RequestCancelInfo{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertPendingSignalInfos(
	inputs map[int64]*persistence.SignalInfo,
) []*persistence.SignalInfo {

	outputs := []*persistence.SignalInfo{}
	for _, item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertUpdateSignalInfos(
	inputs map[*persistence.SignalInfo]struct{},
) []*persistence.SignalInfo {

	outputs := []*persistence.SignalInfo{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}

func convertSignalRequestedIDs(
	inputs map[string]struct{},
) []string {

	outputs := []string{}
	for item := range inputs {
		outputs = append(outputs, item)
	}
	return outputs
}
