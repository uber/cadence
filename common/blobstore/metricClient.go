package blobstore

import (
	"context"
	"github.com/uber/cadence/common/blobstore/blob"
	"github.com/uber/cadence/common/metrics"
)

var _ Client = (*metricClient)(nil)

type metricClient struct {
	client Client
	metricsClient metrics.Client
}

// NewMetricClient creates a new instance of Client that emits metrics
func NewMetricClient(client Client, metricsClient metrics.Client) Client {
	return &metricClient{
		client:        client,
		metricsClient: metricsClient,
	}
}

func (c *metricClient) Upload(ctx context.Context, bucket string, key blob.Key, blob *blob.Blob) error {
	c.metricsClient.IncCounter(metrics.BlobstoreClientUploadScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientUploadScope, metrics.CadenceClientLatency)
	err := c.client.Upload(ctx, bucket, key, blob)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientUploadScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) Download(ctx context.Context, bucket string, key blob.Key) (*blob.Blob, error) {
	c.metricsClient.IncCounter(metrics.BlobstoreClientDownloadScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientDownloadScope, metrics.CadenceClientLatency)
	resp, err := c.client.Download(ctx, bucket, key)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientDownloadScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) Exists(ctx context.Context, bucket string, key blob.Key) (bool, error) {
	c.metricsClient.IncCounter(metrics.BlobstoreClientExistsScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientExistsScope, metrics.CadenceClientLatency)
	resp, err := c.client.Exists(ctx, bucket, key)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientExistsScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) Delete(ctx context.Context, bucket string, key blob.Key) (bool, error) {
	c.metricsClient.IncCounter(metrics.BlobstoreClientDeleteScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientDeleteScope, metrics.CadenceClientLatency)
	resp, err := c.client.Delete(ctx, bucket, key)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientDeleteScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) ListByPrefix(ctx context.Context, bucket string, prefix string) ([]blob.Key, error) {
	c.metricsClient.IncCounter(metrics.BlobstoreClientListByPrefixScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientListByPrefixScope, metrics.CadenceClientLatency)
	resp, err := c.client.ListByPrefix(ctx, bucket, prefix)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientListByPrefixScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) BucketMetadata(ctx context.Context, bucket string) (*BucketMetadataResponse, error) {
	c.metricsClient.IncCounter(metrics.BlobstoreClientBucketMetadataScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.BlobstoreClientBucketMetadataScope, metrics.CadenceClientLatency)
	resp, err := c.client.BucketMetadata(ctx, bucket)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.BlobstoreClientBucketMetadataScope, metrics.CadenceClientFailures)
	}
	return resp, err
}
