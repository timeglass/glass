XC_ARCH = "amd64"
XC_OS = "darwin"
#XC_OS = "darwin linux windows"

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

run-cli-log: build-cli
	glass log


#1. build release binaries
release:
	rm -fr bin/*
	mkdir -p bin/
	@echo "Building..."
	gox \
	    -os=$(XC_OS) \
	    -arch=$(XC_ARCH) \
	    -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
	    -output "bin/{{.OS}}_{{.Arch}}/glass" \
	    .
	gox \
	    -os=$(XC_OS) \
	    -arch=$(XC_ARCH) \
	    -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" \
	    -output "bin/{{.OS}}_{{.Arch}}/glass-daemon" \
	    ./daemon 

#2. tag git commit
publish-tag:
	git tag v$(shell cat VERSION)
	git push --tags

#3 create github release (requires GITHUB_TOKEN environment variable)
publish-release:
	github-release release \
    --user timeglass \
    --repo glass \
    --tag v$(shell cat VERSION) \
    --pre-release

#4 zip binaries
publish-zip:
	rm -fr bin/dist
	mkdir -p bin/dist
	for FOLDER in ./bin/*_* ; do \
		NAME=`basename $$FOLDER`_`CAT VERSION` ; \
		ARCHIVE=bin/dist/$$NAME.zip ; \
		pushd $$FOLDER ; \
		echo Zipping: $$FOLDER... `pwd` ; \
		zip ../dist/$$NAME.zip ./* ; \
		popd ; \
	done 

#5 create checksums of zip archives
publish-checksum:
	cd bin/dist && shasum -a256 * > ./timeglass_$(shell cat VERSION)_SHA256SUMS

#6 upload zip and checksums
publish-upload: 
	for FOLDER in ./bin/*_* ; do \
		NAME=`basename $$FOLDER`_`CAT VERSION` ; \
		ARCHIVE=bin/dist/$$NAME.zip ; \
		github-release upload \
	    --user timeglass \
	    --repo glass \
	    --tag v$(shell cat VERSION) \
	    --name $$NAME.zip \
	    --file $$ARCHIVE ; \
	done	
	github-release upload \
	    --user timeglass \
	    --repo glass \
	    --tag v$(shell cat VERSION) \
	    --name timeglass_$(shell cat VERSION)_SHA256SUMS \
	    --file bin/dist/timeglass_$(shell cat VERSION)_SHA256SUMS