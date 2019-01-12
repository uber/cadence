package blobstore

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
)

const (
	// NoCompression indicates that no compression was used
	NoCompression = "NoCompression"
	// Gzip indicates that compression in package "compress/gzip" was used
	Gzip = "Gzip"
	// CompressionTypeTag indicates the name of blob metadata tag used to identify compression type
	CompressionTypeTag = "CompressionTypeTag"
	// InUseCompression identifies the compression in being used
	InUseCompression = Gzip
)

// TODO: add unit tests for these utilities
// TODO: plumb these utilities through filestore and test locally
// TODO: I should plumb through metrics around how long these are taking
// TODO: when I implement terrablob impl I should make sure I can also easily running locally with terraman (add task todo)
func Compress(blob *Blob) error {
	if _, ok := blob.Tags[CompressionTypeTag]; ok {
		return errors.New("cannot compress a blob which already specifies CompressionTypeTag")
	}
	blob.Tags[CompressionTypeTag] = InUseCompression
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(blob.Body); err != nil {
		return err
	}
	w.Close()
	blob.Body = b.Bytes()
	return nil
}

func Decompress(blob *Blob) error {
	if _, ok := blob.Tags[CompressionTypeTag]; !ok {
		return errors.New("blob does not include CompressionTypeTag, cannot decompress")
	}
	compressionType := blob.Tags[CompressionTypeTag]
	delete(blob.Tags, CompressionTypeTag)
	switch compressionType {
	case NoCompression:
		return nil
	case Gzip:
		decompressedBody, err := decompressGzipBytes(blob.Body)
		if err != nil {
			return err
		}
		blob.Body = decompressedBody
		return nil
	default:
		return fmt.Errorf("blob has unknown compression type of %v, cannot decompress", compressionType)
	}
}

func decompressGzipBytes(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	decompressedData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	r.Close()
	return decompressedData, nil
}
