/*
 *
 * Copyright 2018 Rackspace
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS-IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	UninstallServiceCmd = &cobra.Command{
		Use:   "uninstall-service",
		Short: "Uninstalls the poller as a Windows service",
		Run:   uninstallServiceCmdRun,
	}
)

func uninstallServiceCmdRun(cmd *cobra.Command, args []string) {

	s, logger, err := createServiceInstance()
	if err != nil {
		log.Fatal(err)
	}

	err = s.Uninstall()
	if err != nil {
		logger.Error(err)
	}
}
