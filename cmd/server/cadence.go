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
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/service"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/service/frontend"
	"github.com/uber/cadence/service/history"
	"github.com/uber/cadence/service/matching"
	"github.com/urfave/cli"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	historyService  = "history"
	matchingService = "matching"
	frontendService = "frontend"
)

// validServices is the list of all valid cadence services
var validServices = []string{historyService, matchingService, frontendService}

// main entry point for the cadence server
func main() {
	app := buildCLI()
	app.Run(os.Args)
}

// startHandler handles a server start command from the CLI
func startHandler(c *cli.Context) {

	env := getEnvironment(c)
	zone := getZone(c)
	configDir := getConfigDir(c)

	log.Printf("Loading config; env=%v,zone=%v,configDir=%v\n", env, zone, configDir)

	var cfg config.Config
	config.Load(env, configDir, zone, &cfg)

	log.Printf("config=\n%v\n", cfg.String())

	var wg sync.WaitGroup

	for _, s := range getServices(c) {

		_, ok := cfg.Services[s]
		if !ok {
			log.Fatalf("`%v` service missing config", s)
		}

		wg.Add(1)
		startService(&cfg, s, &wg)
	}

	wg.Wait()
}

// startService starts a service with the given name and config
func startService(cfg *config.Config, name string, wg *sync.WaitGroup) {

	var err error

	params := service.BootstrapParams{}
	params.Name = "cadence-" + name
	params.Logger = cfg.Log.NewBarkLogger()
	params.CassandraConfig = cfg.Cassandra

	params.RingpopFactory, err = cfg.Ringpop.NewFactory()
	if err != nil {
		log.Fatalf("error creating ringpop factory: %v", err)
	}

	svcCfg := cfg.Services[name]

	params.MetricScope = svcCfg.Metrics.NewScope()
	params.TChannelFactory = svcCfg.TChannel.NewFactory()

	var daemon common.Daemon

	switch name {
	case frontendService:
		daemon = frontend.NewService(&params)
	case historyService:
		daemon = history.NewService(&params)
	case matchingService:
		daemon = matching.NewService(&params)
	}

	go execDaemon(daemon, wg)
}

// execDaemon runs the daemon in a separate go routine
func execDaemon(d common.Daemon, wg *sync.WaitGroup) {
	defer wg.Done()
	d.Start()
}

func getEnvironment(c *cli.Context) string {
	return strings.TrimSpace(c.GlobalString("env"))
}

func getZone(c *cli.Context) string {
	return strings.TrimSpace(c.GlobalString("zone"))
}

// getServices parses the services arg from cli
// and returns a list of services to start
func getServices(c *cli.Context) []string {

	val := strings.TrimSpace(c.String("services"))
	tokens := strings.Split(val, ",")

	if len(tokens) == 0 {
		log.Fatalf("list of services is empty")
	}

	for _, t := range tokens {
		if !isValidService(t) {
			log.Fatalf("invalid service `%v` in service list [%v]", t, val)
		}
	}

	return tokens
}

func isValidService(in string) bool {
	for _, s := range validServices {
		if s == in {
			return true
		}
	}
	return false
}

func getConfigDir(c *cli.Context) string {
	return path(getRootDir(c), c.GlobalString("config"))
}

func getRootDir(c *cli.Context) string {
	dirpath := c.GlobalString("root")
	if len(dirpath) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("os.Getwd() failed, err=%v", err)
		}
		return cwd
	}
	return dirpath
}

func path(dir string, file string) string {
	return dir + "/" + file
}

func buildCLI() *cli.App {

	app := cli.NewApp()
	app.Name = "cadence"
	app.Usage = "Cadence server"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "root, r",
			Value:  ".",
			Usage:  "root directory of execution environment",
			EnvVar: config.EnvKeyRoot,
		},
		cli.StringFlag{
			Name:   "config, c",
			Value:  "config",
			Usage:  "config dir path relative to root",
			EnvVar: config.EnvKeyConfigDir,
		},
		cli.StringFlag{
			Name:   "env, e",
			Value:  "development",
			Usage:  "runtime environment",
			EnvVar: config.EnvKeyEnvironment,
		},
		cli.StringFlag{
			Name:   "zone, az",
			Value:  "",
			Usage:  "availability zone",
			EnvVar: config.EnvKeyAvailabilityZone,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{""},
			Usage:   "start cadence server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "services, s",
					Value: strings.Join(validServices, ","),
					Usage: "list of services to start",
				},
			},
			Action: func(c *cli.Context) {
				startHandler(c)
			},
		},
	}

	return app

}
