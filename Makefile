build-hook: build-daemon build-cli
	sourceclock hook

build: build-daemon build-cli

build-daemon:
	go build -o $(GOPATH)/bin/sourceclock-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./daemon 

build-cli:
	go build -o $(GOPATH)/bin/sourceclock -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .

run-daemon: build-daemon
	sourceclock-daemon --bind :10000 --mbu 1s

run-cli: build-cli
	sourceclock

run-cli-start: build-cli
	sourceclock start

run-cli-split: build-cli
	sourceclock split