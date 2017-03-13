package poller_test

import (
	"testing"
	"time"

	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"sync"
)

func TestNewScheduler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		zoneID string
	}
	tests := []struct {
		name   string
		zoneID string
	}{
		{
			name:   "Happy path",
			zoneID: "pzAwesome",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.NewScheduler(tt.zoneID, poller.NewMockConnectionStream(ctrl))
			//assert that zoneID is the same
			assert.Equal(t, tt.zoneID, got.GetZoneID())
		})
	}
}

func TestEleScheduler_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	schedule := poller.NewScheduler("pzAwesome", poller.NewMockConnectionStream(ctrl))
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

type checkIdMatcher struct {
	id string
}

func (c checkIdMatcher) Matches(x interface{}) bool {
	ch, ok := x.(check.Check)
	if !ok {
		return false
	}

	return ch.GetID() == c.id
}

func (c checkIdMatcher) String() string {
	return fmt.Sprintf("has check ID %v", c.id)
}

func TestEleScheduler_ReconcileChecks_AllStart(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStream := poller.NewMockConnectionStream(mockCtrl)
	checkScheduler := poller.NewMockCheckScheduler(mockCtrl)
	checkExecutor := poller.NewMockCheckExecutor(mockCtrl)

	scheduler := poller.NewCustomScheduler("znA", mockStream, checkScheduler, checkExecutor)
	defer scheduler.Close()

	var wg sync.WaitGroup
	done := func(ch check.Check) { wg.Done() }
	checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch1"}).Do(done)
	wg.Add(1)
	checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch2"}).Do(done)
	wg.Add(1)

	cp := loadChecksPreparation(t,
		checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
		checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
	)

	err := scheduler.ValidateChecks(cp)
	assert.NoError(t, err)

	scheduler.ReconcileChecks(cp)

	utils.Timebox(t, 10*time.Second, func(t *testing.T) {
		wg.Wait()
	})

	scheduled := scheduler.GetScheduledChecks()
	assert.Len(t, scheduled, 2)

	// ...verifies expected function calls, above
}

func TestEleScheduler_ReconcileChecks(t *testing.T) {
	origCP := loadChecksPreparation(t,
		checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
		checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
	)

	tests := []struct {
		name              string
		prepMockScheduler func(checkScheduler *poller.MockCheckScheduler)
		cp                *poller.ChecksPreparation
		verify            func(t *testing.T, scheduled []check.Check, scheduledAfter []check.Check)
	}{
		{
			name: "continueAll",

			cp: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
			),

			verify: func(t *testing.T, scheduled []check.Check, scheduledAfter []check.Check) {
				scheduledIdsAfter := extractCheckIds(scheduledAfter)
				assert.Contains(t, scheduledIdsAfter, "ch1")
				assert.Contains(t, scheduledIdsAfter, "ch2")

				assert.Equal(t, findCheck(scheduled, "ch1"), findCheck(scheduledAfter, "ch1"))
				assert.Equal(t, findCheck(scheduled, "ch2"), findCheck(scheduledAfter, "ch2"))
			},
		},
		{
			name: "restartOne",

			prepMockScheduler: func(checkScheduler *poller.MockCheckScheduler) {
				checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch1"})
			},

			cp: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionRestart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
			),

			verify: func(t *testing.T, scheduled []check.Check, scheduledAfter []check.Check) {
				scheduledIdsAfter := extractCheckIds(scheduledAfter)
				assert.Contains(t, scheduledIdsAfter, "ch1")
				assert.Contains(t, scheduledIdsAfter, "ch2")
				assert.Equal(t, findCheck(scheduled, "ch2"), findCheck(scheduledAfter, "ch2"))

				ch1 := findCheck(scheduled, "ch1")
				assertCheckIsDone(t, ch1)
				ch1After := findCheck(scheduledAfter, "ch1")
				assert.NotEqual(t, ch1, ch1After)
			},
		},
		{
			name: "startAnother",

			prepMockScheduler: func(checkScheduler *poller.MockCheckScheduler) {
				checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch3"})
			},

			cp: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
				checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.ping", name: "ping_check", id: "ch3", entityId: "en1", zonedId: "znA"},
			),

			verify: func(t *testing.T, scheduled []check.Check, scheduledAfter []check.Check) {
				scheduledIdsAfter := extractCheckIds(scheduledAfter)
				assert.Contains(t, scheduledIdsAfter, "ch1")
				assert.Contains(t, scheduledIdsAfter, "ch2")
				assert.Contains(t, scheduledIdsAfter, "ch3")
				assert.Equal(t, findCheck(scheduled, "ch1"), findCheck(scheduledAfter, "ch1"))
				assert.Equal(t, findCheck(scheduled, "ch2"), findCheck(scheduledAfter, "ch2"))
			},
		},
		{
			name: "stopOne",

			cp: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
			),

			verify: func(t *testing.T, scheduled []check.Check, scheduledAfter []check.Check) {
				ch2 := findCheck(scheduled, "ch2")
				assertCheckIsDone(t, ch2)
				assert.Len(t, scheduledAfter, 1)
				assert.Contains(t, extractCheckIds(scheduledAfter), "ch1")
				assert.Equal(t, findCheck(scheduled, "ch1"), findCheck(scheduledAfter, "ch1"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockStream := poller.NewMockConnectionStream(mockCtrl)
			checkScheduler := poller.NewMockCheckScheduler(mockCtrl)
			checkExecutor := poller.NewMockCheckExecutor(mockCtrl)

			scheduler := poller.NewCustomScheduler("znA", mockStream, checkScheduler, checkExecutor)
			defer scheduler.Close()

			var wg sync.WaitGroup
			done := func(ch check.Check) { wg.Done() }
			checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch1"}).Do(done)
			wg.Add(1)
			checkScheduler.EXPECT().Schedule(checkIdMatcher{id: "ch2"}).Do(done)
			wg.Add(1)

			//origCP := loadChecksPreparation(t,
			//	checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.tcp", name:"tcp_check",id:"ch1", entityId:"en1", zonedId:"znA"},
			//	checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.ping", name:"ping_check",id:"ch2", entityId:"en1", zonedId:"znA"},
			//)
			//
			err := scheduler.ValidateChecks(origCP)
			assert.NoError(t, err)

			scheduler.ReconcileChecks(origCP)
			utils.Timebox(t, 10*time.Second, func(t *testing.T) {
				wg.Wait()
			})

			scheduled := scheduler.GetScheduledChecks()

			if tt.prepMockScheduler != nil {
				tt.prepMockScheduler(checkScheduler)
			}

			scheduler.ReconcileChecks(tt.cp)

			// allow time for channel ops
			time.Sleep(10 * time.Millisecond)

			scheduledAfter := scheduler.GetScheduledChecks()

			tt.verify(t, scheduled, scheduledAfter)
		})
	}
}

func assertCheckIsDone(t *testing.T, ch check.Check) {
	utils.Timebox(t, 10*time.Millisecond, func(t *testing.T) {
		<-ch.Done()
	})
}

func findCheck(checks []check.Check, id string) check.Check {
	for _, ch := range checks {
		if ch.GetID() == id {
			return ch
		}
	}

	return nil
}

func extractCheckIds(checks []check.Check) []string {
	ids := make([]string, 0, len(checks))

	for _, ch := range checks {
		ids = append(ids, ch.GetID())
	}

	return ids
}

func TestEleScheduler_ValidateChecks_Fails(t *testing.T) {

	tests := []struct {
		name           string
		preCP          *poller.ChecksPreparation
		validateCP     *poller.ChecksPreparation
		expectedErrStr string
		extra          func(t *testing.T)
	}{
		{
			name: "startExisting",
			preCP: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
				checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.ping", name: "ping_check", id: "ch2", entityId: "en1", zonedId: "znA"},
			),
			validateCP: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionStart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
			),
			expectedErrStr: "Reconciling was told to start a check, but it already existed: ch1",
		},
		{
			name: "restartMissing",
			validateCP: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionRestart, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
			),
			expectedErrStr: "Reconciling was told to restart a check, but it does not exist: ch1",
		},
		{
			name: "continueMissing",
			validateCP: loadChecksPreparation(t,
				checkLoadInfo{action: protocol.PrepareActionContinue, checkType: "remote.tcp", name: "tcp_check", id: "ch1", entityId: "en1", zonedId: "znA"},
			),
			expectedErrStr: "Reconciling was told to continue a check, but it does not exist: ch1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockStream := poller.NewMockConnectionStream(mockCtrl)
			checkScheduler := poller.NewMockCheckScheduler(mockCtrl)
			checkExecutor := poller.NewMockCheckExecutor(mockCtrl)

			scheduler := poller.NewCustomScheduler("znA", mockStream, checkScheduler, checkExecutor)
			defer scheduler.Close()

			if tt.preCP != nil {
				var wg sync.WaitGroup
				count := len(tt.preCP.GetActionableChecks())
				wg.Add(count)
				checkScheduler.EXPECT().Schedule(gomock.Any()).Times(count).Do(func(ch check.Check) { wg.Done() })

				scheduler.ReconcileChecks(tt.preCP)
				utils.Timebox(t, 10*time.Second, func(t *testing.T) {
					wg.Wait()
				})
			}

			err := scheduler.ValidateChecks(tt.validateCP)
			assert.EqualError(t, err, tt.expectedErrStr)

			if tt.extra != nil {
				tt.extra(t)
			}
		})
	}

}

func TestEleScheduler_Schedule_DisabledChecks(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStream := poller.NewMockConnectionStream(mockCtrl)
	checkExecutor := poller.NewMockCheckExecutor(mockCtrl)

	scheduler := poller.NewCustomScheduler("znA", mockStream,
		nil, // use default scheduler, since that's what we're testing
		checkExecutor)
	defer scheduler.Close()

	var wg sync.WaitGroup
	done := func(ch check.Check) { wg.Done() }
	checkExecutor.EXPECT().Execute(checkIdMatcher{id: "ch2"}).Do(done)
	wg.Add(1)

	cp := loadChecksPreparation(t,
		checkLoadInfo{action: protocol.PrepareActionStart, name: "tcp_check_disabled", checkType: "remote.tcp", id: "ch1", entityId: "en1", zonedId: "znA"},
		checkLoadInfo{action: protocol.PrepareActionStart, name: "tcp_check", checkType: "remote.tcp", id: "ch2", entityId: "en2", zonedId: "znA"},
	)

	origSpread := poller.CheckSpreadInMilliseconds
	defer func() { poller.CheckSpreadInMilliseconds = origSpread }()
	poller.CheckSpreadInMilliseconds = 10

	scheduler.ReconcileChecks(cp)

	utils.Timebox(t, 2*time.Second, func(t *testing.T) {
		wg.Wait()
	})

	// We'll get report of 2 "scheduled" checks, but via mock expectations above confirmed that only one executed

	scheduled := scheduler.GetScheduledChecks()
	assert.Len(t, scheduled, 2)

	// ...verifies expected function calls, above
}
