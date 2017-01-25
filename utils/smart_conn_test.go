package utils_test

import (
	"testing"

	"fmt"
	"os"

	"errors"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/mock_golang"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

func NewHeartbeat() *HeartbeatTest {
	f := &HeartbeatTest{}
	f.Channel = "test"
	f.EntityId = "test2"
	return f
}

type HeartbeatTest struct {
	Channel  string
	EntityId string
}

func getMockedConnection(t *testing.T) *mock_golang.MockConn {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	return mock_golang.NewMockConn(mockCtrl)
}

func TestStart(t *testing.T) {
	mock_conn := getMockedConnection(t)
	err := utils.NewSmartConn(mock_conn).Start()
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}

func TestStartWithReadKeepalive(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().SetReadDeadline(gomock.Any()).Return(nil)
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.ReadKeepalive = 1
	err := smartConn.Start()
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}

func TestStartWithReadKeepaliveReadDeadlineFail(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().SetReadDeadline(gomock.Any()).Return(
		errors.New("Failure"))
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.ReadKeepalive = 1
	err := smartConn.Start()
	if err == nil || err.Error() != "Failure" {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back the wrong error: %s", err))
	}

}

func TestStartHeartbeatSendSet(t *testing.T) {
	mock_conn := getMockedConnection(t)
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.HeartbeatSendInterval = 1
	smartConn.HeartbeatSender = func(*utils.SmartConn) {}
	err := smartConn.Start()
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}

func TestClose(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Close()
	err := utils.NewSmartConn(mock_conn).Close()
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}

func TestCloseWithHeartbeatTimer(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Close()
	//set up heartbeatTimer by running the start function
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.HeartbeatSendInterval = 1
	smartConn.HeartbeatSender = func(*utils.SmartConn) {}
	err := smartConn.Start()
	// test
	err = smartConn.Close()
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}

func TestWrite(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Write([]byte("test"))
	offset, err := utils.NewSmartConn(mock_conn).Write([]byte("test"))
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
	if offset != 0 {
		t.Fatal(fmt.Fprintf(os.Stdout, "We returned offset: %s", fmt.Sprintf("%v", offset)))
	}
}

func TestWriteAllowanceSet(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Write([]byte("test"))
	// once for regular and once for defer call
	mock_conn.EXPECT().SetWriteDeadline(gomock.Any())
	mock_conn.EXPECT().SetWriteDeadline(gomock.Any())
	//set up encoder by running the start function
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.WriteAllowance = 1
	smartConn.Start()

	offset, err := smartConn.Write([]byte("test"))
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
	if offset != 0 {
		t.Fatal(fmt.Fprintf(os.Stdout, "We returned offset: %s", fmt.Sprintf("%v", offset)))
	}
}

func TestRead(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Read([]byte("test"))
	offset, err := utils.NewSmartConn(mock_conn).Read([]byte("test"))
	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
	if offset != 0 {
		t.Fatal(fmt.Fprintf(os.Stdout, "We returned offset: %s", fmt.Sprintf("%v", offset)))
	}
}

func TestRemoteAddr(t *testing.T) {
	mock_conn := getMockedConnection(t)
	//just validating that RemoteAddr was called on net.conn
	mock_conn.EXPECT().RemoteAddr()
	utils.NewSmartConn(mock_conn).RemoteAddr()
}

func TestWriteJSON(t *testing.T) {
	mock_conn := getMockedConnection(t)
	mock_conn.EXPECT().Write(gomock.Any())
	//set up encoder by running the start function
	smartConn := utils.NewSmartConn(mock_conn)
	smartConn.Start()
	// test
	err := smartConn.WriteJSON(NewHeartbeat())

	if err != nil {
		t.Fatal(fmt.Fprintf(os.Stdout, "We threw back an error: %s", err))
	}
}
