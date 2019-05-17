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

package persistence

import (
	"fmt"
	"github.com/uber/cadence/.gen/go/shared"
	"reflect"
)

// NewVersionHistory initializes new version history
func NewVersionHistory(items []*shared.VersionHistoryItem) shared.VersionHistory {
	if len(items) == 0 {
		panic("version history items cannot be empty")
	}

	return shared.VersionHistory{
		History: items,
	}
}

// UpdateVersionHistory updates the versionHistory slice
func UpdateVersionHistory(current *shared.VersionHistory, item shared.VersionHistoryItem) error {
	currentItem := item
	lastItem := current.GetHistory()[len(current.History)-1]
	if *currentItem.Version < *lastItem.Version {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot update version history with a lower version %v. Last version: %v",
				currentItem.Version,
				lastItem.Version),
		}
	}

	if *currentItem.EventID <= *lastItem.EventID {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot add version history with a lower event id %v. Last event id: %v",
				currentItem.EventID,
				lastItem.EventID),
		}
	}

	if *currentItem.Version > *lastItem.Version {
		// Add a new history
		current.History = append(current.History, &currentItem)
	} else {
		// item.version == lastItem.version && item.eventID > lastItem.eventID
		// Update event  id
		lastItem.EventID = currentItem.EventID
	}
	return nil
}

// FindLowestCommonVersionHistoryItem returns the lowest version history item with the same version
func FindLowestCommonVersionHistoryItem(current shared.VersionHistory, remote shared.VersionHistory) (*shared.VersionHistoryItem, error) {
	localIdx := len(current.History) - 1
	remoteIdx := len(remote.History) - 1

	for localIdx >= 0 && remoteIdx >= 0 {
		localVersionItem := current.GetHistory()[localIdx]
		remoteVersionItem := remote.GetHistory()[remoteIdx]
		if *localVersionItem.Version == *remoteVersionItem.Version {
			if *localVersionItem.EventID > *remoteVersionItem.EventID {
				return remoteVersionItem, nil
			}
			return localVersionItem, nil
		} else if *localVersionItem.Version > *remoteVersionItem.Version {
			localIdx--
		} else {
			// localVersionItem.version < remoteVersionItem.version
			remoteIdx--
		}
	}
	return &shared.VersionHistoryItem{}, &shared.BadRequestError{
		Message: fmt.Sprintf("version history is malformed. No joint point found."),
	}
}

// IsAppendable checks if a version history item is appendable
func IsAppendable(current shared.VersionHistory, item shared.VersionHistoryItem) bool {
	return current.History[len(current.History)-1].Equals(&item)
}

// NewVersionHistories initialize new version histories
func NewVersionHistories(histories []*shared.VersionHistory) shared.VersionHistories {
	if len(histories) == 0 {
		panic("version histories cannot be empty")
	}
	return shared.VersionHistories{
		Histories: histories,
	}
}

// FindLowestCommonVersionHistory finds the lowest common version history item among all version histories
func FindLowestCommonVersionHistory(current shared.VersionHistories, history shared.VersionHistory) (*shared.VersionHistoryItem, *shared.VersionHistory, error) {
	var versionHistoryItem *shared.VersionHistoryItem
	var versionHistory *shared.VersionHistory
	for _, localHistory := range current.Histories {
		item, err := FindLowestCommonVersionHistoryItem(*localHistory, history)
		if err != nil {
			return versionHistoryItem, versionHistory, err
		}

		if versionHistoryItem == nil || *item.EventID > *versionHistoryItem.EventID {
			versionHistoryItem = item
			versionHistory = localHistory
		}
	}
	return versionHistoryItem, versionHistory, nil
}

// AddHistory add new history into version histories
// TODO: merge this func with FindLowestCommonVersionHistory
func AddHistory(current *shared.VersionHistories, item shared.VersionHistoryItem, local shared.VersionHistory, remote shared.VersionHistory) error {
	commonItem := item
	if IsAppendable(local, commonItem) {
		//it won't update h.versionHistories
		for idx, history := range current.Histories {
			if reflect.DeepEqual(history, local) {
				current.Histories[idx] = &remote
			}
		}
	} else {
		current.Histories = append(current.Histories, &remote)
	}
	return nil
}
