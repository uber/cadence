// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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

// Compress modifies input blob to be in compressed form
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

// Decompress modifies input blob to be in decompressed form
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
