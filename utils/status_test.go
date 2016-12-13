package utils_test

import (
	"testing"

	"fmt"
	"os"

	"github.com/racker/rackspace-monitoring-poller/utils"
)

func TestNewStatusLine(t *testing.T) {
	sl := utils.NewStatusLine()
	if sl == nil || sl.String() != "" {
		t.Fatal(fmt.Fprintf(os.Stdout, "new line is set to %s", sl))
	}
}

func TestClear(t *testing.T) {
	sl := utils.NewStatusLine()
	sl.Add("willsmith", "new hotness")
	if sl == nil || sl.String() != "willsmith=new hotness" {
		t.Fatal(fmt.Fprintf(os.Stdout, "new line is set to %s", sl))
	}
	sl.Clear()
	if sl == nil || sl.String() != "" {
		t.Fatal(fmt.Fprintf(os.Stdout, "after Clear() new line is set to %s", sl))
	}
}

func TestAddAndPrint(t *testing.T) {
	sl := utils.NewStatusLine()
	sl.Add("willsmith", "new hotness")
	sl.Add("tommylee", "old and busted")
	if sl == nil || sl.String() != "willsmith=new hotness,tommylee=old and busted" {
		t.Fatal(fmt.Fprintf(os.Stdout, "new line is set to %s", sl))
	}
}
