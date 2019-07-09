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

package provider

import (
	"errors"

	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/filestore"
)

var (
	// ErrUnknownScheme is the error for unknown archiver scheme
	ErrUnknownScheme = errors.New("unknown archiver scheme")
	// ErrEmptyBootStrapContainer is the error for empty bootstrap container
	ErrEmptyBootStrapContainer = errors.New("empty bootstrap container")
)

type (
	// ArchiverProvider returns history or visibility archiver based on the scheme and serviceName.
	// The archiver for each combination of scheme and serviceName will be created only once and cached.
	ArchiverProvider interface {
		GetHistoryArchiver(scheme string, serviceName string, container *archiver.HistoryBootstrapContainer) (archiver.HistoryArchiver, error)
		GetVisibilityArchiver(scheme string, serviceName string, container *archiver.VisibilityBootstrapContainer) (archiver.VisibilityArchiver, error)
	}

	// HistoryArchiverConfigs contain config for all implementations of the HistoryArchiver interface
	HistoryArchiverConfigs struct {
		FileStore *filestore.HistoryArchiverConfig
	}

	// VisibilityArchiverConfigs contain config for all implementations of the VisibilityArchiver interface
	VisibilityArchiverConfigs struct {
		FileStore *filestore.VisibilityArchiverConfig
	}

	archiverProvider struct {
		historyArchiverConfigs    *HistoryArchiverConfigs
		visibilityArchiverConfigs *VisibilityArchiverConfigs

		historyArchivers    map[string]archiver.HistoryArchiver
		visibilityArchivers map[string]archiver.VisibilityArchiver
	}
)

// NewArchiverProvider returns a new Archiver provider
func NewArchiverProvider(
	historyArchiverConfigs *HistoryArchiverConfigs,
	visibilityArchiverConfigs *VisibilityArchiverConfigs,
) ArchiverProvider {
	return &archiverProvider{
		historyArchiverConfigs:    historyArchiverConfigs,
		visibilityArchiverConfigs: visibilityArchiverConfigs,
	}
}

func (p *archiverProvider) GetHistoryArchiver(scheme, serviceName string, container *archiver.HistoryBootstrapContainer) (archiver.HistoryArchiver, error) {
	if historyArchiver, ok := p.historyArchivers[serviceName]; ok {
		return historyArchiver, nil
	}
	if container == nil {
		return nil, ErrEmptyBootStrapContainer
	}

	switch scheme {
	case filestore.URIScheme:
		p.historyArchivers[serviceName] = filestore.NewHistoryArchiver(*container, p.historyArchiverConfigs.FileStore)
		return p.historyArchivers[serviceName], nil
	}
	return nil, ErrUnknownScheme
}

func (p *archiverProvider) GetVisibilityArchiver(scheme, serviceName string, container *archiver.VisibilityBootstrapContainer) (archiver.VisibilityArchiver, error) {
	if visibilityArchiver, ok := p.visibilityArchivers[serviceName]; ok {
		return visibilityArchiver, nil
	}
	if container == nil {
		return nil, ErrEmptyBootStrapContainer
	}

	switch scheme {
	case filestore.URIScheme:
		p.visibilityArchivers[serviceName] = filestore.NewVisibilityArchiver(*container, p.visibilityArchiverConfigs.FileStore)
		return p.visibilityArchivers[serviceName], nil
	}
	return nil, ErrUnknownScheme
}
