package cli

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestAdminDBClean_errorCases(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(app *cli.App) *cli.Context
		inputFileData  string // Simulate the content of the input file
		expectedOutput string
		expectedError  string
	}{
		{
			name: "MissingRequiredFlagScanType",
			setupContext: func(app *cli.App) *cli.Context {
				set := flag.NewFlagSet("test", 0)
				set.String(FlagInputFile, "", "Input file flag")
				// Missing FlagScanType
				_ = set.Set(FlagInputFile, "input.json")
				return cli.NewContext(app, set, nil)
			},
			inputFileData:  ``,
			expectedOutput: "",
			expectedError:  "Required flag not found:",
		},
		{
			name: "UnknownScanType",
			setupContext: func(app *cli.App) *cli.Context {
				set := flag.NewFlagSet("test", 0)
				set.String(FlagScanType, "", "scan type flag")
				set.String(FlagInputFile, "", "Input file flag")
				_ = set.Set(FlagScanType, "unknown")
				_ = set.Set(FlagInputFile, "input.json")
				return cli.NewContext(app, set, nil)
			},
			inputFileData:  ``,
			expectedOutput: "",
			expectedError:  "unknown scan type",
		},
		{
			name: "InvalidInvariantCollection",
			setupContext: func(app *cli.App) *cli.Context {
				set := flag.NewFlagSet("test", 0)
				// Define FlagScanType and FlagInputFile
				set.String(FlagScanType, "", "scan type flag")
				set.String(FlagInputFile, "", "Input file flag")

				// Simulate the collection slice with multiple collections (including an invalid one)
				set.Var(cli.NewStringSlice("invalid_collection", "history"), FlagInvariantCollection, "invariant collection flag")

				// Set actual values for the flags
				_ = set.Set(FlagScanType, "ConcreteExecutionType")
				_ = set.Set(FlagInputFile, "input.json")
				return cli.NewContext(app, set, nil)
			},
			inputFileData:  ``,
			expectedOutput: "",
			expectedError:  "unknown invariant collection",
		},
		{
			name: "NoInvariantsError",
			setupContext: func(app *cli.App) *cli.Context {
				set := flag.NewFlagSet("test", 0)
				set.String(FlagScanType, "", "scan type flag")
				set.String(FlagInvariantCollection, "", "invariant collection flag")
				set.String(FlagInputFile, "", "Input file flag")
				_ = set.Set(FlagScanType, "ConcreteExecutionType")
				_ = set.Set(FlagInvariantCollection, "invalid_collection") // Collection will trigger no invariants
				_ = set.Set(FlagInputFile, "input.json")
				return cli.NewContext(app, set, nil)
			},
			inputFileData:  `[{"Execution": {"ShardID": 1}}]`, // Simulate the content of input file
			expectedOutput: "",
			expectedError:  "no invariants for scantype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp input file with the test's input data
			inputFile, err := os.CreateTemp("", "test_input_*.json")
			assert.NoError(t, err)
			defer os.Remove(inputFile.Name()) // Clean up after test

			// Write input data to the temp file
			if tt.inputFileData != "" {
				_, err = inputFile.WriteString(tt.inputFileData)
				assert.NoError(t, err)
			}
			inputFile.Close()

			// Create test IO handler to capture output
			ioHandler := &testIOHandler{}

			// Set up the CLI app
			app := NewCliApp(nil, WithIOHandler(ioHandler))

			// Set up the CLI context
			c := tt.setupContext(app)

			// Overwrite the FlagInputFile with the actual temp file path
			_ = c.Set(FlagInputFile, inputFile.Name())

			// Call AdminDBClean and validate output and errors
			err = AdminDBClean(c)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Validate the captured output
			assert.Contains(t, ioHandler.outputBytes.String(), tt.expectedOutput)
		})
	}
}
