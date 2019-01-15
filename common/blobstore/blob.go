package blobstore

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
)
// Blob defines a blob
type Blob struct {
	Body []byte
	Tags map[string]string
}

const (
	// NoCompression indicates that no compression was used
	NoCompression = "none"
	// Gzip indicates that "compress/gzip" package was used
	Gzip = "compression/gzip"

	// Compression is the metadata tag name used to identify a blob's compression
	Compression = "Compression"
	// InUseCompression is the compression being used
	InUseCompression = Gzip
)

var (
	errBlobAlreadyCompressed = fmt.Errorf("cannot compress blob which already specifies %v in metadata", Compression)
	errCompressionTagNotFound = fmt.Errorf("cannot decompress blob, %v metadata tag not found", Compression)
)

// Compress modifies blob to be in compressed form
func (b *Blob) Compress() error {
	if _, ok := b.Tags[Compression]; ok {
		return errBlobAlreadyCompressed
	}
	b.Tags[Compression] = InUseCompression
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b.Body); err != nil {
		// cannot call Close in defer because some bytes may not be written to underlying writer until Close is called
		w.Close()
		return err
	}
	w.Close()
	b.Body = buf.Bytes()
	return nil
}

// Decompress modifies blob to be in decompressed form
func (b *Blob) Decompress() error {
	if _, ok := b.Tags[Compression]; !ok {
		return errCompressionTagNotFound
	}
	compressionUsed := b.Tags[Compression]
	delete(b.Tags, Compression)
	switch compressionUsed {
	case NoCompression:
		return nil
	case Gzip:
		decompressedBody, err := decompressGzipBytes(b.Body)
		if err != nil {
			return err
		}
		b.Body = decompressedBody
		return nil
	default:
		return fmt.Errorf("blob has unknown compression of %v, cannot decompress", compressionUsed)
	}
}

func decompressGzipBytes(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func () {
		reader.Close()
	}()
	decompressedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return decompressedData, nil
}
