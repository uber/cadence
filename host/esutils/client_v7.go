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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/require"
)

type (
	v7Client struct {
		client *elastic.Client
	}
)

func newV7Client(url string) (*v7Client, error) {
	esClient, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(128*time.Millisecond, 513*time.Millisecond))),
	)
	return &v7Client{
		client: esClient,
	}, err
}

func (es *v7Client) PutIndexTemplate(t *testing.T, templateConfigFile, templateName string) {
	// This function is used exclusively in tests. Excluding it from security checks.
	// #nosec
	template, err := os.ReadFile(templateConfigFile)
	require.NoError(t, err)
	putTemplate, err := es.client.IndexPutTemplate(templateName).BodyString(string(template)).Do(createContext())
	require.NoError(t, err)
	require.True(t, putTemplate.Acknowledged)
}

func (es *v7Client) CreateIndex(t *testing.T, indexName string) {
	exists, err := es.client.IndexExists(indexName).Do(createContext())
	require.NoError(t, err)

	if exists {
		deleteTestIndex, err := es.client.DeleteIndex(indexName).Do(createContext())
		require.Nil(t, err)
		require.True(t, deleteTestIndex.Acknowledged)
	}

	createTestIndex, err := es.client.CreateIndex(indexName).Do(createContext())
	require.NoError(t, err)
	require.True(t, createTestIndex.Acknowledged)
}

func (es *v7Client) DeleteIndex(t *testing.T, indexName string) {
	deleteTestIndex, err := es.client.DeleteIndex(indexName).Do(createContext())
	require.Nil(t, err)
	require.True(t, deleteTestIndex.Acknowledged)
}

func (es *v7Client) PutMaxResultWindow(indexName string, maxResultWindow int) error {
	_, err := es.client.IndexPutSettings(indexName).
		BodyString(fmt.Sprintf(`{"max_result_window" : %d}`, maxResultWindow)).
		Do(createContext())
	return err
}

func (es *v7Client) GetMaxResultWindow(indexName string) (string, error) {
	settings, err := es.client.IndexGetSettings(indexName).Do(createContext())
	if err != nil {
		return "", err
	}
	return settings[indexName].Settings["index"].(map[string]interface{})["max_result_window"].(string), nil
}
