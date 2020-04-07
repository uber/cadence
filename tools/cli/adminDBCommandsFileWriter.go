package cli

import (
	"encoding/json"
	"os"
	"strings"
)

const (
	flushThreshold = 100
)

type (
	// AdminDBCommandFileWriter is used to batch writes to db scan
	AdminDBCommandFileWriter interface {
		AddCorruptedExecution(*CorruptedExecution)
		AddExecutionCheckFailure(*ExecutionCheckFailure)
		Flush()
	}

	adminDBCommandFileWriterImpl struct {
		executionCheckFailureFile *os.File
		corruptedExecutionFile    *os.File
		corruptedExecutions       []*CorruptedExecution
		executionCheckFailures    []*ExecutionCheckFailure
	}
)

// NewAdminDBCommandFileWriter constructs a new AdminDBCommandFileWriter
func NewAdminDBCommandFileWriter(executionCheckFailureFile *os.File, corruptedExecutionFile *os.File) AdminDBCommandFileWriter {
	return &adminDBCommandFileWriterImpl{
		executionCheckFailureFile: executionCheckFailureFile,
		corruptedExecutionFile:    corruptedExecutionFile,
	}
}

// AddCorruptedExecution adds a CorruptedExecution
func (fw *adminDBCommandFileWriterImpl) AddCorruptedExecution(ce *CorruptedExecution) {
	if fw.shouldFlush() {
		fw.Flush()
	}
	fw.corruptedExecutions = append(fw.corruptedExecutions, ce)
}

// AddExecutionCheckFailure adds a ExecutionCheckFailure
func (fw *adminDBCommandFileWriterImpl) AddExecutionCheckFailure(ecf *ExecutionCheckFailure) {
	if fw.shouldFlush() {
		fw.Flush()
	}
	fw.executionCheckFailures = append(fw.executionCheckFailures, ecf)
}

// Flush flushes contents of writer to file
func (fw *adminDBCommandFileWriterImpl) Flush() {
	var checkFailureBuilder strings.Builder
	for _, ecf := range fw.executionCheckFailures {
		if err := fw.writeToBuffer(checkFailureBuilder, ecf); err != nil {
			ErrorAndExit("failed to marshal executionCheckFailures", err)
		}
	}
	if err := fw.writeToFile(checkFailureBuilder, fw.executionCheckFailureFile); err != nil {
		ErrorAndExit("failed to write executionCheckFailureFile", err)
	}
	fw.executionCheckFailures = nil

	var corruptedExecutionsBuilder strings.Builder
	for _, ce := range fw.corruptedExecutions {
		if err := fw.writeToBuffer(corruptedExecutionsBuilder, ce); err != nil {
			ErrorAndExit("failed to marshal corruptedExecutionsBuilder", err)
		}
	}
	if err := fw.writeToFile(corruptedExecutionsBuilder, fw.corruptedExecutionFile); err != nil {
		ErrorAndExit("failed to write corruptedExecutionFile", err)
	}
	fw.corruptedExecutions = nil
}

func (fw *adminDBCommandFileWriterImpl) writeToBuffer(builder strings.Builder, e interface{}) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	builder.WriteString(string(data))
	builder.WriteString("\r\n")
	return nil
}

func (fw *adminDBCommandFileWriterImpl) writeToFile(builder strings.Builder, f *os.File) error {
	_, err := f.WriteString(builder.String())
	return err
}

func (fw *adminDBCommandFileWriterImpl) shouldFlush() bool {
	return len(fw.corruptedExecutions)+len(fw.executionCheckFailures) >= flushThreshold
}
