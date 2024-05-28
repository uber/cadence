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

package testlogger

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
)

type TestingT interface {
	zaptest.TestingT
	Cleanup(func()) // not currently part of zaptest.TestingT
}

// New is a helper to create new development logger in unit test
func New(t TestingT) log.Logger {
	// test logger that emits all logs (none dropped / sample func never returns false)
	return loggerimpl.NewLogger(NewZap(t), loggerimpl.WithSampleFunc(func(int) bool { return true }))
}

// NewZap makes a new test-oriented logger that prevents bad-lifecycle logs from failing tests.
func NewZap(t TestingT) *zap.Logger {
	/*
		HORRIBLE HACK due to async shutdown, both in our code and in libraries (e.g. gocql):
		normally, logs produced after a test finishes will *intentionally* fail the test and/or cause data to race.

		that's a good thing, it reveals possibly-dangerously-flawed lifecycle management.

		unfortunately some of our code and some libraries do not have good lifecycle management,
		and this cannot easily be patched from the outside.

		so this logger cheats: after a test completes, it logs to stderr rather than TestingT.
		EVERY ONE of these logs is bad and we should not produce them, but it's causing many
		otherwise-useful tests to be flaky, and that's a larger interruption than is useful.
	*/
	logAfterComplete, err := zap.NewDevelopment(
		zap.AddStacktrace(zapcore.DebugLevel), // to help find these logs.
	)
	require.NoError(t, err, "could not build a fallback zap logger")

	replaced := &fallbackTestCore{
		fallback: logAfterComplete.Core(),
		testing:  nil, // populated in WrapCore
	}
	tl := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		replaced.testing = core
		return replaced
	})))

	t.Cleanup(replaced.UseFallback) // switch to fallback before ending the test

	return tl
}

type fallbackTestCore struct {
	fallback  zapcore.Core
	testing   zapcore.Core
	completed atomic.Bool
}

var _ zapcore.Core = (*fallbackTestCore)(nil)

func (f *fallbackTestCore) UseFallback() {
	f.completed.Store(true)
}

func (f *fallbackTestCore) Enabled(level zapcore.Level) bool {
	if f.completed.Load() {
		return f.fallback.Enabled(level)
	}
	return f.testing.Enabled(level)
}

func (f *fallbackTestCore) With(fields []zapcore.Field) zapcore.Core {
	if f.completed.Load() {
		return f.fallback.With(fields)
	}
	return f.testing.With(fields)
}

func (f *fallbackTestCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if f.completed.Load() {
		return f.fallback.Check(entry, checked)
	}
	return f.testing.Check(entry, checked)
}

func (f *fallbackTestCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if f.completed.Load() {
		return f.fallback.Write(entry, fields)
	}
	return f.testing.Write(entry, fields)
}

func (f *fallbackTestCore) Sync() error {
	if f.completed.Load() {
		return f.fallback.Sync()
	}
	return f.testing.Sync()
}

var _ zapcore.Core = (*fallbackTestCore)(nil)
