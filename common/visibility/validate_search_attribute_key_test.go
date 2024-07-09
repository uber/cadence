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

package visibility

import (
	"testing"
)

func TestValidateSearchAttributeKey(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pure character",
			args: args{name: "CustomStringField"},
		},
		{
			name: "alphanumeric",
			args: args{name: "CustomStringField1"},
		},
		{
			name:    "start with number",
			args:    args{name: "1CustomStringField"},
			wantErr: true,
		},
		{
			name:    "contain special character",
			args:    args{name: "CustomStringField!"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateSearchAttributeKey(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchAttributeKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeSearchAttributeKey(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "pure character",
			args: args{name: "CustomStringField"},
			want: "CustomStringField",
		},
		{
			name: "alphanumeric",
			args: args{name: "CustomStringField1"},
			want: "CustomStringField1",
		},
		{
			name: "start with number",
			args: args{name: "1CustomStringField"},
			want: "CustomStringField",
		},
		{
			name: "contain special character",
			args: args{name: "CustomStringField!"},
			want: "CustomStringField",
		},
		{
			name: "contain special character in the middle",
			args: args{name: "CustomString-Field"},
			want: "CustomString_Field",
		},
		{
			name:    "all numbers",
			args:    args{name: "1234567890"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeSearchAttributeKey(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeSearchAttributeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeSearchAttributeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
