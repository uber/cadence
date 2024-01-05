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

package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
)

func Test_validatePermission(t *testing.T) {

	readRequestAttr := &Attributes{Permission: PermissionRead}
	writeRequestAttr := &Attributes{Permission: PermissionWrite}

	readWriteDomainData := domainData{
		common.DomainDataKeyForReadGroups:  "read1",
		common.DomainDataKeyForWriteGroups: "write1",
	}

	readDomainData := domainData{
		common.DomainDataKeyForReadGroups: "read1",
	}

	emptyDomainData := domainData{}

	tests := []struct {
		name       string
		claims     *JWTClaims
		attributes *Attributes
		data       map[string]string
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "no args should always fail",
			claims:     &JWTClaims{},
			attributes: &Attributes{},
			data:       emptyDomainData,
			wantErr:    assert.Error,
		},
		{
			name:       "Empty claims will be denied even when domain data has no groups",
			claims:     &JWTClaims{},
			attributes: writeRequestAttr,
			data:       emptyDomainData,
			wantErr:    assert.Error,
		},
		{
			name:       "Empty claims will be denied when domain data has at least one group",
			claims:     &JWTClaims{},
			attributes: writeRequestAttr,
			data:       readDomainData,
			wantErr:    assert.Error,
		},
		{
			name:       "Read-only groups should not get access to write groups",
			claims:     &JWTClaims{Groups: "read1"},
			attributes: writeRequestAttr,
			data:       readWriteDomainData,
			wantErr:    assert.Error,
		},
		{
			name:       "Write-only groups should get access to read groups",
			claims:     &JWTClaims{Groups: "write1"},
			attributes: readRequestAttr,
			data:       readWriteDomainData,
			wantErr:    assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, validatePermission(tt.claims, tt.attributes, tt.data))
		})
	}
}
