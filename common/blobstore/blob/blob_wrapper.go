package blob

import (
	"github.com/uber/cadence/common"
	"strings"
	"errors"
	"fmt"
)

const (
	wrappersTag = "wrappers"
	encodingKey = "encoding"
	compressionKey = "compression"
	jsonEncoding = "json"
	gzipCompression = "compress/gzip"
)

type (
	WrapFn func(*Blob) error

	UnwrapResult struct {
		EncodingFormat *string
		CompressionFormat *string
	}
)

// JsonEncoded returns a WrapFn which when invoked will indicate the blob was json encoded
func JsonEncoded() WrapFn {
	return func(b *Blob) error {
		wrappers := common.StringPtr(b.Tags[wrappersTag])
		if exists(wrappers, encodingKey) {
			return errors.New("encoding already specified")
		}
		push(wrappers, encodingKey, jsonEncoding)
		b.Tags[wrappersTag] = *wrappers
		return nil
	}
}

func Wrap(blob *Blob, functions ...WrapFn) error {
	for _, f := range functions {
		if err := f(blob); err != nil {
			return err
		}
	}
	return nil
}

func Unwrap(blob *Blob) (*UnwrapResult, error) {
	wrappers, ok := blob.Tags[wrappersTag]
	if !ok {
		return &UnwrapResult{}, nil
	}

	stack := common.StringPtr(wrappers)
	result := &UnwrapResult{}
	for len(*stack) != 0 {
		k, v, err := pop(stack)
		if err != nil {
			return nil, err
		}
		switch k {
		case encodingKey:
			result.EncodingFormat = common.StringPtr(v)
		case compressionKey:
			dBody, err := decompress(v)
			if err != nil {
				return nil, err
			}
			blob.Body = dBody
			result.CompressionFormat = common.StringPtr(v)
		default:
			return nil, errors.New("cannot unwrap encountered unknown key")
		}
	}
	delete(blob.Tags, wrappersTag)
	return result, nil
}

func decompress(compression string) ([]byte, error) {

}

const (
	mapSeparator = ","
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
