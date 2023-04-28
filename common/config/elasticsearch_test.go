// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSSigning_ValidateEmpty(t *testing.T) {

	tests := []struct {
		msg    string
		config AWSSigning
		err    bool
	}{
		{
			msg: "Empty config should error",
			config: AWSSigning{
				StaticCredential:      nil,
				EnvironmentCredential: nil,
			},
			err: true,
		},
		{
			msg: "error when both config sections are provided",
			config: AWSSigning{
				Enable:                false,
				StaticCredential:      &AWSStaticCredential{},
				EnvironmentCredential: &AWSEnvironmentCredential{},
			},
			err: true,
		},
		{
			msg: "StaticCredential must have region set",
			config: AWSSigning{
				Enable:                false,
				StaticCredential:      &AWSStaticCredential{},
				EnvironmentCredential: nil,
			},
			err: true,
		},
		{
			msg: "EnvironmentCredential must have region set",
			config: AWSSigning{
				Enable:                false,
				StaticCredential:      nil,
				EnvironmentCredential: &AWSEnvironmentCredential{},
			},
			err: true,
		},
		{
			msg: "Valid StaticCredential config should have no error ",
			config: AWSSigning{
				Enable:                false,
				StaticCredential:      &AWSStaticCredential{Region: "region1"},
				EnvironmentCredential: nil,
			},
			err: false,
		},
		{
			msg: "Valid EnvironmentCredential config should have no error",
			config: AWSSigning{
				Enable:                false,
				StaticCredential:      nil,
				EnvironmentCredential: &AWSEnvironmentCredential{Region: "region1"},
			},
			err: false,
		},
	}

	for _, tc := range tests {
		if tc.err {
			assert.Error(t, tc.config.Validate(), tc.msg)
		} else {
			assert.NoError(t, tc.config.Validate(), tc.msg)
		}
	}

}
