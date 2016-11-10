package main

import (
	"github.com/spf13/cobra"
	"github.com/racker/rackspace-monitoring-poller/types"
)

var (
	configFilePath string
	serveCmd       = &cobra.Command{
		Use:   "serve",
		Short: "Start the service",
		Long:  "Start the service",
		Run:   serveCmdRun,
	}
)

func init() {
	serveCmd.Flags().StringVar(&configFilePath, "config", "", "Path to a file containing the config, used in "+types.DefaultConfigPathLinux)
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
