// Copyright (c) 2017 Uber Technologies, Inc.
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

package auth

//This follows what is done for the tls support for kafka and cassandra with minor changes.

type (
	// SQLTLS describes TLS configuration (for MySQL)
	SQLTLS struct {
		// CaFile is the path name of the Certificate Authority (CA) certificate file.
		CaFile string `yaml:"caFile"`
		// CertFile is the path name of the certificate signed by the CA.
		CertFile string `yaml:"certFile"`
		// KeyFile is the path name of the key used to sign the certificate.
		KeyFile string `yaml:"keyFile"`
		// ServerName : Common Name
		ServerName string `yaml:"serverName"`
		// Enabled is used to check if TLS is enabled
		Enabled bool `yaml:"enabled"`
		//Config: customValue in DSN
		Config string `yaml:"config"`
	}
)
