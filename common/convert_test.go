package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {

	aString := "a string"
	aStruct := struct{}{}
	anInterface := interface{}(nil)

	t.Run("a string should be converted to a pointer of a string", func(t *testing.T) {
		assert.Equal(t, &aString, Ptr(aString))
	})

	t.Run("a struct should be converted to a pointer of a string", func(t *testing.T) {
		assert.Equal(t, &aStruct, Ptr(aStruct))
	})

	t.Run("a interface should be converted to a pointer of a string", func(t *testing.T) {
		assert.Equal(t, &anInterface, Ptr(anInterface))
	})
}
