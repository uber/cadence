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

// Package shared holds some types that are used between multiple global-ratelimiter
// packages, and need to be broken out to prevent import cycles.
package shared

import (
	"fmt"
	"strings"
)

type (
	// KeyMapper is used to ensure that all keys that get communicated to
	// aggregators are uniquely identifiable, when multiple collections are used.
	//
	// A unique prefix / pattern per collection should be sufficient.
	KeyMapper interface {
		LocalToGlobal(key LocalKey) GlobalKey
		// GlobalToLocal does the reverse of LocalToGlobal.
		//
		// An error will be returned if the global key does not seem to have
		// come from this mapper.
		GlobalToLocal(key GlobalKey) (LocalKey, error)
	}
	LocalKey  string // LocalKey represents a "local" / unique to this collection ratelimit key
	GlobalKey string // GlobalKey represents a "global" / globally unique ratelimit key (i.e. namespaced by collection)

	simplekm struct {
		prefix string
	}
)

var _ KeyMapper = simplekm{}

func (s simplekm) LocalToGlobal(key LocalKey) GlobalKey {
	return GlobalKey(s.prefix + string(key))
}

func (s simplekm) GlobalToLocal(key GlobalKey) (LocalKey, error) {
	k := string(key)
	if !strings.HasPrefix(k, s.prefix) {
		return "", fmt.Errorf("missing prefix %q in global key: %q", s.prefix, key)
	}
	return LocalKey(strings.TrimPrefix(k, s.prefix)), nil
}

// PrefixKey builds a KeyMapper that simply adds (or removes) a fixed prefix to all keys.
// If the global key does not have this prefix, an error will be returned when mapping to local.
func PrefixKey(prefix string) KeyMapper {
	return simplekm{
		prefix: prefix,
	}
}
