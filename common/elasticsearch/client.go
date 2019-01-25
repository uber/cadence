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

package elasticsearch

import (
	"context"
	"github.com/olivere/elastic"
	"time"
)

// Client is a wrapper around ElasticSearch client library.
//
// It simplifies the interface and enables mocking. We intentionally let implementation details of the elastic library
// bleed through, as the main purpose is testability not abstraction.
type (
	Client interface {
		GetRawClient() *elastic.Client
		Search(ctx context.Context, p *SearchParameters) (*elastic.SearchResult, error)
	}

	// SearchParameters holds all required and optional parameters for executing a search
	SearchParameters struct {
		Index    string
		Query    elastic.Query
		From     int
		PageSize int
		Sorter   []elastic.Sorter
	}

	// elasticWrapper implements Client
	elasticWrapper struct {
		client *elastic.Client // retryingClient will use exponential back off to retry failed rpc
	}
)

var _ Client = (*elasticWrapper)(nil)

// newClient returns a new implementation of Client
func newClient(config *Config) (Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(config.URL.String()),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(128*time.Millisecond, 513*time.Millisecond))),
	)
	if err != nil {
		return nil, err
	}
	return &elasticWrapper{client: client}, nil
}

func (c *elasticWrapper) GetRawClient() *elastic.Client {
	return c.client
}

func (c *elasticWrapper) Search(ctx context.Context, p *SearchParameters) (*elastic.SearchResult, error) {
	searchService := c.client.Search(p.Index).
		Query(p.Query).
		From(p.From).
		SortBy(p.Sorter...)

	if p.PageSize != 0 {
		searchService.Size(p.PageSize)
	}

	return searchService.Do(ctx)
}
