// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
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

package cli

import (
	"encoding/json"
	"os"
	"strings"
)

const (
	flushThreshold = 50
)

type (
	// AdminDBCorruptedExecutionBufferedWriter is used to buffer writes to file for entities of type CorruptedExecution
	AdminDBCorruptedExecutionBufferedWriter interface {
		Add(*CorruptedExecution)
		Flush()
	}

	adminDBCorruptedExecutionBufferedWriterImpl struct {
		f *os.File
		entries []*CorruptedExecution
	}

	// AdminDBCheckFailureBufferedWriter is used to buffer writes to file for entities of type ExecutionCheckFailure
	AdminDBCheckFailureBufferedWriter interface {
		Add(*ExecutionCheckFailure)
		Flush()
	}

	adminDBCheckFailureBufferedWriterImpl struct {
		f *os.File
		entries []*ExecutionCheckFailure
	}
)

// NewAdminDBCorruptedExecutionBufferedWriter constructs a new AdminDBCorruptedExecutionBufferedWriter
func NewAdminDBCorruptedExecutionBufferedWriter(f *os.File) AdminDBCorruptedExecutionBufferedWriter {
	return &adminDBCorruptedExecutionBufferedWriterImpl{
		f:       f,
	}
}

// NewAdminDBCheckFailureBufferedWriter constructs a new AdminDBCheckFailureBufferedWriter
func NewAdminDBCheckFailureBufferedWriter(f *os.File) AdminDBCheckFailureBufferedWriter {
	return &adminDBCheckFailureBufferedWriterImpl{
		f:       f,
	}
}

// Add adds a new entity
func (w *adminDBCorruptedExecutionBufferedWriterImpl) Add(e *CorruptedExecution) {
	if len(w.entries) > flushThreshold {
		w.Flush()
	}
	w.entries = append(w.entries, e)
}

// Flush flushes contents to file
func (w *adminDBCorruptedExecutionBufferedWriterImpl) Flush() {
	var builder strings.Builder
	for _, e := range w.entries {
		if err := writeToBuilder(&builder, e); err != nil {
			ErrorAndExit("adminDBCorruptedExecutionBufferedWriterImpl failed to write to builder", err)
		}
	}
	if err := writeBuilderToFile(&builder, w.f); err != nil {
		ErrorAndExit("adminDBCorruptedExecutionBufferedWriterImpl failed to write to file", err)
	}
	w.entries = nil
}

// Add adds a new entity
func (w *adminDBCheckFailureBufferedWriterImpl) Add(e *ExecutionCheckFailure) {
	if len(w.entries) > flushThreshold {
		w.Flush()
	}
	w.entries = append(w.entries, e)
}

// Flush flushes contents to file
func (w *adminDBCheckFailureBufferedWriterImpl) Flush() {
	var builder strings.Builder
	for _, e := range w.entries {
		if err := writeToBuilder(&builder, e); err != nil {
			ErrorAndExit("adminDBCorruptedExecutionBufferedWriterImpl failed to write to builder", err)
		}
	}
	if err := writeBuilderToFile(&builder, w.f); err != nil {
		ErrorAndExit("adminDBCorruptedExecutionBufferedWriterImpl failed to write to file", err)
	}
	w.entries = nil
}

func writeToBuilder(builder *strings.Builder, e interface{}) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	builder.WriteString(string(data))
	builder.WriteString("\r\n")
	return nil
}

func writeBuilderToFile(builder *strings.Builder, f *os.File) error {
	_, err := f.WriteString(builder.String())
	return err
}