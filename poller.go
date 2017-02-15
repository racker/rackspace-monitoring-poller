//
// Copyright 2016 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Main entry point for the Rackspace Monitoring Poller application.
//
// Sub-entry points are declared in the commands package, such as
//
//   rackspace-monitoring-poller serve ...
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/commands"
	"github.com/racker/rackspace-monitoring-poller/version"
	"github.com/spf13/cobra"
	"os/signal"
	"syscall"
)

const (
	DefaultLogfileName = "/var/log/rackspace-monitoring-poller.log"
)

var (
	pollerCmd = &cobra.Command{
		Use: "poller",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initEnv()
		},
	}
	globalFlags struct {
		Debug       bool
		LogfileName string
		NoLogfile   bool
	}

	// Formatter is a log formatter utilized for poller.  Defaulted to JSONFormatter
	// due to simplicity for parsing by 3rd party logging tools
	Formatter log.Formatter = &log.JSONFormatter{
		TimestampFormat: time.RFC1123,
	}
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetOutput(os.Stderr)
	pollerCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "Enable debug")
	pollerCmd.PersistentFlags().BoolVar(&globalFlags.NoLogfile, "no-logfile", false, "Logs to stdout instead of a logfile")
	pollerCmd.PersistentFlags().StringVarP(&globalFlags.LogfileName, "logfile", "l",
		DefaultLogfileName, "Location of the log file")
}

func initEnv() {
	if globalFlags.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if os.Getenv("LOG_TEXT_FORMAT") == "true" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	} else {
		log.SetFormatter(Formatter)
	}

	if !globalFlags.NoLogfile && globalFlags.LogfileName != "" {
		log.WithField("location", globalFlags.LogfileName).Info("Redirecting log output")
		setLogOutput()

		hupChan := make(chan os.Signal, 1)
		signal.Notify(hupChan, os.Interrupt, syscall.SIGHUP)
		go func() {
			for {
				//TODO select on a cancelling context channel, etc in order to exit

				<-hupChan
				setLogOutput()
			}
		}()
	}
}

func setLogOutput() {
	file, err := os.OpenFile(globalFlags.LogfileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open initial log file %v : %v", globalFlags.LogfileName, err)
	}
	log.SetOutput(file)

}

func main() {
	pollerCmd.AddCommand(versionCmd)
	pollerCmd.AddCommand(commands.ServeCmd)
	pollerCmd.AddCommand(commands.EndpointCmd)
	pollerCmd.Execute()
}
