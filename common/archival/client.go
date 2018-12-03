package archival

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dgryski/go-farm"
	"github.com/uber/cadence/common/blobstore"
	"go.uber.org/cadence/.gen/go/shared"
	"strconv"
	"time"
)

// Client is used to store and retrieve blobs
type Client interface {
	Archive(
		ctx context.Context,
		domainName string,
		domainId string,
		workflowInfo *shared.WorkflowExecutionInfo,
		history *shared.History,
	) error

	GetWorkflowExecutionHistory(
		ctx context.Context,
		request *shared.GetWorkflowExecutionHistoryRequest,
	) (*shared.GetWorkflowExecutionHistoryResponse, error)

	ListClosedWorkflowExecutions(
		ctx context.Context,
		ListRequest *shared.ListClosedWorkflowExecutionsRequest,
	) (*shared.ListClosedWorkflowExecutionsResponse, error)
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
	return c.uploadBlobs(ctx, domainName, domainId, data, bucketName, workflowInfo)
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

	indexBlobPaths := c.constructIndexBlobPaths(workflowInfo)
	visBlobReq := &blobstore.UploadBlobRequest{
		PrefixesListable: true,
		Blob: &blobstore.Blob{
			Size:            int64(0),
			Body:            bytes.NewReader([]byte),
			CompressionType: blobstore.NoCompression,
			Tags:            tags,
		},
	}
	for _, path := range indexBlobPaths {
		if err := c.bClient.UploadBlob(ctx, bucketName, path, visBlobReq); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) constructIndexBlobPaths(workflowInfo *shared.WorkflowExecutionInfo) map[IndexBlobPath]blobstore.BlobPath {
	workflowId := *workflowInfo.Execution.WorkflowId
	runId := *workflowInfo.Execution.RunId
	dayYear, second := parse(time.Now())

	paths := make(map[IndexBlobPath]blobstore.BlobPath)
	paths[WorkflowByTimePath] = blobstore.ConstructBlobPath(dayYear, second, workflowId, runId)
	paths[WorkflowPath] = blobstore.ConstructBlobPath(shardDir(workflowId), shardDir(reverse(workflowId)),
		workflowId, dayYear, second, runId)
	paths[WorkflowTypePath] = blobstore.ConstructBlobPath(workflowInfo.Type.String(), dayYear, second, workflowId, runId)
	paths[WorkflowStatusPath] = blobstore.ConstructBlobPath(workflowInfo.CloseStatus.String(), dayYear, second, workflowId, runId)
	return paths
}

func parse(t time.Time) (string, string) {
	year := t.Year()
	day := time.Now().YearDay()
	return fmt.Sprintf("%v%v%v", day, separator, year), strconv.Itoa(t.Second())
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func shardDir(input string) string {
	bytes := []byte(input)
	dirNum := int(farm.Hash32(bytes)) % indexedDirSizeLimit
	return fmt.Sprintf("%v%v%v", shardDirPrefix, separator, dirNum)
}
