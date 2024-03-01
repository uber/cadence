package ratelimited

import "github.com/uber/cadence/common/log/tag"

func (h *historyHandler) allowWfID(domainUUID, workflowID string) bool {
	domainName, err := h.domainCache.GetDomainName(domainUUID)
	if err != nil {
		h.logger.Error("Error when getting domain name", tag.Error(err))
		// Fail open
		return true
	}

	allow := h.workflowIDCache.AllowExternal(domainUUID, workflowID)
	enabled := h.ratelimitExternalPerWorkflowID(domainName)

	return allow || !enabled
}
