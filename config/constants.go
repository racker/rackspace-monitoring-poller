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

// Constants
package config

import (
	"text/template"
)

var (
	DefaultProdSrvEndpoints = []string{
		"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
		"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
		"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
	}
	DefaultStagingSrvEndpoints = []string{
		"_monitoringagent._tcp.dfw1.stage.monitoring.api.rackspacecloud.com",
		"_monitoringagent._tcp.ord1.stage.monitoring.api.rackspacecloud.com",
		"_monitoringagent._tcp.lon3.stage.monitoring.api.rackspacecloud.com",
	}
	ValidSnetRegions = []string{
		"dfw",
		"ord",
		"lon",
		"syd",
		"hkg",
		"iad",
	}
	SnetMonitoringTemplateSrvQueries = []*template.Template{
		template.Must(template.New("0").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region0.prod.monitoring.api.rackspacecloud.com")),
		template.Must(template.New("1").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region1.prod.monitoring.api.rackspacecloud.com")),
		template.Must(template.New("2").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region2.prod.monitoring.api.rackspacecloud.com")),
	}
	SnetMonitoringTemplateSrvQueriesStaging = []*template.Template{
		template.Must(template.New("0").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region0.stage.monitoring.api.rackspacecloud.com")),
		template.Must(template.New("1").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region1.stage.monitoring.api.rackspacecloud.com")),
		template.Must(template.New("2").Parse("_monitoringagent._tcp.snet-{{.SnetRegion}}-region2.stage.monitoring.api.rackspacecloud.com")),
	}
)

const (
	DefaultConfigPathLinux = "/etc/rackspace-monitoring-poller.cfg"

	DefaultPort = "50041"

	EnvStaging    = "STAGING"
	EnvDevCA      = "DEV_CA"
	EnvCleartext  = "CLEARTEXT"
	EnabledEnvOpt = "1"
)
