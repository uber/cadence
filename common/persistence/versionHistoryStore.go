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
)

type (
	// VersionHistoryItem contains the event id and the associated version
	VersionHistoryItem struct {
		eventID int64
		version int64
	}

	// VersionHistory provides operations on version history
	VersionHistory interface {
		GetHistory() []VersionHistoryItem
		Update(VersionHistoryItem) error
		FindLowestCommonVersionHistoryItem(VersionHistory) (VersionHistoryItem, error)
		IsAppendable(VersionHistoryItem) bool
	}

	versionHistoryImpl struct {
		history []VersionHistoryItem
	}

	// VersionHistories contains a set of VersionHistory
	VersionHistories interface {
		GetHistories() []VersionHistory
		AddHistory(VersionHistoryItem, VersionHistory, VersionHistory) error
		FindLowestCommonVersionHistory(VersionHistory) (VersionHistoryItem, VersionHistory, error)
	}

	versionHistoriesImpl struct {
		versionHistories []VersionHistory
	}
)

// NewVersionHistory initializes new version history
func NewVersionHistory(items []VersionHistoryItem) VersionHistory {
	if len(items) == 0 {
		panic("version history items cannot be empty")
	}

	history := make([]VersionHistoryItem, len(items))
	copy(history, items)
	return &versionHistoryImpl{
		history: history,
	}
}

// Update updates the versionHistory slice
func (v *versionHistoryImpl) Update(item VersionHistoryItem) error {
	currentItem := item
	lastItem := &v.history[len(v.history)-1]
	if currentItem.version < lastItem.version {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot update version history with a lower version %v. Last version: %v",
				currentItem.version,
				lastItem.version),
		}
	}

	if currentItem.eventID <= lastItem.eventID {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot add version history with a lower event id %v. Last event id: %v",
				currentItem.eventID,
				lastItem.eventID),
		}
	}

	if currentItem.version > lastItem.version {
		// Add a new history
		v.history = append(v.history, currentItem)
	} else {
		// item.version == lastItem.version && item.eventID > lastItem.eventID
		// Update event  id
		lastItem.eventID = currentItem.eventID
	}
	return nil
}

// FindLowestCommonVersionHistoryItem returns the lowest version history item with the same version
func (v *versionHistoryImpl) FindLowestCommonVersionHistoryItem(remote VersionHistory) (VersionHistoryItem, error) {
	if remote == nil {
		return VersionHistoryItem{}, &shared.BadRequestError{
			Message: fmt.Sprintf("version history cannot be nil"),
		}
	}
	localHistory := v.GetHistory()
	remoteHistory := remote.GetHistory()
	localIdx := len(localHistory) - 1
	remoteIdx := len(remoteHistory) - 1

	for localIdx >= 0 && remoteIdx >= 0 {
		localVersionItem := localHistory[localIdx]
		remoteVersionItem := remoteHistory[remoteIdx]
		if localVersionItem.version == remoteVersionItem.version {
			if localVersionItem.eventID > remoteVersionItem.eventID {
				return remoteVersionItem, nil
			}
			return localVersionItem, nil
		} else if localVersionItem.version > remoteVersionItem.version {
			localIdx--
		} else {
			// localVersionItem.version < remoteVersionItem.version
			remoteIdx--
		}
	}
	return VersionHistoryItem{}, &shared.BadRequestError{
		Message: fmt.Sprintf("version history is malformed. No joint point found."),
	}
}

// IsAppendable checks if a version history item is appendable
func (v *versionHistoryImpl) IsAppendable(item VersionHistoryItem) bool {
	return v.history[len(v.history)-1] == item
}

// GetHistory returns version history
func (v *versionHistoryImpl) GetHistory() []VersionHistoryItem {
	history := make([]VersionHistoryItem, len(v.history))
	copy(history, v.history)
	return history
}

// NewVersionHistories initialize new version histories
func NewVersionHistories(histories []VersionHistory) VersionHistories {
	if histories == nil || len(histories) == 0 {
		panic("version histories cannot be empty")
	}
	return &versionHistoriesImpl{
		versionHistories: histories,
	}
}

// FindLowestCommonVersionHistory finds the lowest common version history item among all version histories
func (h *versionHistoriesImpl) FindLowestCommonVersionHistory(history VersionHistory) (VersionHistoryItem, VersionHistory, error) {
	var versionHistoryItem VersionHistoryItem
	var versionHistory VersionHistory
	for _, localHistory := range h.versionHistories {
		item, err := localHistory.FindLowestCommonVersionHistoryItem(history)
		if err != nil {
			return versionHistoryItem, versionHistory, err
		}

		if item.eventID > versionHistoryItem.eventID {
			versionHistoryItem = item
			versionHistory = localHistory
		}
	}
	return versionHistoryItem, versionHistory, nil
}

// GetHistories returns a collection of version history
func (h *versionHistoriesImpl) GetHistories() []VersionHistory {
	return h.versionHistories
}

// AddHistory add new history into version histories
// Sample
func (h *versionHistoriesImpl) AddHistory(item VersionHistoryItem, local VersionHistory, remote VersionHistory) error {
	commonItem := item
	if local.IsAppendable(commonItem) {
		for _, historyItem := range remote.GetHistory() {
			if historyItem.version < commonItem.version {
				continue
			}
			if err := local.Update(historyItem); err != nil {
				return err
			}
		}
	} else {
		h.versionHistories = append(h.versionHistories, remote)
	}
	return nil
}
