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
		domainID string,
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

// NewClient returns a new archival client
func NewClient(deploymentGroup string, bClient blobstore.Client) Client {
	return &client{
		deploymentGroup: deploymentGroup,
		bClient:         bClient,
	}
}

func (c *client) Archive(
	ctx context.Context,
	domainName string,
	domainID string,
	workflowInfo *shared.WorkflowExecutionInfo,
	history *shared.History,
) error {
	bucketName := blobstore.ConstructBucketName(c.deploymentGroup, domainID)
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
	return c.uploadBlobs(ctx, domainName, domainID, data, bucketName, workflowInfo)
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
	domainID string,
	data []byte,
	bucketName blobstore.BucketName,
	workflowInfo *shared.WorkflowExecutionInfo,
) error {

	workflowID := *workflowInfo.Execution.WorkflowId
	runID := *workflowInfo.Execution.RunId

	tags := map[string]string{
		workflowIDTag:     workflowID,
		runIDTag:          runID,
		workflowTypeTag:   *workflowInfo.Type.Name,
		workflowStatusTag: workflowInfo.CloseStatus.String(),
		domainNameTag:     domainName,
		domainIDTag:       domainID,
		closeTimeTag:      strconv.FormatInt(*workflowInfo.CloseTime, 10),
		startTimeTag:      strconv.FormatInt(*workflowInfo.StartTime, 10),
		historyLengthTag:  strconv.FormatInt(*workflowInfo.HistoryLength, 10),
	}

	historyBlobPath := blobstore.ConstructBlobPath(workflowID, runID)
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

func (c *client) constructIndexBlobPaths(workflowInfo *shared.WorkflowExecutionInfo) map[indexBlobPath]blobstore.BlobPath {
	workflowID := *workflowInfo.Execution.WorkflowId
	runID := *workflowInfo.Execution.RunId
	dayYear, second := parse(time.Now())

	paths := make(map[indexBlobPath]blobstore.BlobPath)
	paths[workflowByTimePath] = blobstore.ConstructBlobPath(dayYear, second, workflowID, runID)
	paths[workflowPath] = blobstore.ConstructBlobPath(shardDir(workflowID), shardDir(reverse(workflowID)),
		workflowID, dayYear, second, runID)
	paths[workflowTypePath] = blobstore.ConstructBlobPath(workflowInfo.Type.String(), dayYear, second, workflowID, runID)
	paths[workflowStatusPath] = blobstore.ConstructBlobPath(workflowInfo.CloseStatus.String(), dayYear, second, workflowID, runID)
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
