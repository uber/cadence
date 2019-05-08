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

	// VersionHistory provides operations on version histroy
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
		AddHistory(VersionHistoryItem, *VersionHistory, VersionHistory) error
		FindLowestCommonVersionHistory(VersionHistory) (VersionHistoryItem, *VersionHistory, error)
	}

	versionHistoriesImpl struct {
		versionHistories []VersionHistory
	}
)

// NewVersionHistory initializes new version history
func NewVersionHistory(items ...VersionHistoryItem) VersionHistory {
	return &versionHistoryImpl{
		history: items,
	}
}

// Update updates the versionHistory slice
func (v *versionHistoryImpl) Update(item VersionHistoryItem) error {
	if item.version < v.getLastItem().version {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot update version history with a lower version %v. Current version: %v",
				item.version,
				v.getLastItem().version),
		}
	}

	if item.eventID < v.getLastItem().eventID {
		return &shared.BadRequestError{
			Message: fmt.Sprintf("cannot add version history with a lower event id %v. Latest event id: %v",
				item.eventID,
				v.getLastItem().eventID),
		}
	}

	if item.version > v.getLastItem().version {
		// Add a new history
		v.history = append(v.history, item)
	} else {
		// Update event id
		v.getLastItem().eventID = item.eventID
	}
	return nil
}

// FindLowestCommonVersionHistoryItem returns the lowest version history item with the same version
func (v *versionHistoryImpl) FindLowestCommonVersionHistoryItem(remote VersionHistory) (VersionHistoryItem, error) {
	localIdx := len(v.GetHistory()) - 1
	remoteIdx := len(remote.GetHistory()) - 1

	for localIdx >= 0 && remoteIdx >= 0 {
		localVersionItem := v.GetHistory()[localIdx]
		remoteVersionItem := remote.GetHistory()[remoteIdx]
		if localVersionItem.version == remoteVersionItem.version {
			if localVersionItem.eventID > remoteVersionItem.eventID {
				return remoteVersionItem, nil
			}
			return localVersionItem, nil
		} else if localVersionItem.version > remoteVersionItem.version {
			localIdx--
		} else {
			remoteIdx--
		}
	}
	return VersionHistoryItem{}, &shared.BadRequestError{
		Message: fmt.Sprintf("version history is malformed. No joint point found."),
	}
}

// IsAppendable checks if a version history item is appendable
func (v *versionHistoryImpl) IsAppendable(item VersionHistoryItem) bool {
	return v.getLastItem().eventID == item.eventID && v.getLastItem().version == item.version
}

func (v *versionHistoryImpl) GetHistory() []VersionHistoryItem {
	return v.history
}

func (v *versionHistoryImpl) getLastItem() *VersionHistoryItem {
	return &v.history[len(v.history)-1]
}

// NewVersionHistories initialize new version histories
func NewVersionHistories(histories ...VersionHistory) VersionHistories {
	return &versionHistoriesImpl{
		versionHistories: histories,
	}
}

// FindLowestCommonVersionHistory finds the lowest common version history item among all version histories
func (h *versionHistoriesImpl) FindLowestCommonVersionHistory(history VersionHistory) (VersionHistoryItem, *VersionHistory, error) {
	var hItem VersionHistoryItem
	var vHistory *VersionHistory
	for i := 0; i < len(h.versionHistories); i++ {
		item, err := h.versionHistories[i].FindLowestCommonVersionHistoryItem(history)
		if err != nil {
			return hItem, vHistory, err
		}

		if item.eventID > hItem.eventID {
			hItem = item
			vHistory = &h.versionHistories[i]

		}
	}
	return hItem, vHistory, nil
}

// AddHistory add new history into version histories
// Sample
func (h *versionHistoriesImpl) AddHistory(item VersionHistoryItem, local *VersionHistory, remote VersionHistory) error {
	if (*local).IsAppendable(item) {
		for _, historyItem := range remote.GetHistory() {
			if historyItem.version < item.version {
				continue
			}
			if err := (*local).Update(historyItem); err != nil {
				return err
			}
		}
	}
	h.versionHistories = append(h.versionHistories, remote)
	return nil
}
