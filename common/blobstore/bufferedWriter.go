package blobstore

import (
	"bytes"
	"encoding/json"
	"sync"
)

type (
	// BufferedWriter is used to buffer entities, construct blobs and write to blobstore.
	// BufferedWriter is thread safe and makes defensive copies in and out.
	// BufferedWriter's state is unchanged whenever any method returns an error.
	BufferedWriter interface {
		AddEntity(interface{}) (bool, error)
		AddTag(string, string)
		Flush() error
		GetKeys() []string
	}

	// PutFn writes blob to blobstore using key which may include current page number.
	// Returns key on success or error on failure.
	PutFn func(Blob, int) (string, error)

	bufferedWriter struct {
		sync.Mutex

		currentBody *bytes.Buffer
		currentTags map[string]string
		currentPage int

		keys []string

		flushThreshold int
		separatorToken []byte
		putFn PutFn
	}
)

// NewBufferedWriter constructs a new BufferedWriter
func NewBufferedWriter(
	putFn PutFn,
	flushThreshold int,
	separatorToken []byte,
	startingPage int,
) BufferedWriter {
	return &bufferedWriter{
		currentBody: &bytes.Buffer{},
		currentTags: make(map[string]string),
		currentPage: startingPage,

		keys: nil,

		flushThreshold: flushThreshold,
		separatorToken: separatorToken,
		putFn: putFn,
	}
}

// AddEntity will add entity to buffer for current blob and flush is threshold is exceeded.
func (bw *bufferedWriter) AddEntity(e interface{}) (bool, error) {
	bw.Lock()
	defer bw.Unlock()

	if err := bw.writeToBody(e); err != nil {
		return false, err
	}
	flushed := false
	if bw.shouldFlush() {
		if err := bw.flush(); err != nil {
			return false, err
		}
		flushed = true
	}

	return flushed, nil
}

// AddTag will add tag to current blob. AddTag never triggers a flush.
func (bw *bufferedWriter) AddTag(key string, value string) {
	bw.Lock()
	defer bw.Unlock()

	bw.currentTags[key] = value
}

// Flush invokes PutFn and advances state of bufferedWriter to next page.
func (bw *bufferedWriter) Flush() error {
	bw.Lock()
	defer bw.Unlock()

	return bw.flush()
}

// GetKeys returns a copy of all keys which were successfully handled by PutFn.
func (bw *bufferedWriter) GetKeys() []string {
	bw.Lock()
	defer bw.Unlock()

	return bw.getKeysCopy()
}

func (bw *bufferedWriter) flush() error {
	currentBlob := bw.constructBlob()
	key, err := bw.putFn(currentBlob, bw.currentPage)
	if err != nil {
		return err
	}
	bw.advancePage()
	bw.keys = append(bw.keys, key)
	return nil
}

func (bw *bufferedWriter) writeToBody(e interface{}) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// write will never return an error, so it can be safely ignored
	bw.currentBody.Write(data)
	bw.currentBody.Write(bw.separatorToken)
	return nil
}

func (bw *bufferedWriter) shouldFlush() bool {
	return bw.currentBody.Len() >= bw.flushThreshold
}

func (bw *bufferedWriter) advancePage() {
	bw.currentBody = &bytes.Buffer{}
	bw.currentTags = make(map[string]string)
	bw.currentPage = bw.currentPage + 1
}

func (bw *bufferedWriter) constructBlob() Blob {
	srcBody := bw.currentBody.Bytes()
	destBody := make([]byte, len(srcBody), len(srcBody))
	for i, b := range srcBody {
		destBody[i] = b
	}
	srcTags := bw.currentTags
	destTags := make(map[string]string, len(srcTags))
	for k, v := range srcTags {
		destTags[k] = v
	}
	return Blob{
		Body: destBody,
		Tags: destTags,
	}
}

func (bw *bufferedWriter) getKeysCopy() []string {
	srcKeys := bw.keys
	destKeys := make([]string, len(srcKeys), len(srcKeys))
	for i, k := range srcKeys {
		destKeys[i] = k
	}
	return destKeys
}