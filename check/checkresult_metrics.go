package check

import (
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

func NewMetricsPostRequest(crs *CheckResultSet) *protocol.MetricsPostRequest {
	req := &protocol.MetricsPostRequest{}
	req.Version = "1"
	req.Method = "check_metrics.post"
	req.Params.EntityId = crs.Check.GetEntityId()
	req.Params.CheckId = crs.Check.GetId()
	req.Params.CheckType = crs.Check.GetCheckType()
	req.Params.Metrics = []protocol.MetricWrap{ConvertToMetricResults(crs)}
	req.Params.MinCheckPeriod = crs.Check.GetPeriod() * 1000
	req.Params.State = crs.State
	req.Params.Status = crs.Status
	req.Params.Timestamp = utils.NowTimestampMillis()
	return req
}

func ConvertToMetricResults(crs *CheckResultSet) protocol.MetricWrap {
	wrappers := make(protocol.MetricWrap, 0)
	wrappers = append(wrappers, nil) // needed for the current protocol
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		mapper := make(map[string]*protocol.MetricTVU)
		for key, m := range cr.Metrics {
			mapper[key] = &protocol.MetricTVU{
				Type:  m.TypeString,
				Value: fmt.Sprintf("%v", m.Value),
				Unit:  m.Unit,
			}
		}
		wrappers = append(wrappers, mapper)
	}
	return wrappers
}
