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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/stretchr/testify/suite"
)

type (
	os2Client struct {
		client *opensearch.Client
	}
)

func newOS2Client(url string) (*os2Client, error) {

	esClient, err := opensearch.NewClient(
		opensearch.Config{
			Addresses:    []string{url},
			MaxRetries:   5,
			RetryBackoff: func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
		},
	)

	return &os2Client{
		client: esClient,
	}, err
}

func (os2 *os2Client) PutIndexTemplate(s suite.Suite, templateConfigFile, templateName string) {
	// This function is used exclusively in tests. Excluding it from security checks.
	// #nosec
	template, err := os.Open(templateConfigFile)
	s.Require().NoError(err)
	req := opensearchapi.IndicesPutIndexTemplateRequest{
		Body: template,
		Name: templateName,
	}

	resp, err := req.Do(createContext(), os2.client)
	s.Require().NoError(err)
	s.Require().False(resp.IsError(), fmt.Sprintf("OS2 error: %s", resp.String()))
	if resp.Body != nil {
		resp.Body.Close()
	}
}

func (os2 *os2Client) CreateIndex(s suite.Suite, indexName string) {
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}
	resp, err := existsReq.Do(createContext(), os2.client)
	s.Require().NoError(err)

	if resp.StatusCode == http.StatusOK {
		deleteReq := opensearchapi.IndicesDeleteRequest{
			Index: []string{indexName},
		}
		resp, err := deleteReq.Do(createContext(), os2.client)
		s.Require().Nil(err)
		s.Require().False(resp.IsError(), fmt.Sprintf("OS2 error: %s", resp.String()))

	}
	if resp.Body != nil {
		resp.Body.Close()
	}

	createReq := opensearchapi.IndicesCreateRequest{
		Index: indexName,
	}

	resp, err = createReq.Do(createContext(), os2.client)

	s.Require().NoError(err)
	s.Require().False(resp.IsError(), fmt.Sprintf("OS2 error: %s", resp.String()))
}

func (os2 *os2Client) DeleteIndex(s suite.Suite, indexName string) {
	deleteReq := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}
	resp, err := deleteReq.Do(createContext(), os2.client)
	s.Require().Nil(err)
	s.Require().False(resp.IsError(), fmt.Sprintf("OS2 error: %s", resp.String()))

	if resp.Body != nil {
		resp.Body.Close()
	}
}

func (os2 *os2Client) PutMaxResultWindow(indexName string, maxResultWindow int) error {

	req := opensearchapi.IndicesPutSettingsRequest{
		Body:  strings.NewReader(fmt.Sprintf(`{"max_result_window" : %d}`, maxResultWindow)),
		Index: []string{indexName},
	}

	_, err := req.Do(createContext(), os2.client)

	return err
}

func (os2 *os2Client) GetMaxResultWindow(indexName string) (string, error) {

	req := opensearchapi.IndicesGetSettingsRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(createContext(), os2.client)

	defer res.Body.Close()

	if res.IsError() {
		return "", errors.New(res.String())
	}

	var data map[string]map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if info, ok := data[indexName]["settings"].(map[string]interface{}); ok {
		return info["index"].(map[string]interface{})["max_result_window"].(string), nil
	}

	return "", fmt.Errorf("no settings for index %q", indexName)

}
