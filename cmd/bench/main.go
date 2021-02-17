// Copyright (c) 2017-2021 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"log"

	"github.com/uber/cadence/bench"
	"github.com/uber/cadence/bench/lib"
	"github.com/uber/cadence/common/service/config"
)

const (
	defaultEnv       = "development"
	defaultConfigDir = "config/bench"
	defaultZone      = ""
)

func main() {
	// TODO: implement bench CLI for specifying env, config dir and zone

	var cfg lib.Config
	if err := config.Load(defaultEnv, defaultConfigDir, defaultZone, &cfg); err != nil {
		log.Fatal("Failed to load config file: ", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid config: ", err)
	}

	benchWorker, err := bench.NewWorker(&cfg)
	if err != nil {
		log.Fatal("Failed to initialize bench worker: ", err)
	}

	if err := benchWorker.Run(); err != nil {
		log.Fatal("Failed to run bench worker: ", err)
	}
}
