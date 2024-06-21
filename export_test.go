package contextprovider

import "runtime"

func StubRuntimeCaller() (restore func()) {
	runtimeCaller = func(_ int) (pc uintptr, file string, line int, ok bool) {
		return
	}
	return func() {
		runtimeCaller = runtime.Caller
	}
}
