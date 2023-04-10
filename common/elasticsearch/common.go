// Copyright (c) 2020 Uber Technologies, Inc.
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

package elasticsearch

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	esaws "github.com/olivere/elastic/aws/v4"

	"github.com/uber/cadence/common/config"
)

const (
	// TODO https://github.com/uber/cadence/issues/3686
	oneMicroSecondInNano = int64(time.Microsecond / time.Nanosecond)

	esDocIDDelimiter = "~"
	esDocType        = "_doc"
	esDocIDSizeLimit = 512
)

// Build Http Client with TLS
func buildTLSHTTPClient(config config.TLS) (*http.Client, error) {
	tlsConfig, err := config.ToTLSConfig()
	if err != nil {
		return nil, err
	}

	// Setup HTTPS client
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	tlsClient := &http.Client{Transport: transport}

	return tlsClient, nil
}

func buildAWSSigningClient(awsconfig config.AWSSigning) (*http.Client, error) {
	if err := config.CheckAWSSigningConfig(awsconfig); err != nil {
		return nil, err
	}

	if awsconfig.EnvironmentCredential != nil {
		return signingClientFromEnv(*awsconfig.EnvironmentCredential)
	}

	return signingClientFromStatic(*awsconfig.StaticCredential)
}

func GetESDocIDSizeLimit() int {
	return esDocIDSizeLimit
}

func GetESDocType() string {
	return esDocType
}

func GetESDocDelimiter() string {
	return esDocIDDelimiter
}

func GenerateDocID(wid, rid string) string {
	return wid + esDocIDDelimiter + rid
}

// refer to https://github.com/olivere/elastic/blob/release-branch.v7/recipes/aws-connect-v4/main.go
func signingClientFromStatic(credentialConfig config.AWSStaticCredential) (*http.Client, error) {
	awsCredentials := credentials.NewStaticCredentials(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
		credentialConfig.SessionToken,
	)
	return esaws.NewV4SigningClient(awsCredentials, credentialConfig.Region), nil
}

func signingClientFromEnv(credentialConfig config.AWSEnvironmentCredential) (*http.Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(credentialConfig.Region)},
	)
	if err != nil {
		return nil, err
	}
	return esaws.NewV4SigningClient(sess.Config.Credentials, credentialConfig.Region), nil
}
