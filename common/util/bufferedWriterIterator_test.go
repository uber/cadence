package util

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBufferedWriterWithIterator(t *testing.T) {
	blobMap := make(map[string][]byte)
	handleFn := func(data []byte, page int) error {
		key := fmt.Sprintf("key_%v", page)
		blobMap[key] = data
		return nil
	}
	bw := NewBufferedWriter(handleFn, json.Marshal, 100, []byte("\r\n"))
	for i := 0; i < 1000; i++ {
		assert.NoError(t, bw.Add(i))
	}
	assert.NoError(t, bw.Flush())
	lastFlushedPage := bw.LastFlushedPage()
	getFn := func(page int) ([]byte, error) {
		key := fmt.Sprintf("key_%v", page)
		return blobMap[key], nil
	}
	itr := NewIterator(0, lastFlushedPage, getFn, []byte("\r\n"))
	i := 0
	for itr.HasNext() {
		val, err := itr.Next()
		assert.NoError(t, err)
		expectedVal, err := json.Marshal(i)
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, val)
		i++
	}
}
