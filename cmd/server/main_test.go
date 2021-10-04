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
	"github.com/uber/cadence/tools/sql"
)

func TestServerStartup(t *testing.T) {
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
	log.Printf("config=\n%v\n", cfg.String())

	cfg.DynamicConfig.FileBased.Filepath = cadence.ConstructPathIfNeed(rootDir, cfg.DynamicConfig.FileBased.Filepath)

	if err := cfg.ValidateAndFillDefaults(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}
	// cassandra schema version validation
	if err := cassandra.VerifyCompatibleVersion(cfg.Persistence); err != nil {
		log.Fatal("cassandra schema version compatibility check failed: ", err)
	}
	// sql schema version validation
	if err := sql.VerifyCompatibleVersion(cfg.Persistence); err != nil {
		log.Fatal("sql schema version compatibility check failed: ", err)
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
