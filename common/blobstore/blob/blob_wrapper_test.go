package blob

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common"
	"testing"
)

type BlobWrapperSuite struct {
	*require.Assertions
	suite.Suite
}

func TestBlobWrapperSuite(t *testing.T) {
	suite.Run(t, new(BlobWrapperSuite))
}

func (s *BlobWrapperSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *BlobWrapperSuite) TestJsonEncodedWrapFn() {
	testCases := []struct {
		inputTags   map[string]string
		inputBody   []byte
		expectError bool
		expectTags  map[string]string
	}{
		{
			inputTags: map[string]string{
				wrappersTag: "encoding:exists,",
			},
			inputBody:   []byte("test-body"),
			expectError: true,
		},
		{
			inputTags:   map[string]string{},
			inputBody:   []byte("test-body"),
			expectError: false,
			expectTags: map[string]string{
				wrappersTag: "encoding:json,",
			},
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "compression:exists,",
			},
			inputBody:   []byte("test-body"),
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "encoding:json,compression:exists,",
			},
		},
	}

	for _, tc := range testCases {
		wrapFn := JsonEncoded()
		blob := NewBlob(tc.inputBody, tc.inputTags)
		err := wrapFn(blob)
		if tc.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
			s.Equal(tc.expectTags, blob.Tags)
			s.Equal(tc.inputBody, blob.Body)
		}
	}
}

func (s *BlobWrapperSuite) TestGzipCompressedWrapFn() {
	testCases := []struct {
		inputTags   map[string]string
		inputBody   []byte
		expectError bool
		expectTags  map[string]string
	}{
		{
			inputTags: map[string]string{
				wrappersTag: "compression:exists,",
			},
			inputBody:   []byte("test-body"),
			expectError: true,
		},
		{
			inputTags:   map[string]string{},
			inputBody:   []byte("test-body"),
			expectError: false,
			expectTags: map[string]string{
				wrappersTag: "compression:compress/gzip,",
			},
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "encoding:exists,",
			},
			inputBody:   []byte("test-body"),
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "compression:compress/gzip,encoding:exists,",
			},
		},
	}

	for _, tc := range testCases {
		wrapFn := GzipCompressed()
		blob := NewBlob(tc.inputBody, tc.inputTags)
		err := wrapFn(blob)
		if tc.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
			s.Equal(tc.expectTags, blob.Tags)
			s.NotContains(tc.inputBody, blob.Body)
		}
	}
}

func (s *BlobWrapperSuite) TestWrap() {
	testCases := []struct {
		inputTags           map[string]string
		inputBody           []byte
		functions           []WrapFn
		expectError         bool
		expectTags          map[string]string
		expectDifferentBody bool
	}{
		{
			inputTags: map[string]string{
				wrappersTag: "compression:exists,",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{GzipCompressed()},
			expectError: true,
		},
		{
			inputTags: map[string]string{
				wrappersTag: "encoding:exists,",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{JsonEncoded()},
			expectError: true,
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "encoding:exists,",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{GzipCompressed()},
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "compression:compress/gzip,encoding:exists,",
			},
			expectDifferentBody: true,
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "compression:exists,",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{JsonEncoded()},
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "encoding:json,compression:exists,",
			},
			expectDifferentBody: false,
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{JsonEncoded(), GzipCompressed()},
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "compression:compress/gzip,encoding:json,",
			},
			expectDifferentBody: true,
		},
		{
			inputTags: map[string]string{
				"user_tag_key": "user_tag_value",
			},
			inputBody:   []byte("test-body"),
			functions:   []WrapFn{GzipCompressed(), JsonEncoded()},
			expectError: false,
			expectTags: map[string]string{
				"user_tag_key": "user_tag_value",
				wrappersTag:    "encoding:json,compression:compress/gzip,",
			},
			expectDifferentBody: true,
		},
	}

	for _, tc := range testCases {
		inputBlob := NewBlob(tc.inputBody, tc.inputTags)
		wrappedBlob, err := Wrap(inputBlob, tc.functions...)
		if tc.expectError {
			s.Error(err)
			s.Nil(wrappedBlob)
		} else {
			s.NoError(err)
			s.NotNil(wrappedBlob)
			s.False(wrappedBlob == inputBlob)
			s.False(&wrappedBlob.Tags == &inputBlob.Tags)
			s.False(&wrappedBlob.Body == &inputBlob.Body)
			s.Equal(tc.expectTags, wrappedBlob.Tags)
			if tc.expectDifferentBody {
				s.NotEqual(inputBlob.Body, wrappedBlob.Body)
			} else {
				s.Equal(inputBlob.Body, wrappedBlob.Body)
			}
		}
	}
}

func (s *BlobWrapperSuite) TestUnwrap() {

	// stack is malformed returns an error
	// wrapper contained unknown key returns error
	// encodingKey only given
	// compressionKey only given
	// both given in each order

	testCases := []struct {
		inputBlob            *Blob
		expectError          bool
		expectWrappingLayers *WrappingLayers
		expectBlob           *Blob
	}{
		{
			inputBlob:            nil,
			expectError:          false,
			expectWrappingLayers: &WrappingLayers{},
			expectBlob:           nil,
		},
		{
			inputBlob: s.wrappedBlob(
				map[string]string{"user_tag_key": "user_tag_value"},
				[]byte("test-body"),
			),
			expectError: false,
			expectWrappingLayers: &WrappingLayers{},
			expectBlob: s.wrappedBlob(
				map[string]string{"user_tag_key": "user_tag_value"},
				[]byte("test-body"),
			),
		},
		{
			inputBlob: s.wrappedBlob(
				map[string]string{
					"user_tag_key": "user_tag_value",
					wrappersTag: "",
				},
				[]byte("test-body"),
			),
			expectError: false,
			expectWrappingLayers: &WrappingLayers{},
			expectBlob: s.wrappedBlob(
				map[string]string{"user_tag_key": "user_tag_value"},
				[]byte("test-body"),
			),
		},
	}

	for _, tc := range testCases {
		unwrappedBlob, layers, err := Unwrap(tc.inputBlob)
		if tc.expectError {
			s.Error(err)
			s.Nil(unwrappedBlob)
			s.Nil(layers)
		} else {
			s.NoError(err)
			s.Equal(*tc.expectWrappingLayers, *layers)
			if tc.inputBlob == nil {
				s.Nil(unwrappedBlob)
				continue
			}
			s.False(unwrappedBlob == tc.inputBlob)
			s.False(&unwrappedBlob.Tags == &tc.inputBlob.Tags)
			s.False(&unwrappedBlob.Body == &tc.inputBlob.Body)
			s.Equal(*tc.expectBlob, *unwrappedBlob)
		}
	}
}

func (s *BlobWrapperSuite) wrappedBlob(tags map[string]string, body []byte, functions ...WrapFn) *Blob {
	blob := NewBlob(body, tags)
	wrappedBlob, err := Wrap(blob, functions...)
	s.NoError(err)
	s.NotNil(wrappedBlob)
	return wrappedBlob
}

func (s *BlobWrapperSuite) TestWrapperStack() {
	// pop of empty stack returns error
	stack := common.StringPtr("")
	k, v, err := pop(stack)
	s.Empty(k)
	s.Empty(v)
	s.Error(err)

	// pop of stack without map separator returns error
	stack = common.StringPtr("malformed:stack")
	k, v, err = pop(stack)
	s.Empty(k)
	s.Empty(v)
	s.Error(err)

	// pop of stack without pair separator returns error
	stack = common.StringPtr("malformed_stack,key:value")
	k, v, err = pop(stack)
	s.Empty(k)
	s.Empty(v)
	s.Error(err)

	// pop of stack with too many pair separators returns error
	stack = common.StringPtr("mal:formed:stack,")
	k, v, err = pop(stack)
	s.Empty(k)
	s.Empty(v)
	s.Error(err)

	// exists of empty stack return false
	stack = common.StringPtr("")
	s.False(exists(stack, "not-exists"))

	// push single item and pop
	stack = common.StringPtr("")
	push(stack, "key", "value")
	s.Equal("key:value,", *stack)
	s.True(exists(stack, "key"))
	s.False(exists(stack, "not-exists"))
	k, v, err = pop(stack)
	s.Equal("key", k)
	s.Equal("value", v)
	s.NoError(err)
	s.Empty(*stack)

	// push two items and pop
	stack = common.StringPtr("")
	push(stack, "k1", "v1")
	push(stack, "k2", "v2")
	s.Equal("k2:v2,k1:v1,", *stack)
	k, v, err = pop(stack)
	s.Equal("k2", k)
	s.Equal("v2", v)
	s.NoError(err)
	s.Equal("k1:v1,", *stack)
}
