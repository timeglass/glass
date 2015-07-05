#! /bin/bash
GOOS=`go env GOOS`
GOARCH=`go env GOARCH`
XGOARCH=$GOARCH

EXT=""
if [ "$GOOS" == "windows" ]; then
	EXT=".exe"
fi

function run_build_daemon {
	echo "building Daemon..."	
	go build -o $GOPATH/bin/glass-daemon${EXT} -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" ./glass-daemon
}

function run_build_cli {
	echo "building CLI..."
	go build -o $GOPATH/bin/glass${EXT} -ldflags "-X main.Version `cat VERSION` -X main.Build `date -u +%Y%m%d%H%M%S`" .
}

function run_run_daemon {
	run_build_daemon
	glass-daemon$EXT -bind :10000
}  

function run_test {
	echo "running all tests..."
	go test ./...
}  

function run_build {	
	run_build_cli
	run_build_daemon
}  

function run_release_prepare_dirs {
	echo "creating release directories..."
	rm -fr bin/${GOOS}*
	mkdir -p bin/${GOOS}_${GOARCH}
	cp $GOPATH/bin/glass-daemon${EXT} bin/${GOOS}_${GOARCH}
	cp $GOPATH/bin/glass${EXT} bin/${GOOS}_${GOARCH}
}

function run_make_installer {
	if [ "$GOOS" == "windows" ]; then
		echo "Creating Windows installers..."
		pushd installers/msi
		./make.bash
		popd
		cp 'installers/msi/bin/Timeglass Setup (x64).msi' bin
	fi

	if [ "$GOOS" == "darwin" ]; then
		echo "Creating OSX installers..."
		pushd installers/pkg
		./make.bash
		popd
		cp 'installers/pkg/bin/Timeglass Setup (x64).pkg' bin
	fi


	if [ "$GOOS" == "linux" ]; then
		echo "No installer yet for linux platform"
	fi
}  

function run_release {
	run_test
	run_build
	run_release_prepare_dirs
	run_make_installer
}

#choose command
echo "Detected OS '$GOOS'"
echo "Detected Arch '$GOARCH'"
case $1 in
    "test") run_test ;;
    "build" ) run_build ;;
	"build-daemon" ) run_build_daemon ;;
	"run-daemon" ) run_run_daemon ;;
	"installer" ) run_make_installer ;;
	"release" ) run_release ;;

	#
	# following commands are not portable
	# and only work on osx with "github-release"
	# "zip" and "shasum" installed and in PATH

 	# 1. zip all binaries
 	"publish-1" )
		rm -fr bin/dist
		mkdir -p bin/dist
		
		#move the installers
		mv bin/Timeglass* bin/dist/
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
		rm bin/dist/*_SHA256SUMS
		cd bin/dist && shasum -a256 * > ./timeglass_`cat ../../VERSION`_SHA256SUMS
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
		echo "Uploading zip files..."
		for FOLDER in ./bin/*_* ; do \
			NAME=`basename ${FOLDER}`_`cat VERSION` ; \
			ARCHIVE=bin/dist/${NAME}.zip ; \
			echo "  $ARCHIVE" ; \
			github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat VERSION` \
		    --name ${NAME}.zip \
		    --file ${ARCHIVE} ; \
		    echo "done!"; \
		done
		echo "Uploading shasums..."
		github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat` \
		    --name timeglass_`cat VERSION`_SHA256SUMS \
		    --file bin/dist/timeglass_`cat VERSION`_SHA256SUMS
		echo "done!"
		echo "Uploading OSX installer..."
		github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat` \
		    --name 'Timeglass Setup (x64).pkg' \
		    --file 'bin/dist/Timeglass Setup (x64).pkg'
		echo "done!"
		echo "Uploading Windows installer..."
		github-release upload \
		    --user timeglass \
		    --repo glass \
		    --tag v`cat` \
		    --name 'Timeglass Setup (x64).msi' \
		    --file 'bin/dist/Timeglass Setup (x64).msi'
		echo "done!"
 		;;
	*) run_test ;;
esac