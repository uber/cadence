// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package scanner

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

type scannerTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
}

func TestScannerSuite(t *testing.T) {
	suite.Run(t, new(scannerTestSuite))
}

func (s *scannerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *scannerTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *scannerTestSuite) TestShardScannerContext() {
	res := resource.NewTest(s.mockCtrl, metrics.Worker)
	ctx := scannerContext{
		Resource:   res,
		cfg:        Config{},
		tallyScope: tally.NoopScope,
	}

	cfg := getShardScannerContext(ctx, &shardscanner.ScannerConfig{
		ScannerWFTypeName: "test_ScannerWFTypeName",
		FixerWFTypeName:   "test_FixerWFTypeName",
		ScannerHooks: func() *shardscanner.ScannerHooks {
			return nil
		},
	})
	s.Equal(shardscanner.ScannerContextKey("test_ScannerWFTypeName"), cfg.ContextKey)
	s.NotNil(cfg.Config)

}

func (s *scannerTestSuite) TestShardFixerContext() {
	res := resource.NewTest(s.mockCtrl, metrics.Worker)
	ctx := scannerContext{
		Resource:   res,
		cfg:        Config{},
		tallyScope: tally.NoopScope,
	}

	cfg := getShardFixerContext(ctx, &shardscanner.ScannerConfig{
		ScannerWFTypeName: "test_ScannerWFTypeName",
		FixerWFTypeName:   "test_FixerWFTypeName",
		FixerHooks: func() *shardscanner.FixerHooks {
			return nil
		},
	})

	s.Equal(shardscanner.ScannerContextKey("test_FixerWFTypeName"), cfg.ContextKey)
	s.NotNil(cfg.Config)
}
