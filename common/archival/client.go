package archival

import (
	"code.uber.internal/devexp/cadence-server/go-build/.go/src/gb2/src/github.com/pkg/errors"
	"context"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore"
	"go.uber.org/cadence/.gen/go/shared"
	"strings"
)

const (
	defaultHistoryPageSize = 1000
	separator              = "/"
)

type Client interface {
	// define methods needed for archival
}

type client struct {
	deploymentGroup string
	bClient         blobstore.Client
	fClient         frontend.Client
}

func NewClient(deploymentGroup string, bClient blobstore.Client, fClient frontend.Client) Client {
	return &client{
		deploymentGroup: deploymentGroup,
		bClient:         bClient,
		fClient:         fClient,
	}
}

func (c *client) Archive(ctx context.Context, domainName, domainID, workflowId, runId string) error {
	bucketName := c.bucketName(domainID)
	ok, err := c.bClient.BucketExists(ctx, common.StringPtr(bucketName))
	if err != nil {
		return err
	}
	if !ok {
		// TODO: need to create bucket
	}

	history, err := c.getHistory(ctx, domainName, workflowId, runId)
	if err != nil {
		return err
	}
	data, err := c.serializeHistory(history)
	if err != nil {
		return err
	}

	request := &blobstore.UploadBlobRequest{}
	c.bClient.UploadBlob(ctx)

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

func (c *client) getHistory(ctx context.Context, domain, workflowId, runId string) ([]*shared.HistoryEvent, error) {
	execution := &shared.WorkflowExecution{
		WorkflowId: common.StringPtr(workflowId),
		RunId:      common.StringPtr(runId),
	}
	historyResponse, err := c.fClient.GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{
		Domain:          common.StringPtr(domain),
		Execution:       execution,
		MaximumPageSize: common.Int32Ptr(defaultHistoryPageSize),
	})

	if err != nil {
		return nil, err
	}
	events := historyResponse.History.Events
	for historyResponse.NextPageToken != nil {
		historyResponse, err = c.fClient.GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{
			Domain:        common.StringPtr(domain),
			Execution:     execution,
			NextPageToken: historyResponse.NextPageToken,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, historyResponse.History.Events...)
	}

	return events, nil
}

func (c *client) serializeHistory([]*shared.HistoryEvent) ([]byte, error) {

	// TODO: this will be implemented in a follow up diff after a design meeting on goals here
	return nil, nil
}

func (c *client) bucketName(domainID string) string {
	return strings.Join([]string{c.deploymentGroup, domainID}, separator)
}
