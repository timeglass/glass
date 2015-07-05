# Manual Installation
If you're on linx and/or like to have more control over how you install Timeglass we provide two ways for manual installation. If you have a 64bit architecture it is possible to just [download the prebuild binaries](https://github.com/timeglass/glass/releases/latest) and go to step 2, else you should build from source.

## Step 1: Building from Source (optional)

First, you'll need install the go toolchain, instructions are [here](https://golang.org/doc/install). With Go installed you can simply run `go get` for _both_ binaries:

```
go get -u github.com/timeglass/glass
go get -u github.com/timeglass/glass/glass-daemon
```

The source code will now be in your workspace and binaries are found in `$GOPATH/bin`. If you want to rebuild the binaries at a later point your can use the build script in the root of the project like so: `./make.bash build`

## Step 2: Placing the Binaries in your PATH

If you downloaded the prebuild binaries you'll first need to unzip them and the copy the contents to a directory thats in your PATH (e.g. /usr/local/bin). If you build from source the binaries are probably already in your path, if not you must copy or link them there yourself. 

Check that this was done successfull by opening a new terminal window and running: `glass`. You should see all the *Timeglass* commands that are now available to you.


## Step 3: Installing the Background Service

An important part of *Timeglass* is the monitoring of file system activity. For this it needs a small background process that is always running, lucky this service can be installed and started by running a single command: `glass install`, if the output is empty everything went as expected.

_NOTE1: **On OSX and linux** this service is currently installed for all accounts and requires you to use sudo: `sudo glass install`_

_NOTE2: **On Windows** this service requires administration privileges so either 'run as Administrator' or log in as the Administrator and run the install command_ 
