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

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/filestore"
	"github.com/uber/cadence/common/log/tag"
)

var (
	// ErrUnknownScheme is the error for unknown archiver scheme
	ErrUnknownScheme = errors.New("unknown archiver scheme")
	// ErrUnknownServiceName is the error for unknown cadence service name
	ErrUnknownServiceName = errors.New("unknown service name")
)

type (
	// Provider returns history or visiblity archiver based on the scheme and serviceName.
	// The archiver for each combination of scheme and serviceName will be created only once and cached.
	Provider interface {
		GetHistoryArchiver(scheme string, serviceName string) (archiver.HistoryArchiver, error)
		GetVisibilityArchiver(scheme string, serviceName string) (archiver.VisibilityArchiver, error)
	}

	// HistoryArchiverConfigs contain config for all implementations of the HistoryArchiver interface
	HistoryArchiverConfigs struct {
		FileStore *filestore.HistoryArchiverConfig
	}

	// VisibilityArchiverConfigs contain config for all implementations of the VisibilityArchiver interface
	VisibilityArchiverConfigs struct {
	}

	provider struct {
		container                 *archiver.BootstrapContainer
		historyArchiverConfigs    *HistoryArchiverConfigs
		visibilityArchiverConfigs *VisibilityArchiverConfigs

		historyArchivers    map[string]archiver.HistoryArchiver
		visibilityArchivers map[string]archiver.VisibilityArchiver
	}
)

// NewArchiverProvider returns a new Archiver provider
func NewArchiverProvider(
	container *archiver.BootstrapContainer,
	historyArchiverConfigs *HistoryArchiverConfigs,
	visibilityArchiverConfigs *VisibilityArchiverConfigs,
) Provider {
	return &provider{
		container:                 container,
		historyArchiverConfigs:    historyArchiverConfigs,
		visibilityArchiverConfigs: visibilityArchiverConfigs,
	}
}

func (p *provider) GetHistoryArchiver(scheme, serviceName string) (archiver.HistoryArchiver, error) {
	key := p.getArchiverKey(scheme, serviceName)
	if archiver, ok := p.historyArchivers[key]; ok {
		return archiver, nil
	}

	var componentTag tag.Tag
	switch serviceName {
	case common.HistoryServiceName:
		componentTag = tag.ComponentTimerQueue
	case common.WorkerServiceName:
		componentTag = tag.ComponentArchiver
	default:
		return nil, ErrUnknownServiceName
	}

	switch scheme {
	case filestore.URIScheme:
		container := *p.container
		container.Logger = container.Logger.WithTags(tag.Service(serviceName), componentTag)
		return filestore.NewHistoryArchiver(container, p.historyArchiverConfigs.FileStore), nil
	}
	return nil, ErrUnknownScheme
}

func (p *provider) GetVisibilityArchiver(scheme, serviceName string) (archiver.VisibilityArchiver, error) {
	return nil, errors.New("GetVisibilityArchiver is not implemented")
}

func (p *provider) getArchiverKey(scheme, serviceName string) string {
	return scheme + ":" + serviceName
}
