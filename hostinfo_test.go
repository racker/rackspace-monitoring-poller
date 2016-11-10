package main

import (
	"testing"
	"bytes"
)

func TestHostInfoMemory_PopulateResult(t *testing.T) {
	hinfo := &HostInfoBase{Type:"MEMORY"}
	hostInfoMemory := NewHostInfoMemory(hinfo)

	cr := NewCheckResult()
	cr.AddMetrics(
		NewMetric("UsedPercentage", "", MetricFloat, 0.75, ""),
		NewMetric("Free", "", MetricNumber, 250, ""),
		NewMetric("Total", "", MetricNumber, 1000, ""),
		NewMetric("Used", "", MetricNumber, 750, ""),
		NewMetric("SwapFree", "", MetricNumber, 50, ""),
		NewMetric("SwapTotal", "", MetricNumber, 200, ""),
		NewMetric("SwapUsed", "", MetricNumber, 150, ""),
		NewMetric("SwapUsedPercentage", "", MetricFloat, 0.75, ""),
	)

	sourceFrame := &FrameMsg{

	}

	NowTimestampMillis = func() int64 { return 100 }

	response := &HostInfoResponse{}
	response.Result = hostInfoMemory.BuildResult(cr)
	response.SetResponseFrameMsg(sourceFrame)

	encoded,err := response.Encode()
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(encoded, []byte("{\"v\":\"\",\"id\":0,\"target\":\"\",\"source\":\"\",\"result\":{\"metrics\":{\"used_percentage\":0.75,\"free\":0,\"total\":0,\"used\":0,\"swap_free\":0,\"swap_total\":0,\"swap_used\":0,\"swap_percentage\":0.75},\"timestamp\":100}}")) {
		t.Error("wrong encoding")
	}
}