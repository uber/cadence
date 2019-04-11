package log

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/uber/cadence/common/log/tag"
	"go.uber.org/zap"
)

// Logger is our abstraction for logging
// Usage examples:
//  import "github.com/uber/cadence/common/log/tag"
//  1) logger = logger.WithFields(
//          tag.TagWorkflowEventID( 123),
//          tag.TagDomainID("test-domain-id"))
//     logger.Info("hello world")
//  2) logger.Info("hello world",
//          tag.TagWorkflowEventID( 123),
//          tag.TagDomainID("test-domain-id"))
//	   )
//  Note: msg should be static, it is not recommended to use fmt.Sprintf() for msg.
//        Anything dynamic should be tagged.
type Logger interface {
	Debug(msg string, tags ...tag.Tag)
	Info(msg string, tags ...tag.Tag)
	Warn(msg string, tags ...tag.Tag)
	Error(msg string, tags ...tag.Tag)
	Fatal(msg string, tags ...tag.Tag)
	WithFields(tags ...tag.Tag) Logger
}

type loggerImpl struct {
	zapLogger *zap.Logger
	skip      int
}

var _ Logger = (*loggerImpl)(nil)

const skipForDefaultLogger = 3

// NewLogger returns a new logger
func NewLogger(zapLogger *zap.Logger) Logger {
	return &loggerImpl{
		zapLogger: zapLogger,
		skip:      skipForDefaultLogger,
	}
}

func caller(skip int) string {
	_, path, lineno, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v:%v", filepath.Base(path), lineno)
}

var callAtTag = tag.LoggingCallAt("").GetKey()

func (lg *loggerImpl) buildFields(tags []tag.Tag) []zap.Field {
	fs := make([]zap.Field, 0, len(tags)+1)
	fs = append(fs, zap.String(callAtTag, caller(lg.skip)))
	for _, t := range tags {
		var f zap.Field
		switch t.GetValueType() {
		case tag.ValueTypeString:
			f = zap.String(t.GetKey(), t.GetString())
		case tag.ValueTypeInteger:
			f = zap.Int64(t.GetKey(), t.GetInteger())
		case tag.ValueTypeDouble:
			f = zap.Float64(t.GetKey(), t.GetDouble())
		case tag.ValueTypeBool:
			f = zap.Bool(t.GetKey(), t.GetBool())
		case tag.ValueTypeError:
			// NOTE: zap already chosed key for error (error)
			f = zap.Error(t.GetError())
		case tag.ValueTypeDuration:
			f = zap.Duration(t.GetKey(), t.GetDuration())
		case tag.ValueTypeTime:
			f = zap.Time(t.GetKey(), t.GetTime())
		case tag.ValueTypeObject:
			f = zap.String(t.GetKey(), fmt.Sprintf("%v", t.GetObject()))
		default:
			panic("not supported tag type!")
		}
		fs = append(fs, f)
	}
	return fs
}

func (lg *loggerImpl) Debug(msg string, tags ...tag.Tag) {
	fields := lg.buildFields(tags)
	lg.zapLogger.Debug(msg, fields...)
}

func (lg *loggerImpl) Info(msg string, tags ...tag.Tag) {
	fields := lg.buildFields(tags)
	lg.zapLogger.Info(msg, fields...)
}

func (lg *loggerImpl) Warn(msg string, tags ...tag.Tag) {
	fields := lg.buildFields(tags)
	lg.zapLogger.Warn(msg, fields...)
}

func (lg *loggerImpl) Error(msg string, tags ...tag.Tag) {
	fields := lg.buildFields(tags)
	lg.zapLogger.Error(msg, fields...)
}

func (lg *loggerImpl) Fatal(msg string, tags ...tag.Tag) {
	fields := lg.buildFields(tags)
	lg.zapLogger.Fatal(msg, fields...)
}

func (lg *loggerImpl) WithFields(tags ...tag.Tag) Logger {
	fields := lg.buildFields(tags)
	zapLogger := lg.zapLogger.With(fields...)
	return NewLogger(zapLogger)
}
