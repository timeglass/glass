#! /bin/bash
GOOS=`go env GOOS`
GOARCH=`go env GOARCH`
XGOARCH=$GOARCH
if [ "$XGOARCH" == "amd64" ]
then
	XGOARCH="386"
else 
	XGOARCH="amd64"
fi

function run_build_daemon {
	go build -o $GOPATH/bin/glass-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./glass-daemon
}

function run_build_cli {
	go build -o $GOPATH/bin/glass -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .
}

function run_run_daemon {
	run_build_daemon
	glass-daemon -bind :10000
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

function run_xbuild {
	echo "Cross Compiling CLI for '$XGOARCH'..."
	mkdir -p bin/${GOOS}_${XGOARCH}
	CGO_ENABLED=1 GOARCH=$XGOARCH go build -o bin/${GOOS}_${XGOARCH}/glass -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .

	echo "Cross Compiling Daemon for '$XGOARCH'..."
	CGO_ENABLED=1 GOARCH=$XGOARCH go build -o bin/${GOOS}_${XGOARCH}/glass-daemon -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./glass-daemon
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
	run_xbuild
}  

#choose command
echo "Detected OS '$GOOS'"
echo "Detected Arch '$GOARCH'"
case $1 in
    "test") run_test ;;
    "build" ) run_build ;;
	"run-daemon" ) run_run_daemon ;;
	"xbuild" ) run_xbuild ;;
	"release" ) run_release ;;

	#
	# following commands are not portable
	# and only work on osx with "github-release"
	# "zip" and "shasum" installed and in PATH

 	# 1. zip all binaries
 	"publish-1" )
		rm -fr bin/dist
		mkdir -p bin/dist
		for FOLDER in ./bin/*_* ; do \
			NAME=`basename ${FOLDER}`_`cat VERSION` ; \
			ARCHIVE=bin/dist/${NAME}.zip ; \
			pushd ${FOLDER} ; \
			echo Zipping: ${FOLDER}... `pwd` ; \
			zip ../dist/${NAME}.zip ./* ; \
			popd ; \
		done 
		;;

	# 2. checksum zips
	"publish-2" )
		cd bin/dist && shasum -a256 * > ./timeglass_`cat VERSION`_SHA256SUMS
		;;

	# 3. create tag and push it
	"publish-3" )
		git tag v`cat VERSION`
		git push --tags
		;;

	# 4. draft a new release
	"publish-4" )
		github-release release \
	    	--user timeglass \
	    	--repo glass \
	    	--tag v`cat VERSION` \
	    	--pre-release
 		;;
 	# 5. upload files
	"publish-5" )
		for FOLDER in ./bin/*_* ; do \
			NAME=`basename ${FOLDER}`_`cat VERSION` ; \
			ARCHIVE=bin/dist/${NAME}.zip ; \
			github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat VERSION` \
		    --name ${NAME}.zip \
		    --file ${ARCHIVE} ; \
		done	
		github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat` \
		    --name timeglass_`cat VERSION`_SHA256SUMS \
		    --file bin/dist/timeglass_`cat VERSION`_SHA256SUMS
 		;;
	*) run_test ;;
esac