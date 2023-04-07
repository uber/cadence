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

package proto

import (
	"testing"

	// import two differently-generated versions of "the same" IDL spec.
	//
	// e.g. this is used to both build the admin client in go.uber.org/cadence,
	// and to contact the admin APIs with the main `cadence` CLI tool.
	//
	// it is quite likely that other packages test this implicitly, but this one exists to
	// serve as a place to document the problem.
	_ "github.com/uber/cadence-idl/go/proto/admin/v1" // will be moved to the client library
	_ "github.com/uber/cadence/.gen/proto/admin/v1"   // server-internal
)

func TestPackageRewritingAllowsBothIDLs(t *testing.T) {
	// this is a test to ensure we can import both server IDL and client IDL,
	// i.e. that our workaround for duplicate registered types worked.
	//
	// if this did not panic at `init` time, it worked.
	//
	// the main purpose of this is to ensure that a monorepo can import different versions of the "same" IDL
	// specs, without running into gogoproto's duplicate-global-registration panics.
	// this allows the server to create and implement new client APIs without also forcing the client library
	// to support the same changes at the same time (a problem in both this repo and any single-version monorepo).
	//
	// this should always be safe as long as on-the-wire backwards compatibility is maintained (it must be),
	// and as long as we do not use `google.protobuf.Any` blindly (client and server names of the same type
	// will be different, but binary-compatible).
	// Currently `google.protobuf.Any` is unused, and this will likely remain true.  A check is done inside
	// the makefile to ensure this remains true.
}
