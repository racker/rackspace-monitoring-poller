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
	log "github.com/sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/protocol"

	"context"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"net"
	"path"
)

const (
	AgentErrorChanSize    = 1
	GreeterChanSize       = 10
	RegistrationsChanSize = 10
	MetricsPostsChanSize  = 50
	MetricsStoreChanSize  = 50
	CheckId               = "id"
	CheckFileSuffix       = ".json"
)

type metricsToStore struct {
	agent  *agent
	params *protocol.MetricsPostContent
}

type metricsPost struct {
	params *protocol.MetricsPostContent
	frame  protocol.Frame
}

type AgentTracker struct {
	cfg *config.EndpointConfig
	// key is FrameMsgCommon.Source. All access to this map must be performed in the AgentTracker.start go routine
	agentsBySource map[string]*agent
	agentsByAddr   map[net.Addr]*agent

	ctx    context.Context
	cancel context.CancelFunc

	agentsToGreet chan *agent
	metricsPosts  chan *metricsPost
	metrics       chan *metricsToStore
	metricsRouter *MetricsRouter
}

func NewAgentTracker(cfg *config.EndpointConfig) *AgentTracker {

	ctx, cancel := context.WithCancel(context.Background())

	return &AgentTracker{
		cfg: cfg,

		agentsBySource: make(map[string]*agent),
		agentsByAddr:   make(map[net.Addr]*agent),

		ctx:    ctx,
		cancel: cancel,

		agentsToGreet: make(chan *agent, GreeterChanSize),
		metricsPosts:  make(chan *metricsPost, MetricsPostsChanSize),
		metrics:       make(chan *metricsToStore, MetricsStoreChanSize),
	}
}

func (at *AgentTracker) Start(ctx context.Context) {
	go at.runAgentInteractions(ctx)
	go at.runMetricsDecomposer(ctx)
}

func (at *AgentTracker) UseMetricsRouter(mr *MetricsRouter) {
	at.metricsRouter = mr
}

// Returns a channel where errors observed about this agent are conveyed
func (at *AgentTracker) NewAgentFromHello(frame protocol.Frame, params *protocol.HandshakeParameters,
	responder *utils.SmartConn) <-chan error {
	newAgent, errors := newAgent(at.ctx, frame, params, responder, at.cfg.PrepareBlockSize)
	if newAgent != nil {
		at.agentsToGreet <- newAgent
		at.agentsByAddr[responder.RemoteAddr()] = newAgent
	}

	return errors
}

func (at *AgentTracker) CloseAgentByAddr(addr net.Addr) {
	if agent, ok := at.agentsByAddr[addr]; ok {
		delete(at.agentsByAddr, addr)
		delete(at.agentsBySource, agent.GetSourceId())
		agent.Close()
	}
}

func (at *AgentTracker) handleCheckMetricsPost(post *metricsPost) {
	at.metricsPosts <- post
}

func (at *AgentTracker) runAgentInteractions(ctx context.Context) {
	log.Infoln("Starting agent tracker")
	defer log.Infoln("Stopping agent tracker")

	for {
		select {
		case greetedAgent := <-at.agentsToGreet:
			at.handleGreetedAgent(greetedAgent)

		case post := <-at.metricsPosts:
			at.handleMetricsPostReq(post)

		case <-ctx.Done():
			return
		}
	}
}

func (at *AgentTracker) runMetricsDecomposer(ctx context.Context) {
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

func (at *AgentTracker) handleMetricsPostReq(post *metricsPost) {
	a, ok := at.agentsBySource[post.frame.GetSource()]
	if !ok {
		log.WithField("sourceId", post.frame.GetSource()).Warn("Trying to handle metrics from unknown source")
		return
	}

	metric := &metricsToStore{
		agent:  a,
		params: post.params,
	}

	at.metrics <- metric
}

func (at *AgentTracker) handleResponse(frame protocol.Frame) {
	a, ok := at.agentsBySource[frame.GetSource()]
	if !ok {
		log.WithField("sourceId", frame.GetSource()).Warn("Got response unknown source")
		return
	}

	a.HandleResponse(frame)
}

func (at *AgentTracker) applyInitialChecks(a *agent) {
	log.WithField("agent", a).Debug("Looking up checks to apply to agent")

	//TODO iterate through zones in poller registration
	zone := a.zones[0]
	pathToChecks := path.Join(at.cfg.AgentsConfigDir, zone, a.id, "checks")

	a.applyInitialChecks(pathToChecks, zone)

}

func (at *AgentTracker) handleGreetedAgent(greetedAgent *agent) {
	log.WithField("agent", greetedAgent).Debug("Handling greeted agent")

	sourceId := greetedAgent.sourceId
	if _, exists := at.agentsBySource[sourceId]; exists {
		log.WithFields(log.Fields{
			"sourceId": sourceId,
		}).Warn("Agent greeting already observed")
		greetedAgent.errors <- AgentTrackingError{
			SourceId: sourceId,
			Message:  "already observed",
		}
		return
	}

	at.agentsBySource[sourceId] = greetedAgent

	go at.applyInitialChecks(greetedAgent)
}

type AgentTrackingError struct {
	Message  string
	SourceId string
}

func (e AgentTrackingError) Error() string {
	return fmt.Sprintf("An error occurred while tracking an agent %s: %s", e.SourceId, e.Message)
}
