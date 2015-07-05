# Timeglass

![Timeglass Screenshot](/docs/screenshot.png?raw=true "Timeglass Screenshot")

Fully automated time tracking for Git repositories. It uses hooks and file monitoring to make sure you'll never forget to start or stop your timer ever again. It is written in [Go](http://golang.org/) and runs 100% on your own workstation: no internet or account registration required. 

__Features:__

- The timer __automatically starts__ when you switch to a (new) branch using `git checkout` or upon detecting any file activity in the repository
- The timer __automatically pauses__ when it doesn't detect any file activity for a while
- The time you spent is automatically added to the next `git commit`
- The timer increments in discreet steps: the _minimal billable unit_ (MBU), by default this is 1 minute. 
- Spent time is stored as metadata using [git-notes](https://git-scm.com/docs/git-notes) and can be pushed and stored automatically to any remote repository (e.g Github)

__Currently Supported:__

- Platforms: __OSX, Linux and Windows__
- Version Control: __Git__

## Getting Started
1. Download and install the latest release using any one of your preferred methods:
	
	1. Automatic [installers](https://github.com/timeglass/glass/releases/latest) for 64bit  _OSX_ and _Windows_ 
	2. [Manual installion](/docs/manual_installation.md) with 64bit precompiled Binaries for OSX, Linux and Windows 
	3. [Manual installion](/docs/manual_installation.md) by __building from source__ for all other architectures

  _Note: For Windows, the documentation assumes you're using Git through a [bash-like CLI](https://msysgit.github.io/) but nothing about the implementation prevents you from using another approach._

2. Use your terminal to navigate to the repository that contains the project you would like to track and then initiate Timeglass:

 ```sh
 cd ~/my-git-project
 glass init
 ```
 
 _NOTE: you'll have to run this once per clone_

3. The timer starts right away but will pause soon unless it detects file activity or the checkout of a branch: 

  ```sh
  # see if the timer is running or paused:
  glass status

  # the timer keeps running while there is file activity
  echo "pretending to work..." > ./my_file.go
  
  # or branches are checked out
  git checkout -b "testing_timeglass"
  ```
  
4. Edit some files, get a coffee, and commit in order to register the time you spent:

  ```sh
  git add -A
  git commit -m "time flies when you're having fun"
  ```

5. Verify that the time was indeed registered correctly by looking at your commit log:

  ```sh
  git log -n 1 --show-notes=time-spent
  ```

## What's Next?
Now you know how to measure the time you are spending on each commit, you might want to learn more about...

- [Querying your measurements](/docs/query.md)
- [Configuring _Timeglass_](/docs/config.md)
- [Sharing data with others](/docs/sharing.md)

And ofcourse, you'll always have the options to uninstall:

- [Uninstalling Timeglass](/docs/uninstall.md)

## Roadmap, input welcome!

- __Supporting Other VCS:__ Timeglass currently only works for git repositories, mainly due to the number of hooks it requires. _What other version control systems would you like to see implemented? Input welcome [here](https://github.com/Timeglass/glass/issues/10)_

## Known Issues

- __Handling `git stash`:__ Git has the ability to stash work for a later commit prior to switching branches. Currently the timer unable to detect this; adding extra time to next commit. Input welcome [here](https://github.com/Timeglass/glass/issues/3)
- __Network Volumes:__ Projects that are kept on network volumes (e.g using NFS) are known to have flaky support for file monitoring. This means timers might error on reboot as network drives weren't available, or the automatic unpausing of the timer might be broken in such projects. *I'm looking for cases that experience such problem, or other information that might be of help over* [here](https://github.com/timeglass/glass/issues/36)

## Contributors
in alphabetical order:

- Kristof Vannotten ([kvannotten](https://github.com/kvannotten))
- Michael Mior ([michaelmior](https://github.com/michaelmior))
