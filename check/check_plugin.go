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
package check

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/racker/rackspace-monitoring-poller/protocol/metric"

	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
)

type PluginCheck struct {
	Base
	protocol.PluginCheckDetails
}

func NewPluginCheck(base *Base) (Check, error) {
	check := &PluginCheck{Base: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix":  "check_plugin",
			"err":     err,
			"details": string(*base.RawDetails),
		}).Error("Unable to unmarshal check details")
		return nil, err
	}
	return check, nil
}

func (ch *PluginCheck) Run() (*ResultSet, error) {
	// Setup timeout
	timeout := uint64(ch.Details.Timeout)
	if timeout == 0 {
		timeout = ch.Timeout
	}
	ctxTimeout := time.Duration(timeout) * time.Second

	log.WithFields(log.Fields{
		"prefix":  ch.GetLogPrefix(),
		"args":    ch.Details.Args,
		"file":    ch.Details.File,
		"id":      ch.Id,
		"timeout": timeout,
	}).Debug("Running Plugin Check")

	// Set Context
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	// Setup results
	cr := NewResult()
	crs := NewResultSet(ch, cr)
	crs.SetStateUnavailable()

	// Setup stdin pipe, which gets closed
	r, _, _ := os.Pipe()

	// Command Setup
	cmd := exec.CommandContext(ctx, ch.Details.File, ch.Details.Args...)

	// Set I/O
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		crs.SetStateUnavailable()
		crs.SetStatus(err.Error())
		return crs, nil
	}
	cmd.Stdin = r

	// Start process and close stdin
	if err := cmd.Start(); err != nil {
		r.Close()
		crs.SetStateUnavailable()
		crs.SetStatus(err.Error())
		return crs, nil
	}
	r.Close()

	stdoutReadDone := make(chan struct{})
	go func() {
		defer close(stdoutReadDone)
		scanner := bufio.NewScanner(stdout)
		statusRegex, _ := regexp.Compile("^status\\s+(err|warn|ok)\\s+(.*)")
		stateRegex, _ := regexp.Compile("^state\\s+(.*?)")
		metricRegex, _ := regexp.Compile("^metric\\s+(.*)\\s+(.*)\\s+(.*)")
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			log.WithFields(log.Fields{
				"prefix": ch.GetLogPrefix() + ":stdout",
				"id":     ch.Id,
				"line":   line,
			}).Debug("output")
			if matches := statusRegex.FindStringSubmatch(line); matches != nil {
				switch strings.ToLower(matches[1]) {
				case "ok":
					fallthrough
				case "warn":
					fallthrough
				case "err":
					crs.SetStatus(matches[2])
				default:
					crs.SetStatus(strings.Join(matches[1:], " "))
				}
			}
			if matches := stateRegex.FindStringSubmatch(line); matches != nil {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					state := strings.ToLower(fields[1])
					switch state {
					case "available":
						crs.SetStateAvailable()
					case "unavailable":
						crs.SetStateUnavailable()
					}
				}
			}
			if matches := metricRegex.FindStringSubmatch(line); matches != nil {
				metricName := matches[1]
				metricType := strings.ToLower(matches[2])
				metricValue := matches[3]
				var pollerType int
				switch metricType {
				case "string":
					pollerType = metric.MetricString
				case "double":
					pollerType = metric.MetricFloat
				case "gauge":
					fallthrough
				case "int":
					fallthrough
				case "int32":
					fallthrough
				case "uint32":
					fallthrough
				case "int64":
					fallthrough
				case "uint64":
					pollerType = metric.MetricNumber
				default:
					pollerType = metric.MetricString
				}
				cr.AddMetric(metric.NewMetric(metricName, "", pollerType, metricValue, metricType))
			}
		}
	}()

	// Wait for commmand to finish
	var errorFlag bool
	if err := cmd.Wait(); err != nil {
		crs.SetStateUnavailable()
		crs.SetStatus(err.Error())
		errorFlag = true
	}
	<-stdoutReadDone

	log.WithFields(log.Fields{
		"prefix": ch.GetLogPrefix(),
		"id":     ch.Id,
		"error":  errorFlag,
	}).Debug("End Plugin Check")

	return crs, nil
}