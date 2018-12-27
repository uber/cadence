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

func (s *UtilSuite) TestMkdirAll() {
	dir, err := ioutil.TempDir("", "test.mkdir.all")
	s.NoError(err)
	defer os.RemoveAll(dir)

	s.assertDirectoryExists(dir)
	s.NoError(mkdirAll(dir))
	s.assertDirectoryExists(dir)

	subdirPath := filepath.Join(dir, "subdir_1", "subdir_2", "subdir_3")
	s.assertDirectoryNotExists(subdirPath)
	s.NoError(mkdirAll(subdirPath))
	s.assertDirectoryExists(subdirPath)
	s.assertCorrectFileMode(subdirPath)

	filename := "test-file-name"
	s.createFile(dir, filename)
	fpath := filepath.Join(dir, filename)
	s.Error(mkdirAll(fpath))
}

func (s *UtilSuite) TestWriteFile() {
	dir, err := ioutil.TempDir("", "test.write.file")
	s.NoError(err)
	defer os.RemoveAll(dir)

	s.assertDirectoryExists(dir)

	filename := "test-file-name"
	fpath := filepath.Join(dir, filename)
	s.NoError(writeFile(fpath, []byte("file body 1")))
	s.assertFileExists(fpath)
	s.assertCorrectFileMode(fpath)

	s.NoError(writeFile(fpath, []byte("file body 2")))
	s.assertFileExists(fpath)
	s.assertCorrectFileMode(fpath)

	s.Error(writeFile(dir, []byte("")))
	s.assertFileExists(fpath)
}

func (s *UtilSuite) TestReadFile() {
	dir, err := ioutil.TempDir("", "test.read.file")
	s.NoError(err)
	defer os.RemoveAll(dir)

	s.assertDirectoryExists(dir)

	filename := "test-file-name"
	fpath := filepath.Join(dir, filename)
	data, err := readFile(fpath)
	s.Error(err)
	s.Empty(data)

	writeFile(fpath, []byte("file contents"))
	data, err = readFile(fpath)
	s.NoError(err)
	s.Equal("file contents", string(data))
}

func (s *UtilSuite) createFile(dir string, filename string) {
	err := ioutil.WriteFile(filepath.Join(dir, filename), []byte("file contents"), mode)
	s.Nil(err)
}

func (s *UtilSuite) assertFileExists(filepath string) {
	exists, err := fileExists(filepath)
	s.NoError(err)
	s.True(exists)
}

func (s *UtilSuite) assertDirectoryExists(path string) {
	exists, err := directoryExists(path)
	s.NoError(err)
	s.True(exists)
}

func (s *UtilSuite) assertDirectoryNotExists(path string) {
	exists, err := directoryExists(path)
	s.NoError(err)
	s.False(exists)
}

func (s *UtilSuite) assertCorrectFileMode(path string) {
	info, err := os.Stat(path)
	s.NoError(err)
	mode := mode
	if info.IsDir() {
		mode = mode | os.ModeDir
	}
	s.Equal(mode, info.Mode())
}
