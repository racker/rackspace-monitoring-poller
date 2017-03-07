package protocol

const (
	// MethodEmpty indicates this message is a response
	MethodEmpty = ""

	// MethodHandshakeHello is SENT TO ---> endpoint server
	MethodHandshakeHello   = "handshake.hello"
	MethodCheckScheduleGet = "check_schedule.get"
	MethodPollerRegister   = "poller.register"
	// MethodHeartbeatPost is SENT TO ---> endpoint server
	MethodHeartbeatPost = "heartbeat.post"
	// MethodCheckMetricsPostMulti is SENT TO ---> endpoint server
	MethodCheckMetricsPostMulti = "check_metrics.post_multi"
	// MethodHostInfoGet is RECV FROM <--- endpoint server
	MethodHostInfoGet = "host_info.get"
	// MethodPollerPrepare is RECV FROM <--- endpoint server
	MethodPollerPrepare = "poller.prepare"
	// MethodPollerPrepareBlock is RECV FROM <--- endpoint server
	MethodPollerPrepareBlock = "poller.prepare.block"
	// MethodPollerPrepareEnd is RECV FROM <--- endpoint server
	MethodPollerPrepareEnd = "poller.prepare.end"
	// MethodPollerCommit is RECV FROM <--- endpoint server
	MethodPollerCommit = "poller.commit"
)
