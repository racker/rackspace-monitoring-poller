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
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	set "github.com/deckarep/golang-set"
	"github.com/fsnotify/fsnotify"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	protocheck "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
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

type prepareBundle struct {
	checksToStart, checksToRestart, unclassified []check.Check
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
	features         []map[string]config.Feature
	zones            []string
	prepareBlockSize int

	preparations chan prepareBundle
	responses    chan protocol.Frame
	// The following are only accessed within runPrepareResponseCoordinator go routine

	checkId         uint64
	outMsgId        uint64
	checksVersion   uint64
	committedChecks set.Set // of check IDs
	// pendingPrepares is keyed by prepare message ID
	pendingPrepares map[uint64]*pendingPrepare
}

func newAgent(parentCtx context.Context, frame protocol.Frame, params *protocol.HandshakeParameters, responder *utils.SmartConn,
	prepareBlockSize int) (*agent, <-chan error) {
	errCh := make(chan error, AgentErrorChanSize)

	if len(params.ZoneIds) == 0 {
		errCh <- errors.New("Handshake is missing one or more zone IDs")
		return nil, errCh
	}

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
		errors:           errCh,
		responder:        responder,
		ctx:              ctx,
		cancel:           cancel,

		pendingPrepares: make(map[uint64]*pendingPrepare, 0),
		preparations:    make(chan prepareBundle),
		responses:       make(chan protocol.Frame),
	}

	go newAgent.runPrepareResponseCoordinator()

	return newAgent, errCh
}

func (a *agent) GetSourceId() string {
	return a.sourceId
}

func (a *agent) Close() {
	a.cancel()
}

func (a *agent) HandleResponse(frame protocol.Frame) {
	a.responses <- frame
}

func (a *agent) handleResponse(frame protocol.Frame) {
	log.WithFields(log.Fields{
		"agent": a,
		"frame": frame,
	}).Debug("Agent handling response frame")

	pending, exists := a.pendingPrepares[frame.GetId()]

	if !exists {
		log.WithFields(log.Fields{
			"agent": a,
			"frame": frame,
		}).Warn("Response frame not expected by agaent")
		return
	}

	if pending.committing {
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

func (a *agent) loadChecks(pathToChecks string, zone string) ([]check.Check, error) {
	checksDir, err := os.Open(pathToChecks)
	if err != nil {
		log.WithField("pathToChecks", pathToChecks).Warn("Unable to access agent checks directory")
		return nil, err
	}
	defer checksDir.Close()

	contents, err := checksDir.Readdir(0)
	if err != nil {
		log.WithField("pathToChecks", pathToChecks).Warn("Unable to read agent checks directory")
		return nil, err
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
					"checksDir": pathToChecks,
					"file":      fileInfo.Name(),
				}).Warn("Unable to read checks file")
				continue
			}

			a.checkId++
			ch.SetID(fmt.Sprintf("ch%08X", a.checkId))

			startChecks = append(startChecks, ch)
		}
	}

	return startChecks, nil
}

func (a *agent) applyInitialChecks(pathToChecks string, zone string) {

	startChecks, err := a.loadChecks(pathToChecks, zone)

	if err != nil {
		log.WithField("err", err).Warn("Unable to load checks during initialization")
		return
	}

	a.preparations <- prepareBundle{checksToStart: startChecks}

	go a.runFileWatcher(pathToChecks, zone)
}

func (a *agent) runFileWatcher(pathToChecks string, zone string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithField("err", err).Warn("Unable to create filesystem watcher")
		return
	}
	defer watcher.Close()

	log.WithField("path", pathToChecks).Debug("Starting file watching")
	defer log.WithField("path", pathToChecks).Debug("Stopping file watching")
	watcher.Add(pathToChecks)

	for {
		select {
		case <-a.ctx.Done():
			return

		case evt := <-watcher.Events:
			// NOTE chmod op seems to be fired when touch'ing a file
			if (fsnotify.Create|fsnotify.Remove|fsnotify.Write|fsnotify.Chmod)&evt.Op != 0 {
				// KEEP IT SIMPLE for now and just re-read all the files. Eventually we can track files to
				// IDs and perform more fine grained prepare-commits.

				log.WithField("event", evt).Debug("Triggering prepare-commit due to filesystem event")

				checks, err := a.loadChecks(pathToChecks, zone)
				if err != nil {
					log.WithField("err", err).Warn("Unable to load checks during file watching")
					continue
				}

				a.preparations <- prepareBundle{unclassified: checks}

			} else {
				log.WithField("event", evt).Debug("Ignoring filesystem event")
			}

		case err := <-watcher.Errors:
			log.WithFields(log.Fields{
				"err":  err,
				"path": pathToChecks,
			}).Warn("Error while watching for filesystem events")
			return
		}
	}
}

func (a *agent) runPrepareResponseCoordinator() {
	log.Debug("Starting prepare-response coodinator")
	defer log.Debug("Stopping prepare-response coodinator")

	for {
		select {
		case <-a.ctx.Done():
			return

		case bundle := <-a.preparations:
			if bundle.unclassified != nil {
				bundle.checksToStart = make([]check.Check, 0)
				bundle.checksToRestart = make([]check.Check, 0)

				for _, ch := range bundle.unclassified {
					if a.committedChecks.Contains(ch.GetID()) {
						bundle.checksToStart =
							append(bundle.checksToStart, ch)
					} else {
						bundle.checksToRestart =
							append(bundle.checksToRestart, ch)
					}
				}

				bundle.unclassified = nil
			}

			a.prepareChecks(bundle.checksToStart, bundle.checksToRestart)

		case respFrame := <-a.responses:
			a.handleResponse(respFrame)

		}
	}
}

func newManifest(ch check.Check, action string) *protocol.PollerPrepareManifest {
	return &protocol.PollerPrepareManifest{
		Action:    action,
		CheckType: ch.GetCheckType(),
		ZoneId:    ch.GetZoneID(),
		EntityId:  ch.GetEntityID(),
		Id:        ch.GetID(),
	}
}

func (a *agent) prepareChecks(checksToStart, checksToRestart []check.Check) {

	allChecks := make([]check.Check, 0)
	manifest := make([]protocol.PollerPrepareManifest, 0, len(checksToStart))

	allChecks = append(allChecks, checksToStart...)
	for _, ch := range checksToStart {
		manifest = append(manifest, *newManifest(ch, protocol.PrepareActionStart))
	}

	allChecks = append(allChecks, checksToRestart...)
	for _, ch := range checksToRestart {
		manifest = append(manifest, *newManifest(ch, protocol.PrepareActionRestart))
	}

	allCheckIds := make([]interface{}, 0, len(allChecks))
	for _, ch := range allChecks {
		allCheckIds = append(allCheckIds, ch.GetID())
	}

	a.checksVersion++
	ourPending := &pendingPrepare{
		version:  a.checksVersion,
		sent:     time.Now(),
		checkIds: allCheckIds,
	}
	msgId := a.allocateMsgId()
	a.pendingPrepares[msgId] = ourPending

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

		prepareCommitParams := &protocol.PollerCommitParams{
			Version: int(ourPending.version),
		}
		ourPending.committing = true
		msgId := a.allocateMsgId()
		delete(a.pendingPrepares, frame.GetId())
		a.pendingPrepares[msgId] = ourPending

		log.WithFields(log.Fields{
			"agent":   a,
			"pending": ourPending,
		}).Debug("Committing checks to agent")
		a.sendTo(msgId, protocol.MethodPollerCommit, prepareCommitParams)

	default:

		delete(a.pendingPrepares, frame.GetId())
		log.WithFields(log.Fields{
			"preparing": ourPending, // nil if unknown
			"resp":      result,
		}).Warn("Failed checks preparation")

	}
}

func (a *agent) handlePollerCommitResponse(frame protocol.Frame, result *protocol.PollerCommitResult, ourPending *pendingPrepare) {

	switch result.Status {
	case protocol.PrepareResultStatusCommitted:

		a.committedChecks = set.NewSetFromSlice(ourPending.checkIds)
		delete(a.pendingPrepares, frame.GetId())

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
	return fmt.Sprintf("version=%v, sent=%v, committing=%v", p.version, p.sent, p.committing)
}
