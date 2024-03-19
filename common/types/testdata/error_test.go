package testdata

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllFieldsSetInTestErrors(t *testing.T) {
	for _, err := range Errors {
		name := reflect.TypeOf(err).Elem().Name()
		t.Run(name, func(t *testing.T) {
			// Test all fields are set in the error
			assert.True(t, checkAllIsSet(err))
		})
	}
}

func checkAllIsSet(err error) bool {
	// All the errors are pointers, so we get the value with .Elem
	errValue := reflect.ValueOf(err).Elem()

	for i := 0; i < errValue.NumField(); i++ {
		field := errValue.Field(i)

		// IsZero checks if the value is the default value (e.g. nil, "", 0 etc)
		if field.IsZero() {
			return false
		}
	}

	return true
}
