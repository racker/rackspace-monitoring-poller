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
	log "github.com/Sirupsen/logrus"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	ErrorNoZones = errors.New("No zones are defined")
	ErrorNoToken = errors.New("No token is defined")
)

type Config struct {
	// Addresses
	UseSrv     bool
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

	// Zones
	ZoneIds []string

	// Timeouts
	TimeoutRead  time.Duration
	TimeoutWrite time.Duration
}

func NewConfig(guid string, useStaging bool) *Config {
	cfg := &Config{}
	cfg.init()
	cfg.Guid = guid
	cfg.Token = os.Getenv("AGENT_TOKEN")
	cfg.AgentId = os.Getenv("AGENT_ID")
	cfg.AgentName = "remote_poller"
	cfg.ProcessVersion = "0.0.1" //TODO
	cfg.BundleVersion = "0.0.1"  //TODO
	cfg.TimeoutRead = time.Duration(10 * time.Second)
	cfg.TimeoutWrite = time.Duration(10 * time.Second)
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

func (cfg *Config) LoadFromFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	regexComment, _ := regexp.Compile("^#")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if regexComment.MatchString(line) {
			continue
		}
		fields := strings.Fields(line)
		err := cfg.ParseFields(fields)
		if err != nil {
			continue
		}
	}
	log.WithField("file", filepath).Info("Loaded configuration")
	return nil
}

func (cfg *Config) ParseFields(fields []string) error {
	if len(fields) < 2 {
		return errors.New("Invalid fields length")
	}
	switch fields[0] {
	case "monitoring_id":
		cfg.AgentId = fields[1]
		log.Printf("cfg: Setting Monitoring Id: %s", cfg.AgentId)
	case "monitoring_token":
		cfg.Token = fields[1]
		log.Printf("cfg: Setting Token")
	case "monitoring_private_zones":
		zones := make([]string, 0)
		for _, zone := range strings.Split(fields[1], ",") {
			zones = append(zones, strings.TrimSpace(zone))
		}
		log.Printf("cfg: Setting Zones: %s", strings.Join(zones, ", "))
		cfg.ZoneIds = zones
	case "monitoring_endpoints":
		cfg.Addresses = strings.Split(fields[1], ",")
		cfg.UseSrv = false
		log.Printf("cfg: Setting Endpoints: %s", fields[1])
	}

	return nil
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

func (cfg *Config) GetReadDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutRead
	return time.Now().Add(time.Duration(offset * time.Second))
}

func (cfg *Config) GetWriteDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutWrite
	return time.Now().Add(time.Duration(offset * time.Second))
}
