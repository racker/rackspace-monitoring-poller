package poller_test

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
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
