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

package utils

import (
	"errors"
	"time"
)

type NowTimestampMillisFunc func() int64

var NowTimestampMillis NowTimestampMillisFunc = func() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// InstallAlternateTimestampFunc is intended for unit testing where a deterministic timestamp needs to be
// temporarily enabled. Be sure to defer re-invoke this function to re-install the prior one.
func InstallAlternateTimestampFunc(newFunc NowTimestampMillisFunc) (priorFunc NowTimestampMillisFunc) {
	priorFunc = NowTimestampMillis
	NowTimestampMillis = newFunc
	return
}

// ScaleFractionalDuration is primarily useful when scaling durations that are "sub second", but more generally
// it's when duration is smaller than targetUnits. In that case, a fractional value is much more meangingful than
// a 0, which is what would happen with plain duration (i.e. integer) division. targetUnits should really be
// one of the Duration constants, such as time.Second.
func ScaleFractionalDuration(duration time.Duration, targetUnits time.Duration) float64 {
	return float64(duration) / float64(targetUnits)
}

// TimeLatencyTracking is used for tracking observations of poller<->endpoint server latency.
type TimeLatencyTracking struct {
	// PollerSendTimestamp is when the poller sent a heartbeat
	PollerSendTimestamp int64
	// PollerRecvTimestamp is when the poller received the heartbeat ack from the server
	PollerRecvTimestamp int64
	// ServerRecvTimestamp is when the server/endpoint recorded receiving our heartbeat
	ServerRecvTimestamp int64
	// ServerRespTimestamp is when the server/endpoint sent its heartbeat ack
	ServerRespTimestamp int64
}

// ComputeSkew uses the NTP algorithm documented here [1] to compute the poller-endpoint/server time offset and transit latency.
//
// [1]: https://www.eecis.udel.edu/~mills/ntp/html/warp.html
func (tt *TimeLatencyTracking) ComputeSkew() (offset int64, latency int64, _ error) {
	if tt.PollerSendTimestamp == 0 || tt.PollerRecvTimestamp == 0 || tt.ServerRecvTimestamp == 0 || tt.ServerRespTimestamp == 0 {
		return 0, 0, errors.New("Unable to compute with any unset timestamp")
	}

	// Variable aliases for the timeline
	T1 := tt.PollerSendTimestamp
	T2 := tt.ServerRecvTimestamp
	T3 := tt.ServerRespTimestamp
	T4 := tt.PollerRecvTimestamp

	offset = ((T2 - T1) + (T3 - T4)) / 2
	latency = ((T4 - T1) + (T3 - T2))

	return
}

// ChannelOfTimer is a nil-safe channel getter which becomes useful when needing to "select on a timer" that
// may not always be activated. For example,
//
//    select {
//    	case t := <- utils.ChannelOfTimer(timer) :
//    		... do something with timeout
//  	case f := <- frames:
//    		timer = time.NewTimer(d)
//    }
func ChannelOfTimer(timer *time.Timer) <-chan time.Time {
	if timer == nil {
		return nil
	} else {
		return timer.C
	}
}
