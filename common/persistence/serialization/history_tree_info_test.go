package serialization

import (
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
)

func TestHistoryTreeInfoEmpty(t *testing.T) {
	// Create a nil HistoryTree
	var nilInfo *HistoryTreeInfo

	// Use assertions to check the values
	assert.Equal(t, time.Unix(0, 0), nilInfo.GetCreatedTimestamp(), "CreatedTimestamp should be zero time")
	assert.Nil(t, nilInfo.GetAncestors(), "Ancestors should be nil")
	assert.Empty(t, nilInfo.GetInfo(), "Info should be empty string")
}

func TestHistoryTreeInfoFilled(t *testing.T) {
	filledInfo := HistoryTreeInfo{}

	f := fuzz.New()

	f.Fuzz(&filledInfo)

	// Use assertions to check the values
	assert.Equal(t, filledInfo.CreatedTimestamp, filledInfo.GetCreatedTimestamp(), "CreatedTimestamp should match")
	assert.Equal(t, filledInfo.Ancestors, filledInfo.GetAncestors(), "Ancestors should match")
	assert.Equal(t, filledInfo.Info, filledInfo.GetInfo(), "Info should match")
}
