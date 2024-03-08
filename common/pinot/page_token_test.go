package pinot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializePageToken(t *testing.T) {
	token := &PinotVisibilityPageToken{
		From: 1,
	}

	tests := map[string]struct {
		token          *PinotVisibilityPageToken
		expectedOutput []byte
		expectedError  error
	}{
		"Case2: normal case with nil response": {
			token:          token,
			expectedOutput: []byte(`{"From":1}`),
			expectedError:  nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualOutput, err := SerializePageToken(test.token)
			assert.Equal(t, test.expectedOutput, actualOutput)
			assert.Nil(t, err)
		})
	}
}
