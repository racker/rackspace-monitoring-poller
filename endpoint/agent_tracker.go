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


package endpoint

import (
	"github.com/racker/rackspace-monitoring-poller/protocol"
	log "github.com/Sirupsen/logrus"

	"fmt"
	"encoding/json"
	"github.com/racker/rackspace-monitoring-poller/config"
	"sync/atomic"
	"path"
	"os"
	"strings"
	"io/ioutil"
)

type agent struct {
	responder      *json.Encoder
	outMsgId       uint64
	sourceId       string
	errors         chan <- error

	// observedToken is the token that was provided by the poller during hello handshake. It may not necessarily be valid.
	observedToken  string

	id             string
	name           string
	processVersion string
	bundleVersion  string
	features       []map[string]string

	zones          []string
}

type registration struct {
	sourceId string
	zones    []string
}

type metricsToStore struct {
	agent  *agent
	params protocol.MetricsPostRequestParams
}

type AgentTracker struct {
	cfg           config.EndpointConfig
	// key is FrameMsgCommon.Source. All access to this map must be performed in the AgentTracker.start go routine
	agents        map[string]*agent

	greeter       chan *agent
	registrations chan registration
	metricReqs    chan *protocol.MetricsPostRequest
	metrics chan *metricsToStore

	metricsRouter *MetricsRouter
}

func (at *AgentTracker) Start(cfg config.EndpointConfig) {
	at.cfg = cfg;

	// channels
	//TODO make channel buffer sizes configurable
	at.greeter = make(chan *agent, 10)
	at.registrations = make(chan registration, 10)
	at.metricReqs = make(chan *protocol.MetricsPostRequest, 50)
	at.metrics = make(chan *metricsToStore, 50)

	// LUTs
	at.agents = make(map[string]*agent)

	go at.start()
	go at.startMetricsDecomposer()
}

func (at *AgentTracker) UseMetricsRouter(mr *MetricsRouter) {
	at.metricsRouter = mr
}

// Returns a channel where errors observed about this agent are conveyed
func (at *AgentTracker) ProcessHello(req *protocol.HandshakeRequest, responder *json.Encoder) <-chan error {
	errors := make(chan error)

	newAgent := &agent{
		sourceId: req.Source,
		observedToken: req.Params.Token,
		id: req.Params.AgentId,
		name: req.Params.AgentName,
		processVersion: req.Params.ProcessVersion,
		bundleVersion: req.Params.BundleVersion,
		features: req.Params.Features,
		errors: errors,
		responder: responder,
	}

	at.greeter <- newAgent

	return errors
}

func (at *AgentTracker) ProcessPollerRegister(req *protocol.PollerRegister) {
	reg := registration{
		sourceId: req.Source,
		zones: req.Params[protocol.PollerZones],
	}

	at.registrations <- reg
}

func (at *AgentTracker) ProcessCheckMetricsPost(req *protocol.MetricsPostRequest) {
	at.metricReqs <- req
}

func (at *AgentTracker) start() {
	log.Infoln("Starting agent tracker")

	for {
		select {
		case greetedAgent := <-at.greeter:
			at.handleGreetedAgent(greetedAgent)

		case reg := <-at.registrations:
			at.handlePollerRegistration(reg)

		case req := <-at.metricReqs:
			at.handleMetricsPostReq(req)
		}
	}
}

func (at *AgentTracker) startMetricsDecomposer() {
	log.Infoln("Starting metrics decomposer")

	for {
		metric := <-at.metrics
		log.WithFields(log.Fields{
			"agentId": metric.agent.id,
			"params": metric.params,
		}).Info("Decomposing metric")

		//TODO do something with failed metrics that only contain state and status

		for _, outerWraps := range metric.params.Metrics {

			for _, metricWrap := range outerWraps {
				for field, entry := range metricWrap {
					metricName := BuildMetricName(metric.params.EntityId,
						metric.agent.id,
						metric.params.CheckType,
						metric.params.CheckId,
						field,
					)

					m := Metric{
						Name: metricName,
						Value: entry.Value,
						MetricType: entry.Type,
					}

					at.metricsRouter.Route(m)
				}

			}

		}

	}
}

func (at *AgentTracker) handleMetricsPostReq(req *protocol.MetricsPostRequest) {
	a, ok := at.agents[req.Source]
	if !ok {
		log.WithField("sourceId", req.Source).Warn("Trying to handle metrics from unknown source")
		return
	}

	metric := &metricsToStore{
		agent: a,
		params: req.Params,
	}

	at.metrics <- metric
}

func (at *AgentTracker) handlePollerRegistration(reg registration) {
	log.Debugln("Handling poller registration", registration{})
	a, ok := at.agents[reg.sourceId]
	if !ok {
		log.WithField("sourceId", reg.sourceId).Warn("Trying to register poller for unknown source")
		return
	}

	a.zones = reg.zones

	go at.lookupChecks(a)
}

func (at *AgentTracker) lookupChecks(a *agent) {
	log.WithField("agent", a).Debug("Looking up checks to apply to agent")

	//TODO iterate through zones in poller registration
	pathToChecks := path.Join(at.cfg.AgentsConfigDir, a.zones[0], a.id, "checks")

	checksDir, err := os.Open(pathToChecks)
	if err != nil {
		log.WithField("pathToChecks", pathToChecks).Warn("Unable to access agent checks directory")
		return
	}
	defer checksDir.Close()

	contents, err := checksDir.Readdir(0)
	if err != nil {
		log.WithField("pathToChecks", pathToChecks).Warn("Unable to read agent checks directory")
		return
	}

	// Look for all .json files in the checks directory...and assume they are valid check details
	//TODO validate check details
	for _, fileInfo := range contents {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".json") {
			jsonContent, err := ioutil.ReadFile(path.Join(pathToChecks, fileInfo.Name()))
			if err != nil {
				log.WithFields(log.Fields{
					"checksDir": checksDir,
					"file": fileInfo,
				}).Warn("Unable to read checks file")
				continue
			}

			go a.sendTo(protocol.MethodPollerChecksAdd, jsonContent)
		}
	}
}

func (at *AgentTracker) handleGreetedAgent(greetedAgent *agent) {
	log.WithField("agent", greetedAgent).Debug("Handling greeted agent")

	sourceId := greetedAgent.sourceId
	if _, exists := at.agents[sourceId]; exists {
		log.WithFields(log.Fields{
			"sourceId": sourceId,
		}).Warn("Agent greeting already observed")
		greetedAgent.errors <- AgentTrackingError{
			SourceId: sourceId,
			Message: "already observed",
		}
		return
	}

	at.agents[sourceId] = greetedAgent
}

func (a *agent) sendTo(method string, params interface{}) {
	msgId := atomic.AddUint64(&a.outMsgId, 1)

	var rawParams json.RawMessage

	switch params.(type) {
	case []byte:
		// already marshalled
		rawParams = params.([]byte)

	default:
		// needs marshalling
		var err error
		rawParams, err = json.Marshal(params)
		if err != nil {
			log.WithField("params", params).Warn("Unable to marshal params")
			return
		}
	}

	frameOut := protocol.FrameMsg{
		FrameMsgCommon: protocol.FrameMsgCommon{
			Version: "1",
			Id: msgId,
			Source: "endpoint",
			Target: "endpoint",
			Method: method,
		},
		RawParams: rawParams,
	}

	log.WithFields(log.Fields{
		"msgId": frameOut.Id,
		"method": frameOut.Method,
		"agentId": a.id,
	}).Debug("SENDing frame to agent")

	a.responder.Encode(&frameOut)
}

type AgentTrackingError struct {
	Message  string
	SourceId string
}

func (e AgentTrackingError) Error() string {
	return fmt.Sprintf("An error occurred while tracking an agent %s: %s", e.SourceId, e.Message)
}
