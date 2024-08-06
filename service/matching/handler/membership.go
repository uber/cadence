package handler

import (
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/service"
	"time"
)

func (e *matchingEngineImpl) subscribeToMembershipChanges() {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("matching membership watcher changes caused a panic, recovering", tag.Dynamic("recovered-panic", r))
		}
	}()

	listener := make(chan *membership.ChangedEvent)
	e.membershipResolver.Subscribe(service.Matching, "matching-engine", listener)

	for {
		select {
		case event := <-listener:
			// There is a race between the event being acted upon in the hashring
			// and this here. In lieu if a better solution, wait a moment to ensure the hashring
			// is stable before attempting to clear out disused tasklists
			time.Sleep(time.Millisecond * 200)
			err := e.shutDownNonOwnedTasklists()
			if err != nil {
				e.logger.Error("Error while trying to determine if tasklists have been shutdown",
					tag.Error(err),
					tag.MembershipChangeEvent(event),
				)
			}
		case <-e.shutdown:
			return
		}
	}
}
