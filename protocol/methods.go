package protocol

const (
	MethodEmpty            = ""
	MethodHandshakeHello   = "handshake.hello"
	MethodCheckScheduleGet = "check_schedule.get"
	MethodPollerRegister   = "poller.register"
	MethodPollerChecksAdd  = "poller.checks.add"
	MethodPollerChecksEnd  = "poller.checks.end"
	MethodHeartbeatPost    = "heartbeat.post"
	MethodCheckMetricsPost = "check_metrics.post"
	MethodHostInfoGet      = "host_info.get"
)
