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

	// pop of stack pair without pair separator returns error
	stack = common.StringPtr("malformed_stack,key:value")
	k, v, err = pop(stack)
	s.Empty(k)
	s.Empty(v)
	s.Error(err)

	// pop of stack pair with too many pair separators returns error
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