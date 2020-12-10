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
