package utils_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"errors"

	"os"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
)

func TestSetupCommand(t *testing.T) {
	commandList := []string{"test", "command"}
	expected := &utils.Result{
		Command: &utils.PollerCmd{
			exec.Command(utils.PollerCommand, commandList...),
		},
		StdOut:         &bytes.Buffer{},
		StdErr:         &bytes.Buffer{},
		ErrorOnTimeout: false,
		Timeout:        time.Duration(1 * time.Nanosecond),
	}
	result := utils.SetupCommand(
		commandList,
		false, time.Duration(1*time.Nanosecond))

	assert.Equal(t, expected.Command.GetArgs(), result.Command.GetArgs())
	assert.Equal(t, expected.ErrorOnTimeout, result.ErrorOnTimeout)
	assert.Equal(t, expected.Timeout, result.Timeout)
}

func TestStartCommand(t *testing.T) {

	tests := []struct {
		name      string
		commander func(mc *utils.MockCommander) utils.Commander
		expected  error
	}{
		{
			name: "Happy path",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().Start()
				return mc
			},
			expected: nil,
		},
		{
			name: "Start errored",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().Start().Return(errors.New("Oops, I did it again."))
				return mc
			},
			expected: errors.New("Oops, I did it again."),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockCommander := utils.NewMockCommander(mockCtrl)
			result := &utils.Result{
				Command: tt.commander(mockCommander),
			}
			utils.StartCommand(result)
			assert.Equal(t, tt.expected, result.Error)
		})
	}
}

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name              string
		commander         func(mc *utils.MockCommander) utils.Commander
		timeout           time.Duration
		errOnTimeout      bool
		expectedErr       bool
		expectedErrString string
	}{
		{
			name: "Wait returns non exiterror",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().Wait().Return(errors.New("some error"))
				return mc
			},
			timeout:           time.Duration(1 * time.Second),
			expectedErr:       true,
			expectedErrString: "some error",
		},
		{
			name: "Wait returns exiterror",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().GetProcessState().Return(&os.ProcessState{})
				mc.EXPECT().Wait().Return(&exec.ExitError{
					ProcessState: &os.ProcessState{},
				})
				return mc
			},
			timeout:     time.Duration(1 * time.Second),
			expectedErr: false,
		},
		{
			name: "Run times out - expected",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().GetProcessState()
				mc.EXPECT().GetProcess().Return(&os.Process{Pid: 99999})
				mc.EXPECT().GetProcess().Return(&os.Process{Pid: 99999})
				mc.EXPECT().Wait().Do(func() { time.Sleep(1 * time.Second) })
				return mc
			},
			timeout:           time.Duration(1 * time.Millisecond),
			expectedErr:       true,
			expectedErrString: "os: process already finished",
		},
		{
			name: "Run times out - unexpected",
			commander: func(mc *utils.MockCommander) utils.Commander {
				mc.EXPECT().GetProcessState()
				mc.EXPECT().GetProcess().Return(&os.Process{Pid: 99999})
				mc.EXPECT().GetProcess().Return(&os.Process{Pid: 99999})
				mc.EXPECT().Wait().Do(func() { time.Sleep(1 * time.Second) })
				return mc
			},
			timeout:           time.Duration(1 * time.Millisecond),
			errOnTimeout:      true,
			expectedErr:       true,
			expectedErrString: "Failed on timeout!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockCommander := utils.NewMockCommander(mockCtrl)
			result := &utils.Result{
				Command:        tt.commander(mockCommander),
				Timeout:        tt.timeout,
				ErrorOnTimeout: tt.errOnTimeout,
			}
			utils.RunCommand(result)

			if tt.expectedErr {
				assert.Equal(t, tt.expectedErrString, result.Error.Error())
			} else {
				assert.NoError(t, result.Error)
			}

		})
	}
}

func TestBufferToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		buf      func() *bytes.Buffer
		expected []*utils.OutputMessage
	}{
		{
			name: "Happy path",
			buf: func() *bytes.Buffer {
				buffer := &bytes.Buffer{}
				buffer.Write([]byte(`{"level": "info", "msg":"test1"}
			{"level": "error", "msg":"test2", "address": "123"}
				{"level": "debug", "msg":"test3", "boundaddr": {"ip": "1.2.3.4", "port": 4567, "zone": "pzA"}}`))
				return buffer
			},
			expected: []*utils.OutputMessage{
				&utils.OutputMessage{
					Level: "info",
					Msg:   "test1",
				},
				&utils.OutputMessage{
					Level:   "error",
					Msg:     "test2",
					Address: "123",
				},
				&utils.OutputMessage{
					Level: "debug",
					Msg:   "test3",
					BoundAddr: &utils.BoundAddress{
						IP:   "1.2.3.4",
						Port: 4567,
						Zone: "pzA",
					},
				},
			},
		},
		{
			name: "Empty buffer",
			buf: func() *bytes.Buffer {
				buffer := &bytes.Buffer{}
				buffer.Write([]byte(``))
				return buffer
			},
			expected: []*utils.OutputMessage{},
		},
		{
			name: "Invalid json",
			buf: func() *bytes.Buffer {
				buffer := &bytes.Buffer{}
				buffer.Write([]byte(`{"random":}`))
				return buffer
			},
			expected: []*utils.OutputMessage{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.BufferToStringSlice(tt.buf())
			assert.Equal(t, tt.expected, got)

		})
	}
}

func TestGetProcess(t *testing.T) {
	pollerCommand := &utils.PollerCmd{
		exec.Command("echo", "test"),
	}

	err := pollerCommand.Start()

	defer pollerCommand.GetProcess().Kill()

	if err != nil {
		assert.Fail(t, fmt.Sprintf("Unable to start process due to %v", err))
	}
	assert.NotNil(t, pollerCommand.GetProcess())
}

func TestGetProcessState(t *testing.T) {
	pollerCommand := &utils.PollerCmd{
		exec.Command("echo", "test"),
	}

	err := pollerCommand.Start()

	defer pollerCommand.GetProcess().Kill()

	if err != nil {
		assert.Fail(t, fmt.Sprintf("Unable to start process due to %v", err))
	}
	assert.Nil(t, pollerCommand.GetProcessState())

	err = pollerCommand.Wait()

	if err != nil {
		assert.Fail(t, fmt.Sprintf("Unable to run process due to %v", err))
	}

	assert.NotNil(t, pollerCommand.GetProcessState())
}

func TestGetArgs(t *testing.T) {
	pollerCommand := &utils.PollerCmd{
		exec.Command(utils.PollerCommand, "echo", "test"),
	}

	assert.Equal(t, []string{utils.PollerCommand, "echo", "test"}, pollerCommand.GetArgs())
}

func TestTimebox(t *testing.T) {
	tests := []struct {
		name     string
		d        time.Duration
		boxed    func(t *testing.T)
		expected bool
	}{
		{
			name: "Expire timeout",
			d:    time.Duration(1 * time.Millisecond),
			boxed: func(t *testing.T) {
				time.Sleep(1 * time.Second)
			},
			expected: false,
		},
		{
			name: "Happy path",
			d:    time.Duration(1 * time.Second),
			boxed: func(t *testing.T) {
				time.Sleep(1 * time.Millisecond)
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed := utils.Timebox(nil, tt.d, tt.boxed)
			if tt.expected {
				assert.True(t, completed)
			} else {
				assert.False(t, completed)
			}
		})
	}
}
