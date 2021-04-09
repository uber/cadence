// Copyright (c) 2019 Uber Technologies, Inc.
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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
)

func TestToString(t *testing.T) {
	var cfg Config
	err := Load("", "../../config", "", &cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.String())
}

func TestFillingDefaultSQLEncodingDecodingTypes(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DataStores: map[string]DataStore{
				"sql": {
					SQL: &SQL{},
				},
			},
		},
	}
	cfg.fillDefaults()
	assert.Equal(t, string(common.EncodingTypeThriftRW), cfg.Persistence.DataStores["sql"].SQL.EncodingType)
	assert.Equal(t, []string{string(common.EncodingTypeThriftRW)}, cfg.Persistence.DataStores["sql"].SQL.DecodingTypes)
}

func TestFillingDefaultRpcName(t *testing.T) {
	cfg := &Config{
		ClusterMetadata: &ClusterMetadata{
			ClusterInformation: map[string]ClusterInformation{
				"clusterA": {},
			},
		},
	}

	cfg.fillDefaults()
	assert.Equal(t, "cadence-frontend", cfg.ClusterMetadata.ClusterInformation["clusterA"].RPCName)
}
