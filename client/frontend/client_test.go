// Copyright (c) 2021 Uber Technologies, Inc.
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

package frontend

import (
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxTimeouts(t *testing.T) {
	// Since MaxTimeouts list every method by string
	// Here we ensure that they reflect all actual methods.
	// We test both ways: the map must contain all client methods and do not contain any other keys.
	c := grpcClient{}
	client := reflect.TypeOf(c)
	for i := 0; i < client.NumMethod(); i++ {
		require.Contains(t, MaxTimeouts, client.Method(i).Name)
	}
	for method := range MaxTimeouts {
		_, exists := client.MethodByName(method)
		require.True(t, exists, "MaxTimeouts contains unknown method %s", method)
	}
}
