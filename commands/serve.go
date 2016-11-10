package commands

import (
	"github.com/racker/rackspace-monitoring-poller/types"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	ServeCmd       = &cobra.Command{
		Use:   "serve",
		Short: "Start the service",
		Long:  "Start the service",
		Run:   serveCmdRun,
	}
)

func init() {
	ServeCmd.Flags().StringVar(&configFilePath, "config", "", "Path to a file containing the config, used in "+types.DefaultConfigPathLinux)
}

func serveCmdRun(cmd *cobra.Command, args []string) {
	cfg := types.NewConfig()
	cfg.LoadFromFile(configFilePath)
	for {
		stream := types.NewConnectionStream(cfg)
		stream.Connect()
		stream.Wait()
	}
}
