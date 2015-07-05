VERSION=`cat ../../VERSION`
echo "Copying version $VERSION binaries..."

mkdir -p ./glass-daemon
cp ../../bin/darwin_amd64/glass-daemon ./glass-daemon
mkdir -p ./glass
cp ../../bin/darwin_amd64/glass ./glass

echo "Building daemon component package..."
pkgbuild --root ./glass-daemon \
		 --identifier com.timeglass.glass-daemon \
		 --install-location /Applications/Timeglass \
		 --scripts ./scripts/glass-daemon \
		 --version $VERSION \
		 glass-daemon.pkg

echo "Building cli component package..."
pkgbuild --root ./glass \
		 --identifier com.timeglass.glass \
		 --install-location /Applications/Timeglass \
		 --scripts ./scripts/glass \
		 --version $VERSION \
		 glass.pkg

echo "Building product...."
productbuild --distribution distribution.xml \
			  Timeglass.pkg

echo "Signing product..."
productsign --sign "Developer ID Installer: A.L. VEER" Timeglass.pkg 'bin/Timeglass Setup (x64).pkg'