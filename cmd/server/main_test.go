// Copyright (c) 2021 Uber Technologies, Inc.
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

// +build !race

package main

import (
	"log"
	"testing"
	"time"

	"github.com/uber/cadence/cmd/server/cadence"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/tools/cassandra"
)

/*
TestServerStartup tests the startup logic for the binary. When this fails, you should be able to reproduce by running "cadence-server start"
*/
func TestServerStartup(t *testing.T) {
	// If you want to test it locally, change it to false
	runInBuildKite := true

	env := "development"
	zone := ""
	rootDir := "../../"
	configDir := cadence.ConstructPathIfNeed(rootDir, "config")

	log.Printf("Loading config; env=%v,zone=%v,configDir=%v\n", env, zone, configDir)

	var cfg config.Config
	err := config.Load(env, configDir, zone, &cfg)
	if err != nil {
		log.Fatal("Config file corrupted.", err)
	}
	// replace local host to docker network
	if runInBuildKite {
		ds := cfg.Persistence.DataStores[cfg.Persistence.DefaultStore]
		ds.NoSQL.Hosts = "cassandra"
		cfg.Persistence.DataStores[cfg.Persistence.DefaultStore] = ds

		ds = cfg.Persistence.DataStores[cfg.Persistence.VisibilityStore]
		ds.NoSQL.Hosts = "cassandra"
		cfg.Persistence.DataStores[cfg.Persistence.VisibilityStore] = ds
	}

	log.Printf("config=\n%v\n", cfg.String())

	cfg.DynamicConfig.FileBased.Filepath = cadence.ConstructPathIfNeed(rootDir, cfg.DynamicConfig.FileBased.Filepath)

	if err := cfg.ValidateAndFillDefaults(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}
	// cassandra schema version validation
	if err := cassandra.VerifyCompatibleVersion(cfg.Persistence); err != nil {
		log.Fatal("cassandra schema version compatibility check failed: ", err)
	}

	var daemons []common.Daemon
	services := service.ShortNames(service.List)
	for _, svc := range services {
		server := cadence.NewServer(svc, &cfg)
		daemons = append(daemons, server)
		server.Start()
	}

	timer := time.NewTimer(time.Second * 10)

	<- timer.C
	for _, daemon := range daemons{
		daemon.Stop()
	}
}
