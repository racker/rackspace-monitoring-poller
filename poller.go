package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	pollerCmd = &cobra.Command{
		Use: "poller",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initEnv()
		},
	}
	globalFlags struct {
		Debug bool
	}
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	})
	log.SetOutput(os.Stderr)
	pollerCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "Enable debug")
}

func initEnv() {
	if globalFlags.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	pollerCmd.AddCommand(serveCmd)
	pollerCmd.Execute()
}
