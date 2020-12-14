// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package types

// HealthStatus is an internal type (TBD...)
type HealthStatus struct {
	Ok  bool    `json:"ok,required"`
	Msg *string `json:"msg,omitempty"`
}

// GetOk is an internal getter (TBD...)
func (v *HealthStatus) GetOk() (o bool) {
	if v != nil {
		return v.Ok
	}
	return
}

// GetMsg is an internal getter (TBD...)
func (v *HealthStatus) GetMsg() (o string) {
	if v != nil && v.Msg != nil {
		return *v.Msg
	}
	return
}

// MetaHealthArgs is an internal type (TBD...)
type MetaHealthArgs struct {
}

// MetaHealthResult is an internal type (TBD...)
type MetaHealthResult struct {
	Success *HealthStatus `json:"success,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MetaHealthResult) GetSuccess() (o *HealthStatus) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}
