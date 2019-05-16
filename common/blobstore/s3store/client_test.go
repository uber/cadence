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

package s3store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/blobstore/blob"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"testing"
)

const (
	bucketInvalidMetadata      = "invalid-metadata"
	bucketNoMetadata           = "bucket-no-metadata"
	defaultBucketName          = "default-bucket-name"
	defaultBucketOwner         = "default-bucket-owner"
	defaultBucketRetentionDays = 10
	customBucketNamePrefix     = "default-bucket-name-custom"
	customBucketOwner          = "custom-bucket-owner"
	customBucketRetentionDays  = 100
	numberOfCustomBuckets      = 5
)

type ClientSuite struct {
	*require.Assertions
	suite.Suite
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *ClientSuite) TestUploadDownload_Success() {

	client := s.constructClient()
	b := blob.NewBlob([]byte("body version 1"), map[string]string{})
	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	s.NoError(client.Upload(context.Background(), defaultBucketName, key, b))
	downloadBlob, err := client.Download(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.NotNil(downloadBlob)
	s.assertBlobEquals(map[string]string{}, "body version 1", downloadBlob)

	b = blob.NewBlob([]byte("body version 2"), map[string]string{"key": "value"})
	s.NoError(client.Upload(context.Background(), defaultBucketName, key, b))
	downloadBlob, err = client.Download(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.NotNil(downloadBlob)
	s.assertBlobEquals(map[string]string{"key": "value"}, "body version 2", downloadBlob)
}

func (s *ClientSuite) TestUpload_Fail_BucketNotExists() {
	client := s.constructClient()

	b := blob.NewBlob([]byte("blob body"), map[string]string{"tagKey": "tagValue"})
	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	s.Equal(blobstore.ErrBucketNotExists, client.Upload(context.Background(), "bucket-not-exists", key, b))
}

/*
Not sure how an error would fail
func (s *ClientSuite) TestUpload_Fail_ErrorOnWrite() {

	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	//	s.NoError(mkdirAll(path.Join(dir, defaultBucketName, key.String(), "foo")))
	client := s.constructClient()

	b := blob.NewBlob([]byte("blob body"), map[string]string{"tagKey": "tagValue"})
	s.Equal(ErrWriteFile, client.Upload(context.Background(), defaultBucketName, key, b))
}

*/

func (s *ClientSuite) TestDownload_Fail_BucketNotExists() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blobname.ext")
	s.NoError(err)
	b, err := client.Download(context.Background(), "bucket-not-exists", key)
	s.Equal(blobstore.ErrBucketNotExists, err)
	s.Nil(b)
}

func (s *ClientSuite) TestDownload_Fail_BlobNotExists() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blobname.ext")
	s.NoError(err)
	b, err := client.Download(context.Background(), defaultBucketName, key)
	s.Equal(blobstore.ErrBlobNotExists, err)
	s.Nil(b)
}

/*
There is no invalid blob format
func (s *ClientSuite) TestDownload_Fail_BlobFormatInvalid() {

	client := s.constructClient()
	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)

	//	s.NoError(writeFile(filepath.Join(dir, defaultBucketName, key.String()), []byte("invalid")))

	b, err := client.Download(context.Background(), defaultBucketName, key)
	s.Equal(blobstore.ErrBlobDeserialization, err)
	s.Nil(b)
}
*/

func (s *ClientSuite) TestUploadDownload_Success_CustomBucket() {

	client := s.constructClient()
	b := blob.NewBlob([]byte("blob body"), map[string]string{})
	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	customBucketName := fmt.Sprintf("%v-%v", customBucketNamePrefix, 3)
	s.NoError(client.Upload(context.Background(), customBucketName, key, b))
	downloadBlob, err := client.Download(context.Background(), customBucketName, key)
	s.NoError(err)
	s.NotNil(downloadBlob)
	s.assertBlobEquals(map[string]string{}, "blob body", downloadBlob)
}

func (s *ClientSuite) TestExists_Fail_BucketNotExists() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	exists, err := client.Exists(context.Background(), "bucket-not-exists", key)
	s.Equal(blobstore.ErrBucketNotExists, err)
	s.False(exists)
}

func (s *ClientSuite) TestExists_Success() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	exists, err := client.Exists(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.False(exists)

	b := blob.NewBlob([]byte("body"), map[string]string{})
	s.NoError(client.Upload(context.Background(), defaultBucketName, key, b))
	exists, err = client.Exists(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.True(exists)
}

func (s *ClientSuite) TestDelete_Fail_BucketNotExists() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	deleted, err := client.Delete(context.Background(), "bucket-not-exists", key)
	s.Equal(blobstore.ErrBucketNotExists, err)
	s.False(deleted)
}

func (s *ClientSuite) TestDelete_Success() {
	client := s.constructClient()

	key, err := blob.NewKeyFromString("blob.blob")
	s.NoError(err)
	deleted, err := client.Delete(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.False(deleted)

	b := blob.NewBlob([]byte("body"), map[string]string{})
	s.NoError(client.Upload(context.Background(), defaultBucketName, key, b))
	deleted, err = client.Delete(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.True(deleted)
	exists, err := client.Exists(context.Background(), defaultBucketName, key)
	s.NoError(err)
	s.False(exists)
}

func (s *ClientSuite) TestListByPrefix_Fail_BucketNotExists() {
	client := s.constructClient()

	files, err := client.ListByPrefix(context.Background(), "bucket-not-exists", "foo")
	s.Equal(blobstore.ErrBucketNotExists, err)
	s.Nil(files)
}

func (s *ClientSuite) TestListByPrefix_Success() {
	client := s.constructClient()

	allFiles := []string{"matching_1.ext", "matching_2.ext", "not_matching_1.ext", "not_matching_2.ext", "matching_3.ext"}
	for _, f := range allFiles {
		key, err := blob.NewKeyFromString(f)
		s.NoError(err)
		b := blob.NewBlob([]byte("blob body"), map[string]string{})
		s.NoError(client.Upload(context.Background(), defaultBucketName, key, b))
	}
	matchingKeys, err := client.ListByPrefix(context.Background(), defaultBucketName, "matching")
	s.NoError(err)
	var matchingFilenames []string
	for _, k := range matchingKeys {
		matchingFilenames = append(matchingFilenames, k.String())
	}
	s.Equal([]string{"matching_1.ext", "matching_2.ext", "matching_3.ext"}, matchingFilenames)
}

func (s *ClientSuite) TestBucketMetadata_Fail_BucketNotExists() {
	client := s.constructClient()

	metadata, err := client.BucketMetadata(context.Background(), "bucket-not-exists")
	s.Equal(blobstore.ErrBucketNotExists, err)
	s.Nil(metadata)
}

/*

Note: bucket metadata always exists as long as the bucket exists and is in the correct format
func (s *ClientSuite) TestBucketMetadata_Fail_FileNotExistsError() {
	client := s.constructClient()
	//	s.NoError(os.Remove(bucketItemPath(dir, defaultBucketName, metadataFilename)))

	metadata, err := client.BucketMetadata(context.Background(), bucketNoMetadata)
	s.Equal(ErrReadFile, err)
	s.Nil(metadata)
}
func (s *ClientSuite) TestBucketMetadata_Fail_InvalidFileFormat() {

	client := s.constructClient()

	metadata, err := client.BucketMetadata(context.Background(), bucketInvalidMetadata)
	s.Equal(ErrBucketConfigDeserialization, err)
	s.Nil(metadata)
}
*/

func (s *ClientSuite) TestBucketMetadata_Success() {
	client := s.constructClient()

	metadata, err := client.BucketMetadata(context.Background(), defaultBucketName)
	s.NoError(err)
	s.NotNil(metadata)
	s.Equal(defaultBucketRetentionDays, metadata.RetentionDays)
	s.Equal(defaultBucketOwner, metadata.Owner)
}

func (s *ClientSuite) TestBucketExists() {
	client := s.constructClient()

	exists, err := client.BucketExists(context.Background(), "bucket-not-exists")
	s.NoError(err)
	s.False(exists)

	exists, err = client.BucketExists(context.Background(), defaultBucketName)
	s.NoError(err)
	s.True(exists)
}

type mockS3Client struct {
	s3iface.S3API
	fs   map[string]string
	tags map[string]string
}

func (m *mockS3Client) HeadBucketWithContext(_ aws.Context, input *s3.HeadBucketInput, _ ...request.Option) (*s3.HeadBucketOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}

	return &s3.HeadBucketOutput{}, nil

}

func (m *mockS3Client) DeleteObjectWithContext(_ aws.Context, input *s3.DeleteObjectInput, _ ...request.Option) (*s3.DeleteObjectOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}

	c := m.fs[*input.Bucket+*input.Key]
	if c == "" {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "", nil)

	}
	delete(m.fs, *input.Bucket+*input.Key)
	return &s3.DeleteObjectOutput{}, nil
}
func (m *mockS3Client) ListObjectsV2WithContext(_ aws.Context, input *s3.ListObjectsV2Input, _ ...request.Option) (*s3.ListObjectsV2Output, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}

	objects := make([]*s3.Object, 0)

	for k := range m.fs {
		if strings.HasPrefix(k, *input.Bucket+*input.Prefix) {

			str := strings.Replace(k, *input.Bucket, "", 1)
			object := s3.Object{
				Key: &str,
			}
			objects = append(objects, &object)
		}
	}
	sort.SliceStable(objects, func(i, j int) bool {
		return *objects[i].Key < *objects[j].Key

	})
	return &s3.ListObjectsV2Output{
		Contents: objects,
	}, nil
}

func (m *mockS3Client) GetObjectTaggingWithContext(_ aws.Context, input *s3.GetObjectTaggingInput, _ ...request.Option) (*s3.GetObjectTaggingOutput, error) {
	t := m.tags[*input.Bucket+*input.Key]

	ql, _ := url.ParseQuery(t)

	tags := make([]*s3.Tag, 0, len(ql))
	for k, v := range ql {

		tag := s3.Tag{
			Key:   &k,
			Value: &v[0],
		}
		tags = append(tags, &tag)

	}
	return &s3.GetObjectTaggingOutput{
		TagSet: tags,
	}, nil
}

func (m *mockS3Client) GetBucketAclWithContext(_ aws.Context, input *s3.GetBucketAclInput, _ ...request.Option) (*s3.GetBucketAclOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}
	return &s3.GetBucketAclOutput{
		Owner: &s3.Owner{
			DisplayName: aws.String(defaultBucketOwner),
		},
	}, nil
}

func (m *mockS3Client) GetBucketLifecycleWithContext(_ aws.Context, input *s3.GetBucketLifecycleInput, _ ...request.Option) (*s3.GetBucketLifecycleOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}

	days := int64(defaultBucketRetentionDays)
	rules := make([]*s3.Rule, 0, 1)
	rules = append(rules, &s3.Rule{
		Expiration: &s3.LifecycleExpiration{
			Days: &days,
		},
	})

	return &s3.GetBucketLifecycleOutput{
		Rules: rules,
	}, nil
}
func (m *mockS3Client) PutObjectWithContext(_ aws.Context, input *s3.PutObjectInput, _ ...request.Option) (*s3.PutObjectOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(input.Body)

	m.fs[*input.Bucket+*input.Key] = buf.String()
	if input.Tagging != nil {
		m.tags[*input.Bucket+*input.Key] = *input.Tagging
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) HeadObjectWithContext(_ aws.Context, input *s3.HeadObjectInput, _ ...request.Option) (*s3.HeadObjectOutput, error) {

	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}

	c := m.fs[*input.Bucket+*input.Key]
	if c == "" {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "", nil)

	}
	return &s3.HeadObjectOutput{}, nil
}

func (m *mockS3Client) GetObjectWithContext(_ aws.Context, input *s3.GetObjectInput, _ ...request.Option) (*s3.GetObjectOutput, error) {

	if *input.Bucket == *aws.String(bucketInvalidMetadata) {
		return &s3.GetObjectOutput{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(
				"dummy-content"))),
		}, nil
	}
	if *input.Bucket == *aws.String(bucketNoMetadata) {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "", nil)

	}
	if !strings.HasPrefix(*input.Bucket, defaultBucketName) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "", nil)
	}
	c := m.fs[*input.Bucket+*input.Key]
	if c == "" {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "", nil)

	}
	return &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader([]byte(c))),
	}, nil

}

func NewMockClient() (blobstore.Client, error) {

	return &client{
		s3cli: &mockS3Client{
			fs:   make(map[string]string),
			tags: make(map[string]string),
		},
	}, nil
}
func (s *ClientSuite) constructClient() blobstore.Client {
	client, err := NewMockClient()
	s.NoError(err)
	s.NotNil(client)
	return client
}

func (s *ClientSuite) assertBlobEquals(expectedTags map[string]string, expectedBody string, actual *blob.Blob) {
	s.Equal(expectedTags, actual.Tags)
	s.Equal(expectedBody, string(actual.Body))
}
