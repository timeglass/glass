# Timeglass
Fully Automated Time Tracking for Git Repositories. It uses hooks and file monitoring to make sure you'll never forget to start or stop your timer ever again. It is written in Go and comes with just a minimal command line interface. 

![Timeglass Screenshot](/../docs/docs/screenshot.png?raw=true "Timeglass Screenshot")

__Features:__

- The timer __automatically starts__ when you switch to a (new) branch using `git checkout`
- The timer __automatically pauses__ when it doesn't detect any file activity for a while
- The time you spent is automatically added to the next `git commit` message
- Time is incremented by a "minimal billable unit" (MBU), one minute by default. 
- No internet or account required and completely free& open source

__Currently Supported (see roadmap):__

- Platforms: __OSX__
- Version Control: __Git__

## Getting Started
*coming soon*


## Roadmap, input welcome!

- __Configuration:__ including configuring the minimal billable unit, ignoring certain directories and configuring the commit message. _What else would you like to configure? input welcome [here](https://github.com/Timeglass/glass/issues/7)_
- __Querying:__ Having intimate knowledge of both the commits and the time it took to create them, opens up enormous potential for interesting data to be extracted. _What would you like to query for? Input welcome [here](https://github.com/Timeglass/glass/issues/9)_
- __Supporting Other VCS:__ Timeglass currently only works for git repositories, mainly due to the number of hooks it provides. _What other version control systems would you like to see implemented? Input welcome [here](https://github.com/Timeglass/glass/issues/10)_
- __Supporting other OSs:__ File monitoring is implemented differently across platforms. The current implementation uses FSEvents (OSX), let me know what other platforms you would like to see implemented [here](https://github.com/Timeglass/glass/issues/11)

## Known Issues

- __Handling `git stash`:__ Git has the ability to stash work for a later commit prior to switching branches. Currently the timer unable to detect this; adding extra time to next commit. Input welcome [here](https://github.com/Timeglass/glass/issues/3)
- __OS Restarts:__ Whenever the OS shuts down the repository might still contain uncommited work and a running timer, currently the timer is not restarted when this happens. _Input on how to achieve this is welcome [here](https://github.com/Timeglass/glass/issues/8)_