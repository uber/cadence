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

package v6

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/olivere/elastic"
	esaws "github.com/olivere/elastic/aws/v4"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/elasticsearch/client"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

type (
	// ElasticV6 implements Client
	ElasticV6 struct {
		client *elastic.Client
		logger log.Logger
	}
)

func (c *ElasticV6) IsNotFoundError(err error) bool {
	return elastic.IsNotFound(err)
}

// NewV6Client returns a new implementation of GenericClient
func NewV6Client(
	connectConfig *config.ElasticSearchConfig,
	logger log.Logger,
	tlsClient *http.Client,
) (*ElasticV6, error) {
	clientOptFuncs := []elastic.ClientOptionFunc{
		elastic.SetURL(connectConfig.URL.String()),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(128*time.Millisecond, 513*time.Millisecond))),
		elastic.SetDecoder(&elastic.NumberDecoder{}), // critical to ensure decode of int64 won't lose precise)
	}
	if connectConfig.DisableSniff {
		clientOptFuncs = append(clientOptFuncs, elastic.SetSniff(false))
	}
	if connectConfig.DisableHealthCheck {
		clientOptFuncs = append(clientOptFuncs, elastic.SetHealthcheck(false))
	}
	if connectConfig.AWSSigning.Enable {
		if err := config.CheckAWSSigningConfig(connectConfig.AWSSigning); err != nil {
			return nil, err
		}
		var signingClient *http.Client
		var err error
		if connectConfig.AWSSigning.EnvironmentCredential != nil {
			signingClient, err = buildSigningHTTPClientFromEnvironmentCredentialV6(*connectConfig.AWSSigning.EnvironmentCredential)
		} else {
			signingClient, err = buildSigningHTTPClientFromStaticCredentialV6(*connectConfig.AWSSigning.StaticCredential)
		}
		if err != nil {
			return nil, err
		}
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(signingClient))
	}
	if tlsClient != nil {
		clientOptFuncs = append(clientOptFuncs, elastic.SetHttpClient(tlsClient))
	}

	client, err := elastic.NewClient(clientOptFuncs...)
	if err != nil {
		return nil, err
	}

	return &ElasticV6{
		client: client,
		logger: logger,
	}, nil
}

// Refer to https://github.com/olivere/elastic/blob/release-branch.v6/recipes/aws-connect-v4/main.go
func buildSigningHTTPClientFromStaticCredentialV6(credentialConfig config.AWSStaticCredential) (*http.Client, error) {
	awsCredentials := credentials.NewStaticCredentials(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
		credentialConfig.SessionToken,
	)
	return esaws.NewV4SigningClient(awsCredentials, credentialConfig.Region), nil
}

func buildSigningHTTPClientFromEnvironmentCredentialV6(credentialConfig config.AWSEnvironmentCredential) (*http.Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(credentialConfig.Region)},
	)
	if err != nil {
		return nil, err
	}
	return esaws.NewV4SigningClient(sess.Config.Credentials, credentialConfig.Region), nil
}

func (c *ElasticV6) PutMapping(ctx context.Context, index, body string) error {
	_, err := c.client.PutMapping().Index(index).Type("_doc").BodyString(body).Do(ctx)
	return err
}

func (c *ElasticV6) CreateIndex(ctx context.Context, index string) error {
	_, err := c.client.CreateIndex(index).Do(ctx)
	return err
}

func (c *ElasticV6) Count(ctx context.Context, index, query string) (int64, error) {
	return c.client.Count(index).BodyString(query).Do(ctx)
}

func (c *ElasticV6) ClearScroll(ctx context.Context, scrollID string) error {
	return elastic.NewScrollService(c.client).ScrollId(scrollID).Clear(ctx)
}
func (c *ElasticV6) Scroll(ctx context.Context, index, body, scrollID string) (*client.Response, error) {
	scrollService := elastic.NewScrollService(c.client)
	var esResult *elastic.SearchResult
	var err error

	// we are not returning error immediately here, as result + error combination is possible
	if len(scrollID) == 0 {
		esResult, err = scrollService.Index(index).Body(body).Do(ctx)
	} else {
		esResult, err = scrollService.ScrollId(scrollID).Do(ctx)
	}

	var byteResult [][]byte
	if esResult != nil && esResult.Hits != nil {
		for _, hit := range esResult.Hits.Hits {
			byteResult = append(byteResult, *hit.Source)
		}
	}

	result := &client.Response{
		TookInMillis: esResult.TookInMillis,
		TotalHits:    esResult.TotalHits(),
		Hits:         byteResult,
		ScrollID:     esResult.ScrollId,
	}

	if len(esResult.Aggregations) > 0 {
		result.Aggregations = make(map[string]json.RawMessage, len(esResult.Aggregations))
		for key, agg := range esResult.Aggregations {
			result.Aggregations[key] = *agg
		}
	}

	return result, err
}

func (c *ElasticV6) Search(ctx context.Context, index, body string) (*client.Response, error) {
	esResult, err := c.client.Search(index).Source(body).Do(ctx)
	if err != nil {
		return nil, err
	}

	if esResult.Error != nil {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch Error: %#v", esResult.Error),
		}
	} else if esResult.TimedOut {
		return nil, types.InternalServiceError{
			Message: fmt.Sprintf("ElasticSearch Error: Request timed out: %v ms", esResult.TookInMillis),
		}
	}

	var sort []interface{}
	var hits [][]byte
	if esResult != nil && esResult.Hits != nil {
		for _, h := range esResult.Hits.Hits {
			hits = append(hits, *h.Source)
			sort = h.Sort
		}
	}

	result := &client.Response{
		TookInMillis: esResult.TookInMillis,
		TotalHits:    esResult.TotalHits(),
		Hits:         hits,
		Sort:         sort,
	}

	if len(esResult.Aggregations) > 0 {
		result.Aggregations = make(map[string]json.RawMessage, len(esResult.Aggregations))
		for key, agg := range esResult.Aggregations {
			result.Aggregations[key] = *agg
		}
	}

	return result, nil
}
