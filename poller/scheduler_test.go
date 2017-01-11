package poller_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/stretchr/testify/assert"
)

func TestNewScheduler(t *testing.T) {
	type args struct {
		zoneID string
		stream poller.ConnectionStream
	}
	tests := []struct {
		name   string
		zoneID string
		stream poller.ConnectionStream
	}{
		{
			name:   "Happy path",
			zoneID: "pzAwesome",
			stream: poller.NewConnectionStream(
				&config.Config{
					AgentId: "awesome agent",
					ZoneIds: []string{"pzAwesome", "pzGreat"},
				},
				x509.NewCertPool(),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.NewScheduler(tt.zoneID, tt.stream)
			//assert that zoneID is the same
			assert.Equal(t, tt.zoneID, got.GetZoneID())
		})
	}
}

func TestEleScheduler_Close(t *testing.T) {
	schedule := poller.NewScheduler("pzAwesome", poller.NewConnectionStream(
		&config.Config{
			AgentId: "awesome",
		},
		x509.NewCertPool(),
	))
	ctx, _ := schedule.GetContext()
	schedule.Close()
	completed := Timebox(t, 100*time.Millisecond, func(t *testing.T) {
		<-ctx.Done()
	})
	assert.True(t, completed, "cancellation channel never notified")
}

func TestEleScheduler_SendMetrics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStream := poller.NewMockConnectionStream(mockCtrl)
	schedule := poller.NewScheduler("pzAwesome", mockStream)
	mockStream.EXPECT().SendMetrics(gomock.Any()).Times(1)
	schedule.SendMetrics(&check.CheckResultSet{})
}

func TestEleScheduler_Register(t *testing.T) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	tests := []struct {
		name        string
		ch          check.Check
		expectedErr bool
	}{
		{
			name: "Happy path",
			ch: check.NewCheck(json.RawMessage(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":0,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":true
			}`), cancelCtx, cancelFunc),
			expectedErr: false,
		},
		{
			name:        "Check is nil",
			ch:          nil,
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := poller.NewScheduler("pzAwesome",
				poller.NewConnectionStream(
					&config.Config{
						AgentId: "test1",
					},
					x509.NewCertPool(),
				))
			if tt.expectedErr {
				assert.Error(t, s.Register(tt.ch))
			} else {
				assert.NoError(t, s.Register(tt.ch))
				checkList := []string{}
				for checkId, _ := range s.GetChecks() {
					checkList = append(checkList, checkId)
				}
				assert.Contains(t, checkList, tt.ch.GetId())
			}
		})
	}
}

func TestEleScheduler_RunFrameConsumer(t *testing.T) {
	t.Skipf("Skipped for now due to metrics being called twice.  Need to validate that's not a bug")
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStream := poller.NewMockConnectionStream(mockCtrl)
	mockStream.EXPECT().SendMetrics(gomock.Any())
	schedule := poller.NewScheduler("pzAwesome", mockStream)

	// set up wait period time measurement
	bakWaitPeriodTimeMeasurement := check.WaitPeriodTimeMeasurement
	bakCheckSpreadInMilliseconds := poller.CheckSpreadInMilliseconds
	defer func() {
		check.WaitPeriodTimeMeasurement = bakWaitPeriodTimeMeasurement
		poller.CheckSpreadInMilliseconds = bakCheckSpreadInMilliseconds
	}()

	check.WaitPeriodTimeMeasurement = time.Millisecond
	// set up jitter
	poller.CheckSpreadInMilliseconds = 100

	go schedule.RunFrameConsumer()
	schedule.GetInput() <- &protocol.FrameMsg{
		FrameMsgCommon: protocol.FrameMsgCommon{
			Id: 123,
		},
		RawParams: json.RawMessage(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":0,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":true
	  }`),
	}
	// wait for jitter amount of time (and add 100 milliseconds to catch the other close)
	time.Sleep(200 * time.Millisecond)

	// close session
	schedule.Close()

	ctx, _ := schedule.GetContext()
	completed := Timebox(t, 1000*time.Millisecond, func(t *testing.T) {
		<-ctx.Done()
	})
	assert.True(t, completed, "cancellation channel never notified")
}

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) bool {
	timer := time.NewTimer(d)
	completed := make(chan struct{})

	go func() {
		boxed(t)
		close(completed)
	}()

	select {
	case <-timer.C:
		if t != nil {
			t.Fatal("Timebox expired")
		}
		return false
	case <-completed:
		timer.Stop()
		return true
	}
}
