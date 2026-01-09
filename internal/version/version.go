package version

import (
	"runtime/debug"
	"sync"
)

// Get returns the version information using build info
var Get = sync.OnceValue(func() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version
	}
	return "development"
})