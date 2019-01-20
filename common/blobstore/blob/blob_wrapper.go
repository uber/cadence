package blob

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/uber/cadence/common"
	"io/ioutil"
	"strings"
)

const (
	wrappersTag     = "wrappers"
	encodingKey     = "encoding"
	compressionKey  = "compression"
	jsonEncoding    = "json"
	gzipCompression = "compress/gzip"
)

type (
	// WrapFn adds a single layer to a blob's wrapping; will update wrapper metadata tag and potentially modify body.
	// WrapFn can leave input blob in an invalid state, but will always return an error in such cases.
	WrapFn func(*Blob) error

	// WrappingLayers indicates the values of every wrapping layer of a blob, nil fields indicate that blob was not wrapped with layer.
	WrappingLayers struct {
		EncodingFormat *string
		Compression    *string
	}
)

// JsonEncoded returns a WrapFn used to indicate that at the encoding layer json was used
func JsonEncoded() WrapFn {
	return func(b *Blob) error {
		wrappers := common.StringPtr(b.Tags[wrappersTag])
		if exists(wrappers, encodingKey) {
			return errors.New("encoding layer already specified")
		}
		push(wrappers, encodingKey, jsonEncoding)
		b.Tags[wrappersTag] = *wrappers
		return nil
	}
}

// GzipCompressed returns a WrapFn used to compresses body using gzip and indicates that at the compression layer gzip was used
func GzipCompressed() WrapFn {
	return func(b *Blob) error {
		wrappers := common.StringPtr(b.Tags[wrappersTag])
		if exists(wrappers, compressionKey) {
			return errors.New("compression layer already specified")
		}
		push(wrappers, compressionKey, gzipCompression)
		b.Tags[wrappersTag] = *wrappers
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		if _, err := w.Write(b.Body); err != nil {
			w.Close()
			return err
		}
		// must call close before accessing buf.Bytes()
		w.Close()
		b.Body = buf.Bytes()
		return nil
	}
}

// Wrap returns a deep copy of input blob with all wrapping functions applied. Input blob is not modified.
func Wrap(blob *Blob, functions ...WrapFn) (*Blob, error) {
	if blob == nil {
		return nil, nil
	}
	wrappedBlob := DeepCopy(blob)
	for _, f := range functions {
		if err := f(wrappedBlob); err != nil {
			return nil, err
		}
	}
	return wrappedBlob, nil
}

func Unwrap(blob *Blob) (*Blob, *WrappingLayers, error) {
	wrappingLayers := &WrappingLayers{}
	if blob == nil {
		return nil, wrappingLayers, nil
	}
	unwrappedBlob := DeepCopy(blob)
	wrappers, ok := blob.Tags[wrappersTag]
	if !ok {
		return unwrappedBlob, wrappingLayers, nil
	}
	stack := common.StringPtr(wrappers)
	for len(*stack) != 0 {
		k, v, err := pop(stack)
		if err != nil {
			return nil, nil, err
		}
		switch k {
		case encodingKey:
			wrappingLayers.EncodingFormat = common.StringPtr(v)
		case compressionKey:
			wrappingLayers.Compression = common.StringPtr(v)
			dBody, err := decompress(v, unwrappedBlob.Body)
			if err != nil {
				return nil, nil, err
			}
			unwrappedBlob.Body = dBody
		default:
			return nil, nil, fmt.Errorf("cannot unwrap encountered unknown wrapping layer: %v", k)
		}
	}
	delete(unwrappedBlob.Tags, wrappersTag)
	return unwrappedBlob, wrappingLayers, nil
}

func decompress(compression string, data []byte) ([]byte, error) {
	switch compression {
	case gzipCompression:
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		dBody, err := ioutil.ReadAll(r)
		r.Close()
		return dBody, err
	default:
		return nil, fmt.Errorf("cannot decompress, encountered unknown compression format: %v", compression)
	}
}

/**

IMPORTANT:
The following provides a naive stack implementation that is tightly coupled
to blob_wrapper's usage. In particular the exists function is not robust.
This naive implementation is used here to avoid allocations of many small objects (this is expensive for golang's GC).

*/

const (
	mapSeparator  = ","
	pairSeparator = ":"
)

func push(stack *string, key string, value string) {
	pair := strings.Join([]string{key, value}, pairSeparator)
	*stack = strings.Join([]string{pair, *stack}, mapSeparator)
}

func pop(stack *string) (key string, value string, err error) {
	if len(*stack) == 0 {
		return "", "", errors.New("stack is empty")
	}
	topIndex := strings.Index(*stack, mapSeparator)
	if topIndex == -1 {
		return "", "", fmt.Errorf("stack is malformed: %v", *stack)
	}
	currStack := *stack
	topPair := strings.Split(currStack[:topIndex], pairSeparator)
	if len(topPair) != 2 {
		return "", "", fmt.Errorf("stack is malformed: %v", *stack)
	}
	*stack = currStack[topIndex+1:]
	return topPair[0], topPair[1], nil
}

func exists(stack *string, key string) bool {
	return strings.Contains(*stack, key)
}
