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
	"github.com/spf13/cobra"
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

	// Formatter is a log formatter utilized for poller.  Defaulted to JSONFormatter
	// due to simplicity for parsing by 3rd party logging tools
	Formatter log.Formatter = &log.JSONFormatter{
		TimestampFormat: time.RFC1123,
	}

	Version = "1.0.0+git"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	if os.Getenv("LOG_TEXT_FORMAT") == "true" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	} else {
		log.SetFormatter(Formatter)
	}
	log.SetOutput(os.Stderr)
	pollerCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "Enable debug")
}

func initEnv() {
	if globalFlags.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	pollerCmd.AddCommand(versionCmd)
	pollerCmd.AddCommand(commands.ServeCmd)
	pollerCmd.AddCommand(commands.EndpointCmd)
	pollerCmd.Execute()
}
