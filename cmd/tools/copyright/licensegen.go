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

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type (
	// task that adds license header to source
	// files, if they don't already exist
	addLicenseHeaderTask struct {
		license string  // license header string to add
		config  *config // root directory of the project source
	}

	// command line config params
	config struct {
		rootDir            string
		verifyOnly         bool
		temporalAddMode    bool
		temporalModifyMode bool
		filePaths          string
	}
)

// licenseFileName is the name of the license file
const licenseFileName = "LICENSE"

// unique prefix that identifies a license header
const licenseHeaderPrefixOld = "Copyright (c)"
const licenseHeaderPrefix = "// The MIT License (MIT)"
const cadenceCopyright = "// Copyright (c) 2017-2020 Uber Technologies Inc."
const cadenceModificationHeader = "// Modifications Copyright (c) 2020 Uber Technologies Inc."
const temporalCopyright = "// Copyright (c) 2020 Temporal Technologies, Inc."
const temporalPartialCopyright = "// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc."

const firstLinesToCheck = 10

var (
	// directories to be excluded
	dirDenylist = []string{"vendor/"}
	// default perms for the newly created files
	defaultFilePerms = os.FileMode(0644)
)

// command line utility that adds license header
// to the source files. Usage as follows:
//
//  ./cmd/tools/copyright/licensegen.go
func main() {

	var cfg config
	flag.StringVar(&cfg.rootDir, "rootDir", ".", "project root directory")
	flag.BoolVar(&cfg.verifyOnly, "verifyOnly", false,
		"don't automatically add headers, just verify all files")
	flag.BoolVar(&cfg.temporalAddMode, "temporalAddMode", false, "add copyright for new file copied from temporal")
	flag.BoolVar(&cfg.temporalModifyMode, "temporalModifyMode", false, "add copyright for existing file which has parts copied from temporal")
	flag.StringVar(&cfg.filePaths, "filePaths", "", "comma separated list of files to run temporal license on")
	flag.Parse()

	if err := verifyCfg(cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	task := newAddLicenseHeaderTask(&cfg)
	if err := task.run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func verifyCfg(cfg config) error {
	if cfg.verifyOnly {
		if cfg.temporalModifyMode || cfg.temporalAddMode {
			return errors.New("invalid config, can only specify one of temporalAddMode, temporalModifyMode or verifyOnly")
		}
	}
	if cfg.temporalAddMode && cfg.temporalModifyMode {
		return errors.New("invalid config, can only specify temporalAddMode or temporalModifyMode")
	}
	if (cfg.temporalModifyMode || cfg.temporalAddMode) && len(cfg.filePaths) == 0 {
		return errors.New("invalid config, when running in temporalAddMode or temporalModifyMode must provide filePaths")
	}
	return nil
}

func newAddLicenseHeaderTask(cfg *config) *addLicenseHeaderTask {
	return &addLicenseHeaderTask{
		config: cfg,
	}
}

func (task *addLicenseHeaderTask) run() error {
	data, err := ioutil.ReadFile(task.config.rootDir + "/" + licenseFileName)
	if err != nil {
		return fmt.Errorf("error reading license file, errr=%v", err.Error())
	}
	task.license, err = commentOutLines(string(data))
	if err != nil {
		return fmt.Errorf("copyright header failed to comment out lines, err=%v", err.Error())
	}
	if task.config.temporalAddMode {
		task.license = fmt.Sprintf("%v\n\n%v\n\n%v", cadenceModificationHeader, temporalCopyright, task.license)
	} else if task.config.temporalModifyMode {
		task.license = fmt.Sprintf("%v\n\n%v\n\n%v", cadenceCopyright, temporalPartialCopyright, task.license)
	}
	if task.config.temporalModifyMode || task.config.temporalAddMode {
		filePaths, fileInfos, err := getFilePaths(task.config.filePaths)
		if err != nil {
			return err
		}
		for i := 0; i < len(filePaths); i++ {
			if err := task.handleFile(filePaths[i], fileInfos[i], nil); err != nil {
				return err
			}
		}
		return nil
	}
	task.license = fmt.Sprintf("%v\n\n%v\n\n%v", licenseHeaderPrefix, cadenceCopyright, task.license)
	err = filepath.Walk(task.config.rootDir, task.handleFile)
	if err != nil {
		return fmt.Errorf("copyright header check failed, err=%v", err.Error())
	}
	return nil
}

func (task *addLicenseHeaderTask) handleFile(path string, fileInfo os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), "_vendor-") {
		return filepath.SkipDir
	}

	if fileInfo.IsDir() {
		return nil
	}

	if !mustProcessPath(path) {
		return nil
	}

	if !strings.HasSuffix(fileInfo.Name(), ".go") && !strings.HasSuffix(fileInfo.Name(), ".proto") {
		return nil
	}

	// Used as part of the cli to write licence headers on files, does not use user supplied input so marked as nosec
	// #nosec
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	ok, err := hasCopyright(f)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if ok {
		if task.config.temporalModifyMode || task.config.temporalAddMode {
			return fmt.Errorf("when running in temporalModifyMode or temporalAddMode please first remove existing license header: %v", path)
		}
		return nil
	}

	// at this point, src file is missing the header
	if task.config.verifyOnly {
		if !isFileAutogenerated(path) {
			return fmt.Errorf("%v missing license header", path)
		}
	}

	// Used as part of the cli to write licence headers on files, does not use user supplied input so marked as nosec
	// #nosec
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(task.license+string(data)), defaultFilePerms)
}

func hasCopyright(f *os.File) (bool, error) {
	scanner := bufio.NewScanner(f)
	lineSuccess := scanner.Scan()
	if !lineSuccess {
		return false, fmt.Errorf("fail to read first line of file %v", f.Name())
	}
	i := 0
	for i < firstLinesToCheck && lineSuccess {
		i++
		line := strings.TrimSpace(scanner.Text())
		if err := scanner.Err(); err != nil {
			return false, err
		}
		if lineHasCopyright(line) {
			return true, nil
		}
		lineSuccess = scanner.Scan()
	}
	return false, nil
}

func isFileAutogenerated(path string) bool {
	return strings.HasPrefix(path, ".gen")
}

func mustProcessPath(path string) bool {
	for _, d := range dirDenylist {
		if strings.HasPrefix(path, d) {
			return false
		}
	}
	return true
}

func commentOutLines(str string) (string, error) {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(str))
	for scanner.Scan() {
		lines = append(lines, "// "+scanner.Text()+"\n")
	}
	lines = append(lines, "\n")

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, ""), nil
}

func lineHasCopyright(line string) bool {
	return strings.Contains(line, licenseHeaderPrefixOld) ||
		strings.Contains(line, licenseHeaderPrefix)
}

func getFilePaths(filePaths string) ([]string, []os.FileInfo, error) {
	paths := strings.Split(filePaths, ",")
	var fileInfos []os.FileInfo
	for _, p := range paths {
		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, nil, err
		}
		fileInfos = append(fileInfos, fileInfo)
	}
	return paths, fileInfos, nil
}
