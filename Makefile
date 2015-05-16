build-hook: build-daemon build-cli
	glass init

build: build-daemon build-cli

test:
	go test ./...

build-daemon:
	go build -o $(GOPATH)/bin/glass-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./daemon 

build-cli:
	go build -o $(GOPATH)/bin/glass -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .

run-daemon: build-daemon
	glass-daemon --bind :10000 --mbu 1s

run-cli: build-cli
	glass

run-cli-start: build-cli
	glass start

run-cli-status: build-cli
	glass status

