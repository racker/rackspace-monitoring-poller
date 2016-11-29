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
)

type agent struct {
	responder      json.Encoder
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

type AgentTracker struct {
	cfg           config.EndpointConfig
	// key is FrameMsgCommon.Source
	agents        map[string]*agent

	greeter       chan *agent
	registrations chan *registration
}

type AgentTrackingError struct {
	Message  string
	SourceId string
}

func (e *AgentTrackingError) Error() string {
	return fmt.Sprintf("An error occurred while tracking an agent %s: %s", e.SourceId, e.Message)
}

// Returns a channel where errors observed about this agent are conveyed
func (at *AgentTracker) ProcessHello(req protocol.HandshakeRequest) <-chan error {
	newAgent := &agent{
		sourceId: req.Source,
		observedToken: req.Params.Token,
		id: req.Params.AgentId,
		name: req.Params.AgentName,
		processVersion: req.Params.ProcessVersion,
		bundleVersion: req.Params.BundleVersion,
		features: req.Params.Features,
	}

	at.greeter = make(chan *agent)
	at.greeter <- newAgent

	return newAgent.errors
}

func (at *AgentTracker) ProcessPollerRegister(req protocol.PollerRegister) {
	reg := registration{
		sourceId: req.Source,
		zones: req.Params[protocol.PollerZones],
	}

	at.registrations <- reg
}

func (at *AgentTracker) Start(cfg config.EndpointConfig) {
	at.cfg = cfg;
	go at.start()
}

func (at *AgentTracker) start() {
	log.Infoln("Starting agent tracker")

	for {
		select {
		case greetedAgent := <-at.greeter:
			at.handleGreetedAgent(greetedAgent)

		case reg := <-at.registrations:
			at.handlePollerRegistration(reg)
		}
	}
}

func (at *AgentTracker) handlePollerRegistration(reg *registration) {
	a, ok := at.agents[reg.sourceId]
	if !ok {
		log.WithField("sourceId", reg.sourceId).Warn("Trying to register poller for unknown source")
		return
	}

	a.zones = reg.zones

	go at.lookupChecks(a)
}

func (at *AgentTracker) lookupChecks(a *agent) {

}

func (at *AgentTracker) handleGreetedAgent(greetedAgent *agent) {
	log.WithField("agent", greetedAgent).Debug("RX greeted agent")

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
	a.outMsgId++
	frameOut := protocol.FrameMsg{
		FrameMsgCommon: protocol.FrameMsgCommon{
			Version: "1",
			Id: a.outMsgId,
			Method: method,
		},
		RawParams: json.Marshal(params),
	}

	a.responder.Encode(frameOut)
}