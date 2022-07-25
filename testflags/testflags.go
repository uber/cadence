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

package testflags

import (
	"flag"
	"testing"
)

var TestFlags struct {
	mysql     bool
	postgres  bool
	mongodb   bool
	cassandra bool
}

func init() {
	flag.BoolVar(&TestFlags.mysql, "mysql", false, "MySQL external dependency")
	flag.BoolVar(&TestFlags.postgres, "postgres", false, "PostGreSQL external dependency")
	flag.BoolVar(&TestFlags.mongodb, "mongodb", false, "MongoDB external dependency")
	flag.BoolVar(&TestFlags.cassandra, "cassandra", false, "Cassandra external dependency")
}

// TODO: Better instructions:
// 1) How to start external dependencies - maybe link to docs so we can change it easily?
// 2) How to run the tests
func RequireMySQL(t *testing.T) {
	if !TestFlags.mysql {
		t.Skip("Skipping test that requires 'mysql' to run - start XXX and append '-mysql' to run this test.")
	}
}

func RequirePostgres(t *testing.T) {
	if !TestFlags.postgres {
		t.Skip("Skipping test that requires 'postgres' to run - start XXX and append '-postgres' to run this test.")
	}
}

func RequireMongoDB(t *testing.T) {
	if !TestFlags.mongodb {
		t.Skip("Skipping test that requires 'mongodb' to run - start XXX and append '-mongodb' to run this test.")
	}
}

func RequireCassandra(t *testing.T) {
	if !TestFlags.cassandra {
		t.Skip("Skipping test that requires 'cassandra' to run - start XXX and append '-cassandra' to run this test.")
	}
}
