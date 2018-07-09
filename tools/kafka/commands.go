package kafka

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/fatih/color"
	"github.com/uber/cadence/common/messaging"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

type (
	// ClustersConfig describes the kafka clusters
	ClustersConfig struct {
		Clusters map[string]messaging.ClusterConfig
	}
)

var (
	colorRed     = color.New(color.FgRed).SprintFunc()
	colorMagenta = color.New(color.FgMagenta).SprintFunc()
	colorGreen   = color.New(color.FgGreen).SprintFunc()
)

func mergeDLQ(c *cli.Context) {

}

func purgeTopic(c *cli.Context) {

}

func resetTopic(c *cli.Context) {

}

// load the kafka cluster config file from host and convert it to a useable config
func loadClusterConfig(clusterFile string) (map[string]messaging.ClusterConfig, error) {
	contents, err := ioutil.ReadFile(clusterFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load kafka cluster info from %v., error: %v", clusterFile, err)
	}
	clustersConfig := ClustersConfig{}
	if err := yaml.Unmarshal(contents, &clustersConfig); err != nil {
		return nil, err
	}
	return clustersConfig.Clusters, nil
}

// ErrorAndExit print easy to understand error msg first then error detail in a new line
func ErrorAndExit(msg string, err error) {
	fmt.Printf("%s %s\n%s %+v\n", colorRed("Error:"), msg, colorMagenta("Error Details:"), err)
	os.Exit(1)
}
