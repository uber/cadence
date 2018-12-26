package filestore

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type UtilSuite struct {
	*require.Assertions
	suite.Suite
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *UtilSuite) TestFileExists() {
	dir, err := ioutil.TempDir("", "test.file.exists")
	s.NoError(err)
	defer os.RemoveAll(dir)

	exists, err := fileExists(dir)
	s.Error(err)
	s.False(exists)

	filename := "test-file-name"
	exists, err = fileExists(filepath.Join(dir, filename))
	s.NoError(err)
	s.False(exists)

	s.createFile(dir, filename)
	exists, err = fileExists(filepath.Join(dir, filename))
	s.NoError(err)
	s.True(exists)
}

func (s *UtilSuite) TestDirectoryExists() {
	dir, err := ioutil.TempDir("", "test.directory.exists")
	s.NoError(err)
	defer os.RemoveAll(dir)

	exists, err := directoryExists(dir)
	s.NoError(err)
	s.True(exists)

	subdir := "subdir"
	exists, err = directoryExists(filepath.Join(dir, subdir))
	s.NoError(err)
	s.False(exists)

	filename := "test-file-name"
	s.createFile(dir, filename)
	fpath := filepath.Join(dir, filename)

	exists, err = directoryExists(fpath)
	s.Error(err)
	s.False(exists)
}

func (s *UtilSuite) TestCreateDirIfNotExists() {
	dir, err := ioutil.TempDir("", "test.create.dir.if.not.exists")
	s.NoError(err)
	defer os.RemoveAll(dir)

	exists, err := directoryExists(dir)
	s.NoError(err)
	s.True(exists)
	s.NoError(createDirIfNotExists(dir))
	exists, err = directoryExists(dir)
	s.NoError(err)
	s.True(exists)

	subdirPath := filepath.Join(dir, "subdir")
	exists, err = directoryExists(subdirPath)
	s.NoError(err)
	s.False(exists)
	s.NoError(createDirIfNotExists(subdirPath))
	exists, err = directoryExists(subdirPath)
	s.NoError(err)
	s.True(exists)
	s.checkModesMatch()
}

func (s *UtilSuite) TestWriteFile() {
	dir, err := ioutil.TempDir("", "test.write.file")
	s.NoError(err)
	defer os.RemoveAll(dir)

	s.Error(writeFile(dir, []byte("")))
	exists, err := directoryExists(dir)
	s.NoError(err)
	s.True(exists)

	filename := "test-file-name"
	fpath := filepath.Join(dir, filename)
	data := []byte("file contents 1")
	s.checkFileNotExists(fpath)
	s.NoError(writeFile(fpath, data))
	s.checkFileContentsMatch(fpath, "file contents 1")


	// file already exists
	// check file permissions
	// check that data actually gets updated
}


// TODO: clean up these methods the tests look weird right now



func (s *UtilSuite) createFile(dir string, filename string) {
	err := ioutil.WriteFile(filepath.Join(dir, filename), []byte("file contents"), mode)
	s.Nil(err)
}

func (s *UtilSuite) checkFileExists(filepath string) {
	exists, err := fileExists(filepath)
	s.NoError(err)
	s.True(exists)
}

func (s *UtilSuite) checkFileNotExists(filepath string) {
	exists, err := fileExists(filepath)
	s.NoError(err)
	s.False(exists)
}

func (s *UtilSuite) checkFileContentsMatch(filepath string, expected string) {
	s.checkFileExists(filepath)
	data, err := ioutil.ReadFile(filepath)
	s.NoError(err)
	s.Equal(expected, string(data))
}

func (s *UtilSuite) checkModesMatch(filepath string, mode os.FileMode) {
	s.checkFileExists(filepath)
	info, err := os.Stat(filepath)
	s.NoError(err)
	s.Equal(mode, info.Mode())
}