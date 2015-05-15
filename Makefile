run-deamon:
	go run -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
		./daemon/main.go \
		./daemon/server.go \
		./daemon/timer.go \
		--bind :10000 --mbu 1s

run-cli:
	go run -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
		./main.go