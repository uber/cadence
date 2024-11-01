package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/tools/cli/clitest"
	"github.com/urfave/cli/v2"
)

func TestAdminDBScan(t *testing.T) {

	cases := []struct {
		name        string
		testSetup   func(td *cliTestData) *cli.Context
		errContains string // empty if no error is expected
	}{
		{
			name: "scan type not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app)
			},
			errContains: "unknown scan type",
		},
		{
			name: "unknown scan type provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "some_unknown_scan_type"),
				)
			},
			errContains: "unknown scan type",
		},
		{
			name: "number of shards not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "invariant collection not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 1),
				)
			},
			errContains: "no invariants for scan type \"ConcreteExecutionType\" and collections []",
		},
		{
			name: "invalid invariant collection provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 1),
					clitest.StringSliceArgument("invariant_collection", "some_unknown_invariant_collection"),
				)
			},
			errContains: "unknown invariant collection: some_unknown_invariant_collection",
		},
		{
			name: "input file not found",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 1),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input.json"),
				)
			},
			errContains: "Input file not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			td := newCLITestData(t)
			testCtx := tc.testSetup(td)
			err := AdminDBScan(testCtx)
			if tc.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
