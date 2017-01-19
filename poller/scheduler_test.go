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
	"github.com/racker/rackspace-monitoring-poller/utils"
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
	completed := utils.Timebox(t, 100*time.Millisecond, func(t *testing.T) {
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
	schedule.SendMetrics(&check.ResultSet{})
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
			ch: check.NewCheck(cancelCtx, json.RawMessage(`{
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
			}`), cancelFunc),
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
				assert.Contains(t, checkList, tt.ch.GetID())
			}
		})
	}
}

func TestEleScheduler_RunFrameConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStream := poller.NewMockConnectionStream(mockCtrl)
	mockStream.EXPECT().SendMetrics(gomock.Any()).AnyTimes()
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

	// NOTE: period is set to 90 seconds so that the check
	// only runs once in the jitter period (jitter is set to 100 ms)
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
	  "period":120,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":true
	  }`),
	}
	go schedule.RunFrameConsumer()

	// wait for jitter amount of time (and add 100 milliseconds to catch the other close)
	time.Sleep(200 * time.Millisecond)

	// close session
	schedule.Close()

	ctx, _ := schedule.GetContext()
	completed := utils.Timebox(t, 1000*time.Millisecond, func(t *testing.T) {
		<-ctx.Done()
	})
	assert.True(t, completed, "cancellation channel never notified")
}

// This is a variant of TestEleScheduler_RunFrameConsumer that eliminates all use of go routines, but narrows the
// scope of testing.
func TestEleScheduler_FrameConsumer_Scheduling(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCheckScheduler := poller.NewMockCheckScheduler(mockCtrl)
	mockCheckExecutor := poller.NewMockCheckExecutor(mockCtrl)

	mockStream := poller.NewMockConnectionStream(mockCtrl)
	scheduler := poller.NewCustomScheduler("pzAwesome", mockStream, mockCheckScheduler, mockCheckExecutor)

	scheduler.GetInput() <- &protocol.FrameMsg{
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
	  "period":120,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":true
	  }`),
	}

	mockCheckScheduler.EXPECT().Schedule(check.ExpectedCheckType("remote.tcp")).Do(func(ch check.Check) {
		// then close the frame consumer
		scheduler.Close()
	})

	scheduler.RunFrameConsumer()
}
