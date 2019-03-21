package logging

import (
	"fmt"
	"runtime/debug"

	"github.com/uber-common/bark"
)

var defaultPanicError = fmt.Errorf("panic object is not error")

// This function is used to capture panic, it will log the panic and also return the error through pointer.
// If the panic value is not error then a default error is returned
// We have to use pointer is because in golang: "recover return nil if was not called directly by a deferred function."
// And we have to set the returned error otherwise our handler will return nil as error which is incorrect
func CapturePanic(logger bark.Logger, retError *error) {
	if errPanic := recover(); errPanic != nil {
		err, ok := errPanic.(error)
		if !ok {
			err = defaultPanicError
		}

		st := string(debug.Stack())

		logger.WithFields(bark.Fields{
			TagStackTrace: st,
			TagPanicError: err.Error(),
		}).Errorf("Panic is captured")

		*retError = err
	}
}
