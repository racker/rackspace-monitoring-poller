.PHONY: generate-mocks

all:

generate-mocks:
	mockgen -source=poller/poller.go -package=poller -destination=poller/poller_mock_test.go
	mockgen -destination mock_golang/mock_conn.go -package mock_golang net Conn