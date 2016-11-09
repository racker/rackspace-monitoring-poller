package main

import (
	"github.com/spf13/cobra"
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
	serveCmd.Flags().StringVar(&configFilePath, "config", "", "Path to a file containing the config, used in "+DefaultConfigPathLinux)
}

func serveCmdRun(cmd *cobra.Command, args []string) {
	cfg := NewConfig()
	cfg.LoadFromFile(configFilePath)
	for {
		stream := NewConnectionStream(cfg)
		stream.Connect()
		stream.Wait()
	}
}
