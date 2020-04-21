package filestore

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/service/config"

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

func (s *ClientSuite) TestNewFilestoreClient_InvalidConfig() {
	_, err := NewFilestoreClient(nil)
	s.Error(err)
	_, err = NewFilestoreClient(&config.FileBlobstore{})
	s.Error(err)
}

func (s *ClientSuite) TestNewFilestoreClient_DirectoryAlreadyExists() {
	name := s.createTempDir("TestNewFilestoreClient_DirectoryAlreadyExists")
	defer os.RemoveAll(name)
	c, err := NewFilestoreClient(&config.FileBlobstore{OutputDirectory: name})
	s.NoError(err)
	s.Equal(name, c.(*client).outputDirectory)
}

func (s *ClientSuite) TestNewFilestoreClient_DirectoryNotAlreadyExists() {
	name := s.createTempDir("TestNewFilestoreClient_DirectoryNotAlreadyExists")
	os.RemoveAll(name)
	exists, err := common.DirectoryExists(name)
	s.NoError(err)
	s.False(exists)
	c, err := NewFilestoreClient(&config.FileBlobstore{OutputDirectory: name})
	s.NoError(err)
	s.Equal(name, c.(*client).outputDirectory)
	exists, err = common.DirectoryExists(name)
	s.NoError(err)
	s.True(exists)
	os.RemoveAll(name)
}

func (s *ClientSuite) TestCrudOperations() {
	name := s.createTempDir("TestCrudOperations")
	defer os.RemoveAll(name)
	c, err := NewFilestoreClient(&config.FileBlobstore{OutputDirectory: name})
	s.NoError(err)

	// put three blobs in blobstore
	key1 := uuid.New()
	key2 := uuid.New()
	key3 := uuid.New()
	blob1 := blobstore.Blob{
		Tags: nil,
		Body: []byte{1, 2, 3},
	}
	blob2 := blobstore.Blob{
		Tags: map[string]string{"key1": "value1"},
		Body: nil,
	}
	blob3 := blobstore.Blob{
		Tags: map[string]string{"key1": "value1", "key2": "value2"},
		Body: []byte{1, 2, 3, 4, 5},
	}
	_, err = c.Put(nil, &blobstore.PutRequest{
		Key: key1,
		Blob: blob1,
	})
	s.NoError(err)
	_, err = c.Put(nil, &blobstore.PutRequest{
		Key: key2,
		Blob: blob2,
	})
	s.NoError(err)
	_, err = c.Put(nil, &blobstore.PutRequest{
		Key: key3,
		Blob: blob3,
	})
	s.NoError(err)

	// get the blobs back
	get1, err := c.Get(nil, &blobstore.GetRequest{Key: key1})
	s.NoError(err)
	s.Nil(get1.Blob.Tags)
	s.Equal([]byte{1, 2, 3}, get1.Blob.Body)
	get2, err := c.Get(nil, &blobstore.GetRequest{Key: key2})
	s.NoError(err)
	s.Equal(map[string]string{"key1": "value1"}, get2.Blob.Tags)
	s.Nil(get2.Blob.Body)
	get3, err := c.Get(nil, &blobstore.GetRequest{Key: key3})
	s.NoError(err)
	s.Equal(map[string]string{"key1": "value1", "key2": "value2"}, get3.Blob.Tags)
	s.Equal([]byte{1, 2, 3, 4, 5}, get3.Blob.Body)

	// confirm all the blobs exist
	exists1, err := c.Exists(nil, &blobstore.ExistsRequest{Key: key1})
	s.NoError(err)
	s.True(exists1.Exists)
	exists2, err := c.Exists(nil, &blobstore.ExistsRequest{Key: key2})
	s.NoError(err)
	s.True(exists2.Exists)
	exists3, err := c.Exists(nil, &blobstore.ExistsRequest{Key: key3})
	s.NoError(err)
	s.True(exists3.Exists)

	// delete a blob and confirm no longer can get and that no longer exists
	_, err = c.Delete(nil, &blobstore.DeleteRequest{Key: key1})
	s.NoError(err)
	exists1, err = c.Exists(nil, &blobstore.ExistsRequest{Key: key1})
	s.NoError(err)
	s.False(exists1.Exists)
	get1, err = c.Get(nil, &blobstore.GetRequest{Key: key1})
	s.Error(err)
	s.Nil(get1)
}

func (s *ClientSuite) createTempDir(prefix string) string {
	name, err := ioutil.TempDir("", prefix)
	s.NoError(err)
	return name
}