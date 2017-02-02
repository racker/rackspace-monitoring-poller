//
// Copyright 2017 Rackspace
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
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	set "github.com/deckarep/golang-set"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	protocheck "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type pendingPrepare struct {
	version  uint64
	checkIds []interface{}
	// committing indicates if the poller.commit has been sent
	committing bool
	sent       time.Time
}

type agent struct {
	responder *utils.SmartConn
	sourceId  string
	errors    chan<- error
	ctx       context.Context
	cancel    context.CancelFunc

	// observedToken is the token that was provided by the poller during hello handshake. It may not necessarily be valid.
	observedToken string

	id               string
	name             string
	processVersion   string
	bundleVersion    string
	features         []map[string]string
	zones            []string
	prepareBlockSize int

	outMsgId uint64

	checksLock      sync.Mutex
	checksVersion   uint64
	committedChecks set.Set // of check IDs
	// pendingPrepares is keyed by prepare message ID
	pendingPrepares map[uint64]*pendingPrepare
}

func newAgent(parentCtx context.Context, frame protocol.Frame, params *protocol.HandshakeParameters, responder *utils.SmartConn,
	prepareBlockSize int) (*agent, <-chan error) {
	errors := make(chan error, AgentErrorChanSize)

	ctx, cancel := context.WithCancel(parentCtx)

	newAgent := &agent{
		sourceId:         frame.GetSource(),
		observedToken:    params.Token,
		id:               params.AgentId,
		name:             params.AgentName,
		processVersion:   params.ProcessVersion,
		bundleVersion:    params.BundleVersion,
		features:         params.Features,
		zones:            params.ZoneIds,
		prepareBlockSize: prepareBlockSize,
		errors:           errors,
		responder:        responder,
		ctx:              ctx,
		cancel:           cancel,

		pendingPrepares: make(map[uint64]*pendingPrepare, 0),
	}

	return newAgent, errors
}

func (a *agent) handleResponse(frame protocol.Frame) {
	log.WithFields(log.Fields{
		"agent": a,
		"frame": frame,
	}).Warn("Agent handling response frame")

	a.checksLock.Lock()
	pending, exists := a.pendingPrepares[frame.GetId()]
	committing := pending.committing
	a.checksLock.Unlock()

	if !exists {
		log.WithFields(log.Fields{
			"agent": a,
			"frame": frame,
		}).Warn("Response frame not expected by agaent")
		return
	}

	if committing {
		result := &protocol.PollerCommitResult{}

		err := json.Unmarshal(frame.GetRawResult(), result)
		if err != nil {
			log.WithFields(log.Fields{
				"agent": a,
				"frame": frame,
				"err":   err,
			}).Warn("Failed to unmarshal PollerCommitResult")
			return
		}

		a.handlePollerCommitResponse(frame, result, pending)
	} else {
		result := &protocol.PollerPrepareResult{}

		err := json.Unmarshal(frame.GetRawResult(), result)
		if err != nil {
			log.WithFields(log.Fields{
				"agent": a,
				"frame": frame,
				"err":   err,
			}).Warn("Failed to unmarshal PollerPrepareResult")
			return
		}

		a.handlePollerPrepareResponse(frame, result, pending)
	}

}

func (a *agent) applyInitialChecks(pathToChecks string, zone string) {
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

	startChecks := make([]check.Check, 0)

	// Look for all .json files in the checks directory...and assume they are valid check details
	for _, fileInfo := range contents {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), CheckFileSuffix) {
			filename := fileInfo.Name()
			jsonContent, err := ioutil.ReadFile(path.Join(pathToChecks, filename))

			ch, err := a.decodeCheck(jsonContent, zone, filename)

			if err != nil {
				log.WithFields(log.Fields{
					"checksDir": checksDir,
					"file":      fileInfo,
				}).Warn("Unable to read checks file")
				continue
			}

			startChecks = append(startChecks, ch)
		}
	}

	a.prepareChecks(startChecks)
}

func (a *agent) prepareChecks(checksToStart []check.Check) {

	manifest := make([]protocol.PollerPrepareManifest, 0, len(checksToStart))
	for _, ch := range checksToStart {
		manifest = append(manifest, protocol.PollerPrepareManifest{
			Action:    protocol.PrepareActionStart,
			CheckType: ch.GetCheckType(),
			ZoneId:    ch.GetZoneID(),
			EntityId:  ch.GetEntityID(),
			Id:        ch.GetID(),
		})
	}

	allChecks := make([]check.Check, 0)
	allChecks = append(allChecks, checksToStart...)
	allCheckIds := make([]interface{}, 0, len(allChecks))
	for _, ch := range allChecks {
		allCheckIds = append(allCheckIds, ch.GetID())
	}

	a.checksLock.Lock()
	a.checksVersion++
	ourPending := &pendingPrepare{
		version:  a.checksVersion,
		sent:     time.Now(),
		checkIds: allCheckIds,
	}
	msgId := a.allocateMsgId()
	a.pendingPrepares[msgId] = ourPending
	a.checksLock.Unlock()

	prepareParams := &protocol.PollerPrepareStartParams{
		Version:  int(a.checksVersion),
		Manifest: manifest,
	}
	a.sendTo(msgId, protocol.MethodPollerPrepare, prepareParams)

	var prepareBlockParams *protocol.PollerPrepareBlockParams

	for _, ch := range allChecks {

		if prepareBlockParams == nil {
			prepareBlockParams = &protocol.PollerPrepareBlockParams{
				Version: int(a.checksVersion),
				Block:   make([]*protocheck.CheckIn, 0, a.prepareBlockSize),
			}
		}

		prepareBlockParams.Block = append(prepareBlockParams.Block, ch.GetCheckIn())

		if len(prepareBlockParams.Block) >= a.prepareBlockSize {
			a.sendTo(a.allocateMsgId(), protocol.MethodPollerPrepareBlock, prepareBlockParams)
			prepareBlockParams = nil
		}
	}

	if prepareBlockParams != nil {
		a.sendTo(a.allocateMsgId(), protocol.MethodPollerPrepareBlock, prepareBlockParams)
	}

	prepareEndParams := &protocol.PollerPrepareEndParams{
		Version:   int(a.checksVersion),
		Directive: protocol.PrepareDirectivePrepare,
	}
	a.sendTo(a.allocateMsgId(), protocol.MethodPollerPrepareEnd, prepareEndParams)

	log.WithFields(log.Fields{
		"agent":   a,
		"pending": ourPending,
	}).Debug("Prepared checks with agent")

	// later the response comes via handlePollerPrepareResponse...and then we can commit

}

func (a *agent) handlePollerPrepareResponse(frame protocol.Frame, result *protocol.PollerPrepareResult, ourPending *pendingPrepare) {
	switch result.Status {
	case protocol.PrepareResultStatusPrepared:

		a.checksLock.Lock()

		prepareCommitParams := &protocol.PollerCommitParams{
			Version: int(ourPending.version),
		}
		ourPending.committing = true
		msgId := a.allocateMsgId()
		delete(a.pendingPrepares, frame.GetId())
		a.pendingPrepares[msgId] = ourPending

		a.checksLock.Unlock()

		log.WithFields(log.Fields{
			"agent":   a,
			"pending": ourPending,
		}).Debug("Committing checks to agent")
		a.sendTo(msgId, protocol.MethodPollerCommit, prepareCommitParams)

	default:

		a.checksLock.Lock()
		log.WithFields(log.Fields{
			"preparing": ourPending, // nil if unknown
			"resp":      result,
		}).Warn("Failed checks preparation")
		delete(a.pendingPrepares, frame.GetId())
		a.checksLock.Unlock()

	}
}

func (a *agent) handlePollerCommitResponse(frame protocol.Frame, result *protocol.PollerCommitResult, ourPending *pendingPrepare) {

	switch result.Status {
	case protocol.PrepareResultStatusCommitted:
		a.checksLock.Lock()

		a.committedChecks = set.NewSetFromSlice(ourPending.checkIds)
		delete(a.pendingPrepares, frame.GetId())

		a.checksLock.Unlock()

		log.WithFields(log.Fields{
			"agent":   a,
			"pending": ourPending,
		}).Debug("Committed checks")
	}
}

func (a *agent) decodeCheck(jsonContent []byte, zone string, filename string) (check.Check, error) {
	ch, err := check.NewCheck(a.ctx, json.RawMessage(jsonContent))

	return ch, err
}

// String renders the meaningful identifiers of this agent instance
func (a agent) String() string {
	return fmt.Sprintf("[id=%v, name=%v, processVersion=%v, bundleVersion=%v, features=%v, zones=%v]",
		a.id, a.name, a.processVersion, a.bundleVersion, a.features, a.zones)
}

func (a *agent) allocateMsgId() uint64 {
	return atomic.AddUint64(&a.outMsgId, 1)
}

func (a *agent) sendTo(msgId uint64, method string, params interface{}) {

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

func (p *pendingPrepare) String() string {
	return fmt.Sprintf("[version=%v, sent=%v, committing=%v]", p.version, p.sent, p.committing)
}
