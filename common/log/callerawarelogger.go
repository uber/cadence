package log

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/uber/cadence/common/log/tag"
)

const defaultSkipLevel = 3

type SkipLevelAware interface {
	WithIncreasedSkipLevel() Logger
}

func NewCallerAwareLogger(logger Logger) Logger{
	return &callerAwareLogger{
		logger:logger,
		skipLevel: defaultSkipLevel,
	}
}

type callerAwareLogger struct {
	logger    Logger
	skipLevel int
}

func (lg *callerAwareLogger) getCaller() string {
	_, path, lineno, ok := runtime.Caller(lg.skipLevel)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v:%v", filepath.Base(path), lineno)
}

func (lg *callerAwareLogger) Debug(msg string, tags ...tag.Tag) {
	tags = append(tags, tag.CallAt(lg.getCaller()))
	lg.logger.Debug(msg, tags...)
}

func (lg *callerAwareLogger) Info(msg string, tags ...tag.Tag) {
	tags = append(tags, tag.CallAt(lg.getCaller()))
	lg.logger.Info(msg, tags...)
}

func (lg *callerAwareLogger) Warn(msg string, tags ...tag.Tag) {
	tags = append(tags, tag.CallAt(lg.getCaller()))
	lg.logger.Warn(msg, tags...)
}

func (lg *callerAwareLogger) Error(msg string, tags ...tag.Tag) {
	tags = append(tags, tag.CallAt(lg.getCaller()))
	lg.logger.Error(msg, tags...)
}

func (lg *callerAwareLogger) Fatal(msg string, tags ...tag.Tag) {
	tags = append(tags, tag.CallAt(lg.getCaller()))
	lg.logger.Fatal(msg, tags...)
}

func (lg *callerAwareLogger) WithTags(tags ...tag.Tag) Logger {
	return &callerAwareLogger{logger: lg.logger.WithTags(tags...), skipLevel:lg.skipLevel}
}

func (lg *callerAwareLogger) WithIncreasedSkipLevel() Logger {
	return &callerAwareLogger{logger: lg.logger, skipLevel: lg.skipLevel + 1}
}