package xdc

import (
	"github.com/uber/cadence/common"
)

const (
	emptyCommonAncestor = -1
)
type (
	eventNode struct {
		eventID int64
		version int64
	}

	eventBranch struct {
		branchToken []byte
		versionHistory []eventNode
	}

	eventReplicationMetadata struct {
		eventsBranches [] eventBranch
	}
)

//Not thread-safe
func newEventBranch(token []byte, eventNodes ...eventNode) *eventBranch {
	return &eventBranch{
		branchToken: token,
		versionHistory: eventNodes,
	}
}

// add is to add new node into the versionHistory slice
func (e *eventBranch) add(node eventNode) {
	if node.eventID <= e.getLastNode().eventID || node.version <= e.getLastNode().version {
		// Or panic?
		return
	}
	e.versionHistory = append(e.versionHistory, node)
}

// updateEventID updates the last event id
func (e *eventBranch) updateEventID(eventID int64) {
	if eventID <= e.getLastNode().eventID {
		// Or panic?
		return
	}
	e.getLastNode().eventID = eventID
}

// append is to add or update the versionHistory slice
func (e *eventBranch) append(node eventNode) {
	if node.version < e.getLastNode().version {
		// Or panic?
		return
	}

	if node.version > e.getLastNode().version {
		e.add(node)
	} else {
		e.updateEventID(node.eventID)
	}
}

func (e *eventBranch) getLastNode() eventNode {
	return e.versionHistory[len(e.versionHistory)-1]
}

func newEventReplicationMetadata(branches ...eventBranch) *eventReplicationMetadata {
	return &eventReplicationMetadata {
		eventsBranches: branches,
	}
}

func (m *eventReplicationMetadata) merge(branch eventBranch) {
	eventID, branchIdx := m.findBranchWithMaxEventID(branch)

	if m.eventsBranches[branchIdx].versionHistory[len(m.eventsBranches[branchIdx].versionHistory) - 1].eventID == eventID {
		for _, event := range branch.versionHistory {
			if event.eventID > eventID {
				m.eventsBranches[branchIdx].append(event)
			}
		}
	} else {
		m.eventsBranches = append(m.eventsBranches, branch)
	}
}


//Lowest common ancestor among all branches
func (m *eventReplicationMetadata) findBranchWithMaxEventID(branch eventBranch) (eventID int64, branchIndex int) {
	for i := 0; i < len(m.eventsBranches); i++ {
		eID := m.findMinEventIDWithMaxSameVersion(m.eventsBranches[i], branch)
		if eID < 0 {
			return emptyCommonAncestor, emptyCommonAncestor
		}
		if eID > eventID {
			eventID = eID
			branchIndex = i
		}
	}
	return
}

func (m *eventReplicationMetadata) findMinEventIDWithMaxSameVersion(local, remote eventBranch) int64 {
	localIdx := len(local.versionHistory) - 1
	remoteIdx := len(remote.versionHistory) - 1

	for localIdx >= 0 && remoteIdx >= 0 {
		localNode := local.versionHistory[localIdx]
		remoteNode := remote.versionHistory[remoteIdx]
		if localNode.version == remoteNode.version {
			return common.MinInt64(localNode.eventID, remoteNode.eventID)
		} else if localNode.version > remoteNode.version {
			localIdx--
		} else {
			remoteIdx--
		}
	}
	return emptyCommonAncestor
}






