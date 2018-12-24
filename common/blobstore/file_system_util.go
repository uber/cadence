package blobstore

import (
	"errors"
	"io/ioutil"
	"os"
)

const fileMode = 0777

func fileExists(filepath string) (bool, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func directoryExists(dirpath string) (bool, error) {
	info, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func ensureDirectoryExists(dirpath string) error {
	exists, err := directoryExists(dirpath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := os.Mkdir(dirpath, fileMode); err != nil {
		return err
	}
	return nil
}

func writeFile(data []byte, filepath string) error {
	ensureNotExists := func() error {
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

	if err := ensureNotExists(); err != nil {
		return err
	}
	f, err := os.Create(filepath)
	defer f.Close()
	if err != nil {
		return err
	}
	if err = f.Chmod(fileMode); err != nil {
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
		return nil, errors.New("attempted to read file which does not exist")
	}
	return ioutil.ReadFile(filepath)
}