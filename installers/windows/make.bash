#! /bin/bash

cp ../../bin/windows_amd64/* ./

VERSION=$(cat ../../VERSION)
sed -i "s/\sVersion=".*"/ Version=\"$VERSION\"/" Product.wxs

$WINDIR/Microsoft.NET/Framework/v4.0.30319/MSBuild.exe glass.wixproject