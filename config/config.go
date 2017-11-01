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
	log "github.com/sirupsen/logrus"
	"text/template"
	"net/url"
	errors2 "github.com/pkg/errors"
)

var (
	ErrorNoZones = errors.New("No zones are defined")
	ErrorNoToken = errors.New("No token is defined")
	prefix       = "config"
)

const (
	DefaultTimeoutRead            = 10 * time.Second
	DefaultTimeoutWrite           = 10 * time.Second
	DefaultTimeoutPrepareEnd      = 60 * time.Second
	DefaultTimeoutAuth            = 30 * time.Second
	DefaultAgentId                = "-poller-"
	DefaultReconnectMinBackoff    = 25 * time.Second
	DefaultReconnectMaxBackoff    = 180 * time.Second
	DefaultReconnectFactorBackoff = 2
)

type Feature struct {
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

type Config struct {
	// Addresses
	UseSrv     bool
	UseStaging bool
	SrvQueries []string
	Addresses  []string
	ProxyUrl   *url.URL

	// Agent Info
	AgentId        string
	AgentName      string
	Features       []Feature
	Guid           string
	BundleVersion  string
	ProcessVersion string
	Token          string
	SnetRegion     string

	// Zones
	ZoneIds []string

	// MinBackoff the minimum backoff in seconds
	ReconnectMinBackoff time.Duration
	// MaxBackoff the maximum backoff in seconds
	ReconnectMaxBackoff time.Duration
	// FactorBackoff the factor for the backoff
	ReconnectFactorBackoff float64

	// Timeouts
	TimeoutRead  time.Duration
	TimeoutWrite time.Duration
	TimeoutAuth  time.Duration
	// TimeoutPrepareEnd declares the max time to elapse between poller.prepare and poller.prepare.end, but
	// is reset upon receipt of each poller.prepare.block.
	TimeoutPrepareEnd time.Duration

	// If configured, then metrics will be pushed to a Prometheus push gateway at the given URI.
	// The URI may either have a scheme of "srv" or "tcp". A "srv" refers to a DNS SRV name and "tcp" conveys
	// a host:port address, typically for local/onsite usage. The given service name will be qualified by the
	// service "_prometheus" and proto "_tcp".
	PrometheusUri string

	// If configured, all check's metrics will be distributed to the configured statsd endpoint.
	// The address is formatted as host:port.
	StatsdEndpoint string
}

type configEntry struct {
	Name      string
	ValuePtr  interface{}
	Tweak     func()
	Allowed   []string
	Sensitive bool
}

func NewConfig(guid string, useStaging bool, features []Feature) *Config {
	cfg := &Config{}
	if features != nil {
		cfg.Features = features
	} else {
		cfg.Features = make([]Feature, 0)
	}
	cfg.Guid = guid
	cfg.Token = os.Getenv("AGENT_TOKEN")
	cfg.AgentId = os.Getenv("AGENT_ID")
	cfg.AgentName = "remote_poller"
	cfg.ProcessVersion = version.Version
	cfg.BundleVersion = version.Version
	cfg.ReconnectMinBackoff = DefaultReconnectMinBackoff
	cfg.ReconnectMaxBackoff = DefaultReconnectMaxBackoff
	cfg.ReconnectFactorBackoff = DefaultReconnectFactorBackoff
	cfg.TimeoutRead = DefaultTimeoutRead
	cfg.TimeoutWrite = DefaultTimeoutWrite
	cfg.TimeoutPrepareEnd = DefaultTimeoutPrepareEnd
	cfg.TimeoutAuth = DefaultTimeoutAuth
	cfg.UseStaging = useStaging
	if useStaging {
		cfg.SrvQueries = DefaultStagingSrvEndpoints
		log.WithField("prefix", prefix).Warn("Using staging endpoints")
	} else {
		cfg.SrvQueries = DefaultProdSrvEndpoints
	}
	if cfg.AgentId == "" {
		cfg.AgentId = DefaultAgentId
	}
	cfg.UseSrv = true
	return cfg
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
		if strings.TrimSpace(line) == "" {
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

	log.WithFields(log.Fields{
		"prefix": prefix,
		"file":   filepath,
	}).Info("Loaded configuration")
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
			Name:      "monitoring_token",
			ValuePtr:  &cfg.Token,
			Sensitive: true,
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
		{
			Name:     "monitoring_proxy_url",
			ValuePtr: &cfg.ProxyUrl,
		},
		{
			Name:     "prometheus_uri",
			ValuePtr: &cfg.PrometheusUri,
		},
		{
			Name:     "statsd_endpoint",
			ValuePtr: &cfg.StatsdEndpoint,
		},
	}
}

func ApplyMask(entry *configEntry, data interface{}) interface{} {
	if entry.Sensitive {
		switch data.(type) {
		case []string:
			return strings.Repeat("*", 10)
		case string:
			return strings.Repeat("*", len(data.(string)))
		}
	}
	return data
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
				log.WithFields(log.Fields{
					"prefix": prefix,
					"name":   entry.Name,
					"value":  ApplyMask(&entry, *valuePtr),
				}).Debug("Setting configuration field")

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
				log.WithFields(log.Fields{
					"prefix": prefix,
					"name":   entry.Name,
					"value":  ApplyMask(&entry, *valuePtr),
				}).Debug("Setting configuration field")

			case **url.URL:
				// using ParseRequestURI rather than Parse since it is stricter about absolute URI or paths
				parsed, err := url.ParseRequestURI(fields[1])
				if err != nil {
					return errors2.WithMessage(err, fmt.Sprintf("%s is not a valid URL", entry.Name))
				}
				*valuePtr = parsed

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
	return time.Now().Add(offset)
}

func (cfg *Config) ComputeWriteDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutWrite
	return time.Now().Add(offset)
}

func IsUsingStaging() bool {
	return os.Getenv(EnvStaging) == EnabledEnvOpt
}

func IsUsingCleartext() bool {
	return os.Getenv(EnvCleartext) == EnabledEnvOpt
}
