// Copyright (c) 2020 Uber Technologies, Inc.
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

package esutils

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
)

type (
	// ESClient is ElasicSearch client for running test suite to be implemented in different versions of ES.
	// Those interfaces are only being used by tests so we don't implement in common/elasticsearch pkg.
	ESClient interface {
		PutIndexTemplate(s suite.Suite, templateConfigFile, templateName string)
		CreateIndex(s suite.Suite, indexName string)
		DeleteIndex(s suite.Suite, indexName string)
		PutMaxResultWindow(indexName string, maxResultWindow int) error
		GetMaxResultWindow(indexName string) (string, error)
	}
)

// CreateESClient create ElasticSearch client for test
func CreateESClient(s suite.Suite, url string, version string) ESClient {
	var client ESClient
	var err error
	switch version {
	case "v6":
		client, err = newV6Client(url)
	case "v7":
		client, err = newV7Client(url)
	default:
		s.Fail("not supported ES version")
	}
	s.Require().NoError(err)
	return client
}

func createContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 90*time.Second)
	return ctx
}
