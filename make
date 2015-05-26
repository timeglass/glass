#! /bin/sh
GOOS=`go env GOOS`
GOARCH=`go env GOARCH`

function run_build_daemon {
	go build -o $GOPATH/bin/glass-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./daemon 
}

function run_build_cli {
	go build -o $GOPATH/bin/glass -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .
}

function run_test {
	echo "running all tests..."
	go test ./...
}  

function run_build {
	echo "building CLI..."
	run_build_cli
	echo "building Daemon..."
	run_build_daemon
}  

function run_release_prepare_dirs {
	echo "creating release directories..."
	rm -fr bin/*
	mkdir -p bin/${GOOS}_${GOARCH}
	cp $GOPATH/bin/glass-daemon bin/${GOOS}_${GOARCH}
	cp $GOPATH/bin/glass bin/${GOOS}_${GOARCH}
}

function run_release {
	run_test
	run_build
	run_release_prepare_dirs
}  

#choose command
echo "Detected OS '$GOOS'"
echo "Detected Arch '$GOARCH'"
case $1 in
    "test") run_test ;;
    "build" ) run_build ;;
	"release" ) run_release ;;
	*) run_test ;;
esac