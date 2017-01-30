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
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/protocol"

	"context"
	"encoding/json"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync/atomic"
)

const (
	AgentErrorChanSize    = 1
	GreeterChanSize       = 10
	RegistrationsChanSize = 10
	MetricsReqChanSize    = 50
	MetricsStoreChanSize  = 50
	CheckId               = "id"
	CheckFileSuffix       = ".json"
)

type agent struct {
	responder *utils.SmartConn
	outMsgId  uint64
	sourceId  string
	errors    chan<- error

	// observedToken is the token that was provided by the poller during hello handshake. It may not necessarily be valid.
	observedToken string

	id             string
	name           string
	processVersion string
	bundleVersion  string
	features       []map[string]string

	zones []string
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
	cfg *config.EndpointConfig
	// key is FrameMsgCommon.Source. All access to this map must be performed in the AgentTracker.start go routine
	agents map[string]*agent

	greeter       chan *agent
	registrations chan registration
	metricReqs    chan *protocol.MetricsPostRequest
	metrics       chan *metricsToStore

	metricsRouter *MetricsRouter
}

func NewAgentTracker(cfg *config.EndpointConfig) *AgentTracker {
	return &AgentTracker{
		cfg: cfg,

		agents: make(map[string]*agent),

		greeter:       make(chan *agent, GreeterChanSize),
		registrations: make(chan registration, RegistrationsChanSize),
		metricReqs:    make(chan *protocol.MetricsPostRequest, MetricsReqChanSize),
		metrics:       make(chan *metricsToStore, MetricsStoreChanSize),
	}
}

func (at *AgentTracker) Start(ctx context.Context) {
	go at.start(ctx)
	go at.startMetricsDecomposer(ctx)
}

func (at *AgentTracker) UseMetricsRouter(mr *MetricsRouter) {
	at.metricsRouter = mr
}

// Returns a channel where errors observed about this agent are conveyed
func (at *AgentTracker) ProcessHello(req protocol.HandshakeRequest, responder *utils.SmartConn) <-chan error {
	errors := make(chan error, AgentErrorChanSize)

	newAgent := &agent{
		sourceId:       req.Source,
		observedToken:  req.Params.Token,
		id:             req.Params.AgentId,
		name:           req.Params.AgentName,
		processVersion: req.Params.ProcessVersion,
		bundleVersion:  req.Params.BundleVersion,
		features:       req.Params.Features,
		errors:         errors,
		responder:      responder,
	}

	at.greeter <- newAgent

	return errors
}

func (at *AgentTracker) ProcessCheckMetricsPost(req protocol.MetricsPostRequest) {
	at.metricReqs <- &req
}

func (at *AgentTracker) start(ctx context.Context) {
	log.Infoln("Starting agent tracker")
	defer log.Infoln("Stopping agent tracker")

	for {
		select {
		case greetedAgent := <-at.greeter:
			at.handleGreetedAgent(greetedAgent)

		case reg := <-at.registrations:
			at.handlePollerRegistration(reg)

		case req := <-at.metricReqs:
			at.handleMetricsPostReq(req)

		case <-ctx.Done():
			return
		}
	}
}

func (at *AgentTracker) startMetricsDecomposer(ctx context.Context) {
	log.Infoln("Starting metrics decomposer")
	defer log.Infoln("Stopping metrics decomposer")

	for {
		select {
		case metric := <-at.metrics:
			log.WithFields(log.Fields{
				"agentId": metric.agent.id,
				"params":  metric.params,
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
							Name:       metricName,
							Value:      entry.Value,
							MetricType: entry.Type,
						}

						at.metricsRouter.Route(m)
					}

				}

			}

		case <-ctx.Done():
			return
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
		agent:  a,
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
	zone := a.zones[0]
	pathToChecks := path.Join(at.cfg.AgentsConfigDir, zone, a.id, "checks")

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
	for _, fileInfo := range contents {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), CheckFileSuffix) {
			filename := fileInfo.Name()
			jsonContent, err := ioutil.ReadFile(path.Join(pathToChecks, filename))

			check, err := at.decodeAdjustAndValidate(jsonContent, zone, filename, a)

			if err != nil {
				log.WithFields(log.Fields{
					"checksDir": checksDir,
					"file":      fileInfo,
				}).Warn("Unable to read checks file")
				continue
			}

			log.WithField("check", check).Warn("TODO: implement poller.prepare")
			// TODO ...implement version of poller.prepare exchange
		}
	}
}

func (at *AgentTracker) decodeAdjustAndValidate(jsonContent []byte, zone string, filename string, a *agent) (map[string]interface{}, error) {
	var check map[string]interface{}
	err := json.Unmarshal(jsonContent, &check)
	if err != nil {
		return nil, err
	}

	// Only do a cursory validation so that we don't have to maintain thorough validation rules and allow for
	// development time flexibility.
	if !utils.ContainsAllKeys(check, "type", "entity_id", "zone_id") {
		return nil, AgentTrackingError{Message: "Missing an expected field", SourceId: a.sourceId}
	}

	check[CheckId] = fmt.Sprintf("ch-%s-%s-%s", zone, a.id,
		strings.Map(
			utils.IdentifierSafe, strings.TrimSuffix(
				filename, CheckFileSuffix)))

	return check, nil
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
			Message:  "already observed",
		}
		return
	}

	at.agents[sourceId] = greetedAgent
}

// String renders the meaningful identifiers of this agent instance
func (a agent) String() string {
	return fmt.Sprintf("[id=%v, name=%v, processVersion=%v, bundleVersion=%v, features=%v, zones=%v]",
		a.id, a.name, a.processVersion, a.bundleVersion, a.features, a.zones)
}

func (a *agent) sendTo(method string, params interface{}) {
	msgId := atomic.AddUint64(&a.outMsgId, 1)

	rawParams, err := json.Marshal(params)
	if err != nil {
		log.WithField("params", params).Warn("Unable to marshal params")
		return
	}

	frameOut := &protocol.FrameMsg{
		FrameMsgCommon: protocol.FrameMsgCommon{
			Version: protocol.ProtocolVersion,
			Id:      msgId,
			Source:  "endpoint",
			Target:  "endpoint",
			Method:  method,
		},
		RawParams: rawParams,
	}

	log.WithFields(log.Fields{
		"msgId":   frameOut.Id,
		"method":  frameOut.Method,
		"agentId": a.id,
	}).Debug("SENDing frame to agent")

	a.responder.WriteJSON(frameOut)
}

type AgentTrackingError struct {
	Message  string
	SourceId string
}

func (e AgentTrackingError) Error() string {
	return fmt.Sprintf("An error occurred while tracking an agent %s: %s", e.SourceId, e.Message)
}
