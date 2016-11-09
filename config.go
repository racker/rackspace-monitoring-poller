package main

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"os"
	"regexp"
	"strings"
	"time"
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
	PrivateZones []string

	// Timeouts
	TimeoutRead  time.Duration
	TimeoutWrite time.Duration
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.init()
	cfg.Guid = "7A11A061-1A4D-4E61-AFED-8A5E2870F900" //TODO CHANGEME
	cfg.Token = os.Getenv("AGENT_TOKEN")
	cfg.AgentId = os.Getenv("AGENT_ID")
	cfg.AgentName = "remote_poller"
	cfg.ProcessVersion = "0.0.1"
	cfg.BundleVersion = "0.0.1"
	cfg.TimeoutRead = time.Duration(10 * time.Second)
	cfg.TimeoutWrite = time.Duration(10 * time.Second)
	cfg.SrvQueries = DefaultProdSrvEndpoints
	cfg.UseSrv = true
	return cfg
}

func (cfg *Config) init() {
	cfg.Features = make([]map[string]string, 0)
}

func (cfg *Config) LoadFromFile(filepath string) error {
	log.Printf("Using config file: %s", filepath)
	_, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	regexComment, _ := regexp.Compile("^#")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if regexComment.MatchString(line) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		cfg.ParseFields(fields)
	}
	return nil
}

func (cfg *Config) ParseFields(fields []string) {
	switch fields[0] {
	case "monitoring_id":
		cfg.AgentId = fields[1]
		log.Printf("cfg: Setting Monitoring Id: %s", cfg.AgentId)
	case "monitoring_token":
		cfg.Token = fields[1]
		log.Printf("cfg: Setting Token")
	case "monitoring_endpoints":
		cfg.Addresses = strings.Split(fields[1], ",")
		cfg.UseSrv = false
		log.Printf("cfg: Setting Endpoints: %s", fields[1])
	}
}

func (cfg *Config) SetPrivateZones(zones []string) {
	cfg.PrivateZones = zones
}

func (cfg *Config) GetReadDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutRead
	return time.Now().Add(time.Duration(offset * time.Second))
}

func (cfg *Config) GetWriteDeadline(offset time.Duration) time.Time {
	offset = offset + cfg.TimeoutWrite
	return time.Now().Add(time.Duration(offset * time.Second))
}
