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

package filestore

import (
	"errors"
	"io/ioutil"
	"os"
)

const mode = os.FileMode(0600)

var (
	errReadFileNotExists = errors.New("attempted to read a file which does not exist")
)

func fileExists(filepath string) (bool, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if info.IsDir() {
		return false, errors.New("specified directory not file")
	}
	return true, nil
}

func directoryExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, errors.New("specified file not directory")
	}
	return true, nil
}

func createDirIfNotExists(dir string) error {
	exists, err := directoryExists(dir)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := os.Mkdir(dir, mode); err != nil {
		return err
	}
	return nil
}

func writeFile(filepath string, data []byte) error {
	removeIfExists := func() error {
		info, err := os.Stat(filepath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return errors.New("attempted to delete directory")
		}
		return os.Remove(filepath)
	}

	if err := removeIfExists(); err != nil {
		return err
	}
	f, err := os.Create(filepath)
	defer f.Close()
	if err != nil {
		return err
	}
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	return nil
}

func readFile(filepath string) ([]byte, error) {
	exists, err := fileExists(filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errReadFileNotExists
	}
	return ioutil.ReadFile(filepath)
}
