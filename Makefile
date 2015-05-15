run-deamon:
	go build -o $(GOPATH)/bin/sourceclock-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./daemon/... 
	sourceclock --bind :10000 --mbu 1s

run-cli-start:
	go run -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
		./main.go start

run-cli-pause:
	go run -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
		./main.go pause

run-cli-stop:
	go run -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
		./main.go stop