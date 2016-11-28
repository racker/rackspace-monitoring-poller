package hostinfo

type HostInfoMemoryMetrics struct {
	UsedPercentage     float64 `json:"used_percentage"`
	Free               uint64  `json:"free"`
	Total              uint64  `json:"total"`
	Used               uint64  `json:"used"`
	SwapFree           uint64  `json:"swap_free"`
	SwapTotal          uint64  `json:"swap_total"`
	SwapUsed           uint64  `json:"swap_used"`
	SwapUsedPercentage float64 `json:"swap_percentage"`
}

type HostInfoMemoryResult struct {
	Metrics   HostInfoMemoryMetrics `json:"metrics"`
	Timestamp int64                 `json:"timestamp"`
}

