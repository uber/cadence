package sysworkflow

import "github.com/uber-common/bark"

type replayBarkLogger struct {
	logger            bark.Logger
	isReplay          bool
	enableLogInReplay bool
}

// NewReplayBarkLogger creates a bark logger which is aware of cadence's replay mode
func NewReplayBarkLogger(logger bark.Logger, isReplay bool, enableLogInReplay bool) bark.Logger {
	return &replayBarkLogger{
		logger:            logger,
		isReplay:          isReplay,
		enableLogInReplay: enableLogInReplay,
	}
}

// Debug logs at debug level
func (r *replayBarkLogger) Debug(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Debug(args)
}

// Debugf logs at debug level with fmt.Printf-like formatting
func (r *replayBarkLogger) Debugf(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Debugf(format, args)
}

// Info logs at info level
func (r *replayBarkLogger) Info(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Info(args)
}

// Infof logs at info level with fmt.Printf-like formatting
func (r *replayBarkLogger) Infof(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Infof(format, args)
}

// Warn logs at warn level
func (r *replayBarkLogger) Warn(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Warn(args)
}

// Warnf logs at warn level with fmt.Printf-like formatting
func (r *replayBarkLogger) Warnf(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Warnf(format, args)
}

// Error logs at error level
func (r *replayBarkLogger) Error(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Error(args)
}

// Errorf logs at error level with fmt.Printf-like formatting
func (r *replayBarkLogger) Errorf(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Errorf(format, args)
}

// Fatal logs at fatal level, then terminate process (irrecoverable)
func (r *replayBarkLogger) Fatal(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Fatal(args)
}

// Fatalf logs at fatal level with fmt.Printf-like formatting, then terminate process (irrecoverable)
func (r *replayBarkLogger) Fatalf(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Fatalf(format, args)
}

// Panic logs at panic level, then panic (recoverable)
func (r *replayBarkLogger) Panic(args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Panic(args)
}

// Panicf logs at panic level with fmt.Printf-like formatting, then panic (recoverable)
func (r *replayBarkLogger) Panicf(format string, args ...interface{}) {
	if r.isReplay && !r.enableLogInReplay {
		return
	}
	r.logger.Panicf(format, args)
}

// WithField returns a logger with the specified key-value pair set, to be logged in a subsequent normal logging call
func (r *replayBarkLogger) WithField(key string, value interface{}) bark.Logger {
	return &replayBarkLogger{
		logger:            r.logger.WithField(key, value),
		isReplay:          r.isReplay,
		enableLogInReplay: r.enableLogInReplay,
	}
}

// WithFields returns a logger with the specified key-value pairs set, to be included in a subsequent normal logging call
func (r *replayBarkLogger) WithFields(keyValues bark.LogFields) bark.Logger {
	return &replayBarkLogger{
		logger:            r.logger.WithFields(keyValues),
		isReplay:          r.isReplay,
		enableLogInReplay: r.enableLogInReplay,
	}
}

// WithError returns a logger with the specified error set, to be included in a subsequent normal logging call
func (r *replayBarkLogger) WithError(err error) bark.Logger {
	return &replayBarkLogger{
		logger:            r.logger.WithError(err),
		isReplay:          r.isReplay,
		enableLogInReplay: r.enableLogInReplay,
	}
}

// Fields returns map fields associated with this logger, if any (i.e. if this logger was returned from WithField[s])
// If no fields are set, returns nil
func (r *replayBarkLogger) Fields() bark.Fields {
	return r.logger.Fields()
}
