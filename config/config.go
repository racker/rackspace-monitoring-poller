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

// Package config declares the data structures used for all execution entry points
package config

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/racker/rackspace-monitoring-poller/version"

	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"text/template"
)

var (
	ErrorNoZones = errors.New("No zones are defined")
	ErrorNoToken = errors.New("No token is defined")
)

const (
	DefaultTimeoutRead       = 10 * time.Second
	DefaultTimeoutWrite      = 10 * time.Second
	DefaultTimeoutPrepareEnd = 60 * time.Second
)

type Config struct {
	// Addresses
	UseSrv     bool
	UseStaging bool
	SrvQueries []string
	Addresses  []string

	// Agent Info
	AgentId        string
	AgentName      string
	Features       []map[string]string
	Guid           string
	BundleVersion  string
	ProcessVersion string
	Token          string
	SnetRegion     string

	// Zones
	ZoneIds []string

	// Timeouts
	TimeoutRead  time.Duration
	TimeoutWrite time.Duration
	// TimeoutPrepareEnd declares the max time to elapse between poller.prepare and poller.prepare.end, but
	// is reset upon receipt of each poller.prepare.block.
	TimeoutPrepareEnd time.Duration
}

type configEntry struct {
	Name     string
	ValuePtr interface{}
	Tweak    func()
	Allowed  []string
}

func NewConfig(guid string, useStaging bool) *Config {
	cfg := &Config{}
	cfg.init()
	cfg.Guid = guid
	cfg.Token = os.Getenv("AGENT_TOKEN")
	cfg.AgentId = os.Getenv("AGENT_ID")
	cfg.AgentName = "remote_poller"
	cfg.ProcessVersion = version.Version
	cfg.BundleVersion = version.Version
	cfg.TimeoutRead = DefaultTimeoutRead
	cfg.TimeoutWrite = DefaultTimeoutWrite
	cfg.TimeoutPrepareEnd = DefaultTimeoutPrepareEnd
	cfg.UseStaging = useStaging
	if useStaging {
		cfg.SrvQueries = DefaultStagingSrvEndpoints
		log.Warn("Using staging endpoints")
	} else {
		cfg.SrvQueries = DefaultProdSrvEndpoints
	}
	cfg.UseSrv = true
	return cfg
}

func (cfg *Config) init() {
	cfg.Features = make([]map[string]string, 0)
}

// LoadFromFile populates this Config with the values defined in that file and then calls PostProcess.
func (cfg *Config) LoadFromFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	configEntries := cfg.DefineConfigEntries()

	regexComment, _ := regexp.Compile("^#")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if regexComment.MatchString(line) {
			continue
		}
		fields := strings.Fields(line)
		if err := cfg.ParseFields(configEntries, fields); err != nil {
			return err
		}
	}

	if err := cfg.PostProcess(); err != nil {
		return err
	}

	log.WithField("file", filepath).Info("Loaded configuration")
	return nil
}

func (cfg *Config) PostProcess() error {
	if cfg.SnetRegion != "" {
		if !cfg.UseStaging {
			if err := cfg.processSnetSrvTemplates(SnetMonitoringTemplateSrvQueries); err != nil {
				return err
			}
		} else {
			if err := cfg.processSnetSrvTemplates(SnetMonitoringTemplateSrvQueriesStaging); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cfg *Config) processSnetSrvTemplates(templates []*template.Template) error {
	cfg.SrvQueries = make([]string, len(templates))
	for i, t := range templates {
		buf := new(bytes.Buffer)
		if err := t.Execute(buf, cfg); err != nil {
			return err
		}
		cfg.SrvQueries[i] = buf.String()
	}

	return nil
}

func (cfg *Config) DefineConfigEntries() []configEntry {
	return []configEntry{
		{
			Name:     "monitoring_id",
			ValuePtr: &cfg.AgentId,
		},
		{
			Name:     "monitoring_token",
			ValuePtr: &cfg.Token,
		},
		{
			Name:     "monitoring_private_zones",
			ValuePtr: &cfg.ZoneIds,
		},
		{
			Name:     "monitoring_endpoints",
			ValuePtr: &cfg.Addresses,
			Tweak: func() {
				cfg.UseSrv = false
			},
		},
		{
			Name:     "monitoring_snet_region",
			ValuePtr: &cfg.SnetRegion,
			Allowed:  ValidSnetRegions,
		},
	}
}

func (cfg *Config) ParseFields(configEntries []configEntry, fields []string) error {
	if len(fields) < 2 {
		return fmt.Errorf("Invalid fields length: %v", fields)
	}

	for _, entry := range configEntries {
		if entry.Name == fields[0] {
			switch valuePtr := entry.ValuePtr.(type) {
			case *string:
				if err := entry.IsAllowed(fields[1]); err != nil {
					return fmt.Errorf("Disallowed value in %s : %v", entry.Name, err)
				}

				*valuePtr = fields[1]
			case *[]string:
				rawParts := strings.Split(fields[1], ",")
				parts := make([]string, len(rawParts))
				for i, p := range rawParts {
					v := strings.TrimSpace(p)
					if err := entry.IsAllowed(v); err != nil {
						return fmt.Errorf("Disallowed value in %s : %v", entry.Name, err)
					}
					parts[i] = v
				}
				*valuePtr = parts

			default:
				return fmt.Errorf("Unsupported config entry type for %s", entry.Name)
			}

			if entry.Tweak != nil {
				entry.Tweak()
			}
		}
	}

	return nil
}

func (e *configEntry) IsAllowed(actualValue string) error {
	if len(e.Allowed) == 0 {
		return nil
	}

	for _, a := range e.Allowed {
		if a == actualValue {
			return nil
		}
	}

	return fmt.Errorf("The value '%s' is not allowed. Exepcted %v", actualValue, e.Allowed)
}

func (cfg *Config) Validate() error {
	if len(cfg.ZoneIds) == 0 {
		return ErrorNoZones
	}
	if len(cfg.Token) == 0 {
		return ErrorNoToken
	}
	return nil
}

func (cfg *Config) SetPrivateZones(zones []string) {
	cfg.ZoneIds = zones
}

func (cfg *Config) ComputeReadDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutRead
	return time.Now().Add(time.Duration(offset * time.Second))
}

func (cfg *Config) ComputeWriteDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutWrite
	return time.Now().Add(time.Duration(offset * time.Second))
}

func IsUsingStaging() bool {
	return os.Getenv(EnvStaging) == EnabledEnvOpt
}
