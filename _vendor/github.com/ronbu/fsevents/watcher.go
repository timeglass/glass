package fsevents

/*
#cgo LDFLAGS: -framework CoreServices
#include <CoreServices/CoreServices.h>
FSEventStreamRef fswatch_create(
	FSEventStreamContext*,
	CFMutableArrayRef,
	FSEventStreamEventId,
	CFTimeInterval,
	FSEventStreamCreateFlags);
FSEventStreamRef fswatch_create_relative_to_device(
	dev_t,
	FSEventStreamContext*,
	CFMutableArrayRef,
	FSEventStreamEventId,
	CFTimeInterval,
	FSEventStreamCreateFlags);
static CFMutableArrayRef fswatch_make_mutable_array() {
  return CFArrayCreateMutable(NULL, 0, &kCFTypeArrayCallBacks);
}

*/
import "C"
import (
	"time"

	"unsafe"
)

type CreateFlags uint32

const (
	CF_NONE       CreateFlags = 0
	CF_USECFTYPES CreateFlags = 1 << (iota - 1) //this flag is ignored
	CF_NODEFER
	CF_WATCHROOT
	CF_IGNORESELF
	CF_FILEEVENTS
)

type EventFlags uint32

const (
	EF_NONE            EventFlags = 0
	EF_MUSTSCANSUBDIRS EventFlags = 1 << (iota - 1)
	EF_USERDROPPED
	EF_KERNELDROPPED
	EF_EVENTIDSWRAPPED
	EF_HISTORYDONE
	EF_ROOTCHANGED
	EF_MOUNT
	EF_UNMOUNT

	EF_CREATED
	EF_REMOVED
	EF_INODEMETAMOD
	EF_RENAMED
	EF_MODIFIED
	EF_FINDERINFOMOD
	EF_CHANGEOWNER
	EF_XATTRMOD
	EF_ISFILE
	EF_ISDIR
	EF_ISSYMLINK
)

type EventID C.FSEventStreamEventId

// EventID has type UInt64 but this constant is represented as -1
// which is represented by 63 1's in memory
const NOW EventID = (1 << 64) - 1
const ALL EventID = 0

// UUID representing a Mountpoint in OS X.
// fi, _ := os.Stat("")
// dev := Device(fi.Sys().(*syscall.Stat_t).Dev)
type Device C.dev_t

type Stream struct {
	Chan    chan []Event
	cstream C.FSEventStreamRef
	runloop C.CFRunLoopRef
}

type Event struct {
	Id    EventID
	Path  string
	Flags EventFlags
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_4
func Current() EventID {
	return EventID(C.FSEventsGetCurrentEventId())
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_5
func LastEventBefore(dev Device, ts time.Time) EventID {
	return EventID(
		C.FSEventsGetLastEventIdForDeviceBeforeTime(
			C.dev_t(dev),
			C.CFAbsoluteTime(ts.Unix())))
}

// TODO: FSEventsPurgeEventsForDeviceUpToEventId

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_8
func (s Stream) Paths() []string {
	cpaths := C.FSEventStreamCopyPathsBeingWatched(s.cstream)
	defer C.CFRelease(C.CFTypeRef(cpaths))
	count := C.CFArrayGetCount(cpaths)
	paths := make([]string, count)
	var i C.CFIndex
	for ; i < count; i++ {
		cpath := C.CFStringRef(C.CFArrayGetValueAtIndex(cpaths, i))
		paths[i] = fromCFString(cpath)
	}
	return paths
}

func fromCFString(cstr C.CFStringRef) string {
	defer C.CFRelease(C.CFTypeRef(cstr))

	var (
		buf  []C.char
		ok   C.Boolean
		size uint = 1024
	)
	for ok == C.FALSE {
		buf = make([]C.char, size)
		ok = C.CFStringGetCString(cstr, &buf[0],
			C.CFIndex(len(buf)), C.kCFStringEncodingUTF8)
		size *= 2
	}
	return C.GoString(&buf[0])
}

func New(dev Device, since EventID, interval time.Duration, flags CreateFlags,
	paths ...string) *Stream {

	cpaths := C.fswatch_make_mutable_array()
	defer C.free(unsafe.Pointer(cpaths))
	for _, dir := range paths {
		path := C.CString(dir)
		defer C.free(unsafe.Pointer(path))
		str := C.CFStringCreateWithCString(nil, path, C.kCFStringEncodingUTF8)
		defer C.free(unsafe.Pointer(str))
		C.CFArrayAppendValue(cpaths, unsafe.Pointer(str))
	}

	csince := C.FSEventStreamEventId(since)
	cinterval := C.CFTimeInterval(interval / time.Second)
	cflags := C.FSEventStreamCreateFlags(flags &^ CF_USECFTYPES)

	s := new(Stream)
	s.Chan = make(chan []Event)

	ctx := C.FSEventStreamContext{info: unsafe.Pointer(&s.Chan)}

	var cstream C.FSEventStreamRef
	if dev == 0 {
		cstream = C.fswatch_create(&ctx, cpaths, csince, cinterval, cflags)
	} else {
		cdev := C.dev_t(dev)
		cstream = C.fswatch_create_relative_to_device(
			cdev, &ctx, cpaths, csince, cinterval, cflags)
	}
	s.cstream = cstream

	return s
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_14
func (s Stream) LatestEventID() EventID {
	return EventID(C.FSEventStreamGetLatestEventId(s.cstream))
}

// Schedules the Stream on a new thread
// and starts delivering events on the channel.
//
// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_20
// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_17
func (s *Stream) Start() bool {
	type watchSuccessData struct {
		runloop C.CFRunLoopRef
		stream  Stream
	}

	successChan := make(chan C.CFRunLoopRef)

	go func() {
		C.FSEventStreamScheduleWithRunLoop(s.cstream,
			C.CFRunLoopGetCurrent(), C.kCFRunLoopCommonModes)
		ok := C.FSEventStreamStart(s.cstream) != C.FALSE
		if ok {
			successChan <- C.CFRunLoopGetCurrent()
			C.CFRunLoopRun()
		} else {
			successChan <- nil
		}
	}()

	runloop := <-successChan

	if runloop == nil {
		return false
	}
	s.runloop = runloop
	return true
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_12
func (s Stream) Flush() {
	C.FSEventStreamFlushSync(s.cstream)
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_11
func (s Stream) FlushAsync() EventID {
	return EventID(C.FSEventStreamFlushAsync(s.cstream))
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_13
func (s Stream) Device() Device {
	return Device(C.FSEventStreamGetDeviceBeingWatched(s.cstream))
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_21
func (s Stream) Stop() {
	C.FSEventStreamStop(s.cstream)
}

// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_15
func (s Stream) Invalidate() {
	C.FSEventStreamInvalidate(s.cstream)
}

// Releases all resources and closes the channel
// https://developer.apple.com/library/mac/documentation/Darwin/Reference/FSEvents_Ref/Reference/reference.html#jumpTo_16
func (s Stream) Release() {
	C.FSEventStreamRelease(s.cstream)
	C.CFRunLoopStop(s.runloop)
}

// Convenience function: flushes, stops, invalidates and releases the stream
func (s Stream) Close() {
	s.Stop()
	s.Invalidate()
	s.Release()
}

//export goCallback
func goCallback(stream C.FSEventStreamRef, info unsafe.Pointer,
	count C.size_t, paths **C.char,
	flags *C.FSEventStreamEventFlags, ids *C.FSEventStreamEventId) {

	var events []Event
	for i := 0; i < int(count); i++ {
		cpaths := uintptr(unsafe.Pointer(paths)) + (uintptr(i) * unsafe.Sizeof(*paths))
		cpath := *(**C.char)(unsafe.Pointer(cpaths))

		cflags := uintptr(unsafe.Pointer(flags)) + (uintptr(i) * unsafe.Sizeof(*flags))
		cflag := *(*C.FSEventStreamEventFlags)(unsafe.Pointer(cflags))

		cids := uintptr(unsafe.Pointer(ids)) + (uintptr(i) * unsafe.Sizeof(*ids))
		cid := *(*C.FSEventStreamEventId)(unsafe.Pointer(cids))

		events = append(events, Event{
			Id:    EventID(cid),
			Path:  C.GoString(cpath),
			Flags: EventFlags(cflag),
		})
	}

	ch := *((*chan []Event)(info))
	ch <- events
}
