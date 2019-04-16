package mocks

import "github.com/uber/cadence/common/log"
import "github.com/stretchr/testify/mock"

import "github.com/uber/cadence/common/log/tag"

type Logger struct {
	mock.Mock
}

// Debug provides a mock function with given fields: msg, tags
func (_m *Logger) Debug(msg string, tags ...tag.Tag) {
	_m.Called(msg, tags)
}

// Info provides a mock function with given fields: msg, tags
func (_m *Logger) Info(msg string, tags ...tag.Tag) {
	_m.Called(msg, tags)
}

// Warn provides a mock function with given fields: msg, tags
func (_m *Logger) Warn(msg string, tags ...tag.Tag) {
	_m.Called(msg, tags)
}

// Error provides a mock function with given fields: msg, tags
func (_m *Logger) Error(msg string, tags ...tag.Tag) {
	_m.Called(msg, tags)
}

// Fatal provides a mock function with given fields: msg, tags
func (_m *Logger) Fatal(msg string, tags ...tag.Tag) {
	_m.Called(msg, tags)
}

// WithTags provides a mock function with given fields: tags
func (_m *Logger) WithTags(tags ...tag.Tag) log.Logger {
	ret := _m.Called(tags)

	var r0 log.Logger
	if rf, ok := ret.Get(0).(func(...tag.Tag) log.Logger); ok {
		r0 = rf(tags...)
	} else {
		r0 = ret.Get(0).(log.Logger)
	}

	return r0
}
