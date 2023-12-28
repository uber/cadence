// Copyright (c) 2017 Uber Technologies, Inc.
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

package loggerimpl

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

type loggerImpl struct {
	zapLogger     *zap.Logger
	skip          int
	sampleLocalFn func(int) bool
}

const (
	skipForDefaultLogger = 3
	// we put a default message when it is empty so that the log can be searchable/filterable
	defaultMsgForEmpty = "none"
)

// NewNopLogger returns a no-op logger
func NewNopLogger() log.Logger {
	return &loggerImpl{
		zapLogger: zap.NewNop(),
	}
}

// NewDevelopment returns a logger at debug level and log into STDERR
func NewDevelopment() (log.Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return NewLogger(zapLogger), nil
}

// NewLogger returns a new logger
func NewLogger(zapLogger *zap.Logger) log.Logger {
	return &loggerImpl{
		zapLogger: zapLogger,
		skip:      skipForDefaultLogger,
		sampleLocalFn: func(i int) bool {
			return rand.Intn(i) == 0
		},
	}
}

// NewLoggerWithSampleFunc returns a new logger with sample function.
func NewLoggerWithSampleFunc(zapLogger *zap.Logger, sampleLocalFn func(int) bool) log.Logger {
	return &loggerImpl{
		zapLogger:     zapLogger,
		skip:          skipForDefaultLogger,
		sampleLocalFn: sampleLocalFn,
	}
}

func caller(skip int) string {
	_, path, lineno, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v:%v", filepath.Base(path), lineno)
}

func (lg *loggerImpl) buildFieldsWithCallat(tags []tag.Tag) []zap.Field {
	fs := lg.buildFields(tags)
	fs = append(fs, zap.String(tag.LoggingCallAtKey, caller(lg.skip)))
	return fs
}

func (lg *loggerImpl) buildFields(tags []tag.Tag) []zap.Field {
	fs := make([]zap.Field, 0, len(tags))
	for _, t := range tags {
		f := t.Field()
		if f.Key == "" {
			// ignore empty field(which can be constructed manually)
			continue
		}
		fs = append(fs, f)

		if obj, ok := f.Interface.(zapcore.ObjectMarshaler); ok && f.Type == zapcore.ErrorType {
			fs = append(fs, zap.Object(f.Key+"-details", obj))
		}
	}
	return fs
}

func setDefaultMsg(msg string) string {
	if msg == "" {
		return defaultMsgForEmpty
	}
	return msg
}

func (lg *loggerImpl) Debugf(msg string, args ...any) {
	ce := lg.zapLogger.Check(zap.DebugLevel, setDefaultMsg(fmt.Sprintf(msg, args...)))
	if ce == nil {
		return
	}

	fields := lg.buildFieldsWithCallat(nil)
	ce.Write(fields...)
}

func (lg *loggerImpl) Debug(msg string, tags ...tag.Tag) {
	ce := lg.zapLogger.Check(zap.DebugLevel, setDefaultMsg(msg))
	if ce == nil {
		return
	}

	fields := lg.buildFieldsWithCallat(tags)
	ce.Write(fields...)
}

func (lg *loggerImpl) Info(msg string, tags ...tag.Tag) {
	msg = setDefaultMsg(msg)
	fields := lg.buildFieldsWithCallat(tags)
	lg.zapLogger.Info(msg, fields...)
}

func (lg *loggerImpl) Warn(msg string, tags ...tag.Tag) {
	msg = setDefaultMsg(msg)
	fields := lg.buildFieldsWithCallat(tags)
	lg.zapLogger.Warn(msg, fields...)
}

func (lg *loggerImpl) Error(msg string, tags ...tag.Tag) {
	msg = setDefaultMsg(msg)
	fields := lg.buildFieldsWithCallat(tags)
	lg.zapLogger.Error(msg, fields...)
}

func (lg *loggerImpl) Fatal(msg string, tags ...tag.Tag) {
	msg = setDefaultMsg(msg)
	fields := lg.buildFieldsWithCallat(tags)
	lg.zapLogger.Fatal(msg, fields...)
}

func (lg *loggerImpl) WithTags(tags ...tag.Tag) log.Logger {
	fields := lg.buildFields(tags)
	zapLogger := lg.zapLogger.With(fields...)
	return &loggerImpl{
		zapLogger: zapLogger,
		skip:      lg.skip,
	}
}

func (lg *loggerImpl) SampleInfo(msg string, sampleRate int, tags ...tag.Tag) {
	if lg.sampleLocalFn(sampleRate) {
		msg = setDefaultMsg(msg)
		fields := lg.buildFieldsWithCallat(tags)
		lg.zapLogger.Info(msg, fields...)
	}
}
