package corruptionchecker

type Checker interface {
	//StaleWorkflowCheck
	//MissingHistoryCheck
	SupiciousWorkflowCheck(workflowID string, domainID string, runID string) error
}

type wfchecker struct {
}

func (w wfchecker) SupiciousWorkflowCheck(workflowID string, domainID string, runID string) error {
	//reuse some components from the Invariant checks
	//delete the validated wfs

	return nil
}
