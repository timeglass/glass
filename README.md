# Timeglass

![Timeglass Screenshot](/docs/screenshot.png?raw=true "Timeglass Screenshot")

Fully automated time tracking for Git repositories. It uses hooks and file monitoring to make sure you'll never forget to start or stop your timer ever again. It is written in [Go](http://golang.org/) and runs 100% on your own workstation: no internet or account registration required. 

__Features:__

- The timer __automatically starts__ when you switch to a (new) branch using `git checkout`
- The timer __automatically pauses__ when it doesn't detect any file activity for a while
- The time you spent is automatically added to the next `git commit`
- The timer increments in discreet steps: the _minimal billable unit_ (MBU), by default this is 1m. 
- Spent time is stored as metadata using [git-notes](https://git-scm.com/docs/git-notes) and pushed automatically

__Currently Supported:__

- Platforms: __OSX, Linux and Windows__
- Version Control: __Git__

## Getting Started
1. Download the [latest release](https://github.com/timeglass/glass/releases/latest) for your platform and unzip the contents into a directory that is in your systems PATH (e.g /usr/local/bin).   

  _Note 1: We currently only support 64-bit prebuild binaries, for other architectures please build from source (see below)._  
  _Note 2: For Windows, the documentation assumes you're using Git through a [bash-like CLI](https://msysgit.github.io/) but nothing about the implementation prevents you from using another approach._

2. Use your terminal to navigate to the repository that contains the project you would like to track and install the hooks:

 ```sh
 cd ~/my-git-project
 glass init
 ```
 
 _NOTE: you'll have to run this once per clone_

3. Start the timer by creating a new branch: 

  ```sh
  git checkout -b "testing_timeglass"
  ```
  
4. Edit some files, get a coffee, and commit in order to register the time you spent:

  ```sh
  git add -A
  git commit -m "time flies when you're having fun"
  ```

5. Verify that the time was indeed registered correctly by looking at your commit log:

  ```sh
  git log -n 1
  ```

## What's Next?
Now you know how to measure the time you are spending on each commit, you might want to learn more about...

- [Querying your measurements](/docs/query.md)
- [Configuring _Timeglass_](/docs/config.md)
- [Sharing data with others](/docs/sharing.md)

And ofcourse, you'll always have the options to uninstall:

- [Uninstalling Timeglass](/docs/uninstall.md)

## Building from Source
First, you'll need install the go toolchain, instructions are [here](https://golang.org/doc/install). With Go installed you can simply run `go get` for _both_ binaries:

```
go get -u github.com/timeglass/glass
go get -u github.com/timeglass/glass/glass-daemon
```

The source code will now be in your workspace and binaries are found in `$GOPATH/bin`, happy hacking!

## Roadmap, input welcome!

- __Supporting Other VCS:__ Timeglass currently only works for git repositories, mainly due to the number of hooks it provides. _What other version control systems would you like to see implemented? Input welcome [here](https://github.com/Timeglass/glass/issues/10)_

## Known Issues

- __Handling `git stash`:__ Git has the ability to stash work for a later commit prior to switching branches. Currently the timer unable to detect this; adding extra time to next commit. Input welcome [here](https://github.com/Timeglass/glass/issues/3)
- __OS Restarts:__ Whenever the OS shuts down the repository might still contain uncommited work and a running timer, currently the timer is not restarted when this happens. _Input on how to achieve this is welcome [here](https://github.com/Timeglass/glass/issues/8)_
- __Network Volumes:__ Projects that are kept on network volumes (e.g using NFS) are known to have flaky support for file monitoring, especially on Linux (using inotify). This means automatic unpausing of the timer when editing a file might be broken in such projects. *I'm looking for cases that experience such problem, or other information that might be of help over* [here](https://github.com/timeglass/glass/issues/36)

## Contributors
in alphabetical order:

- Kristof Vannotten ([kvannotten](https://github.com/kvannotten))
- Michael Mior ([michaelmior](https://github.com/michaelmior))
