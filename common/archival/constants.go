package archival

type IndexBlobPath int

const (
	WorkflowByTimePath IndexBlobPath = iota
	WorkflowPath
	WorkflowTypePath
	WorkflowStatusPath
)

const (
	workflowIdTag     = "workflowId"
	runIdTag          = "runId"
	workflowTypeTag   = "workflowType"
	workflowStatusTag = "workflowStatus"
	domainNameTag     = "domainName"
	domainIdTag       = "domainId"
	closeTimeTag      = "closeTime"
	startTimeTag      = "startTime"
	historyLengthTag  = "historyLength"

	separator           = "_"
	indexedDirSizeLimit = 100000
	shardDirPrefix      = "shard"
)
