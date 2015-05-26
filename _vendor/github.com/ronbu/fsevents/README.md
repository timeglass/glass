Go binding for [FSEvents](https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#//apple_ref/doc/uid/FSEvents.h-DontLinkElementID_33)

Code is based on [sdeguti's go.fsevents](https://github.com/sdegutis)
and [samjacobson's changes to fsnotify](https://github.com/samjacobson/fsnotify)

*Documentation:* [GoDoc](http://godoc.org/github.com/ronbu/fsevents)

## TODO/Limitations

* Creates new thread for every stream
* Does not give access to the whole FSEvents API
* Better testing