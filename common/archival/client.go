package archival

import (
	"bytes"
	"context"
	"errors"
	"github.com/uber/cadence/common/blobstore"
	"go.uber.org/cadence/.gen/go/shared"
	"strconv"
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

	defaultHistoryPageSize = 1000
)

type Client interface {
	// define methods needed for archival
}

type client struct {
	deploymentGroup string
	bClient         blobstore.Client
}

func NewClient(deploymentGroup string, bClient blobstore.Client) Client {
	return &client{
		deploymentGroup: deploymentGroup,
		bClient:         bClient,
	}
}

func (c *client) Archive(
	ctx context.Context,
	domainName string,
	domainId string,
	workflowInfo *shared.WorkflowExecutionInfo,
	history *shared.History,
) error {
	bucketName := blobstore.ConstructBucketName(c.deploymentGroup, domainId)
	ok, err := c.bClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !ok {
		c.bClient.CreateBucket(ctx, bucketName, &blobstore.BucketPolicy{
			// TODO: decide on how we want to do ACLs?
			// TODO: plumb through domain config to control retention policy
		})
	}

	data, err := c.serializeHistory(history)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *shared.GetWorkflowExecutionHistoryRequest,
) (*shared.GetWorkflowExecutionHistoryResponse, error) {

	// TODO: implement once histories are being successfully archived
	return nil, errors.New("not implemented")
}

func (c *client) ListClosedWorkflowExecutions(
	ctx context.Context,
	ListRequest *shared.ListClosedWorkflowExecutionsRequest,
) (*shared.ListClosedWorkflowExecutionsResponse, error) {

	// TODO: implement once histories are being successfully archived
	return nil, errors.New("not implemented")
}

// TODO: this will be implemented in a follow up diff after a design meeting on goals here
func (c *client) serializeHistory(*shared.History) ([]byte, error) {
	return nil, nil
}

func (c *client) uploadBlobs(
	ctx context.Context,
	domainName string,
	domainId string,
	data []byte,
	bucketName blobstore.BucketName,
	workflowInfo *shared.WorkflowExecutionInfo,
) error {

	workflowId := *workflowInfo.Execution.WorkflowId
	runId := *workflowInfo.Execution.RunId

	tags := map[string]string{
		workflowIdTag:     workflowId,
		runIdTag:          runId,
		workflowTypeTag:   *workflowInfo.Type.Name,
		workflowStatusTag: workflowInfo.CloseStatus.String(),
		domainNameTag:     domainName,
		domainIdTag:       domainId,
		closeTimeTag:      strconv.FormatInt(*workflowInfo.CloseTime, 10),
		startTimeTag:      strconv.FormatInt(*workflowInfo.StartTime, 10),
		historyLengthTag:  strconv.FormatInt(*workflowInfo.HistoryLength, 10),
	}

	historyBlobPath := blobstore.ConstructBlobPath(workflowId, runId)
	historyBlobReq := &blobstore.UploadBlobRequest{
		PrefixesListable: false,
		Blob: &blobstore.Blob{
			Size:            int64(len(data)),
			Body:            bytes.NewReader(data),
			CompressionType: blobstore.NoCompression,
			Tags:            tags,
		},
	}
	if err := c.bClient.UploadBlob(ctx, bucketName, historyBlobPath, historyBlobReq); err != nil {
		return err
	}

	return nil
}

//dayYear/minute/workflowID/runID
//shardDir/shardDir/workflowID/dayYear/minute/runID
//workflowType/dayYear/minute/workflowID/runID
//status/dayYear/minute/workflowID/runID
