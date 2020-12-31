package runutil

import (
	"fmt"
	"runtime/debug"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func Recover(logger log.Logger, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		level.Error(logger).Log("msg", "recover panic", "panic", fmt.Sprintf("%s\n%s", p, string(debug.Stack())))
	}
}

func WithRecover(fn func(), logger log.Logger) {
	defer Recover(logger)
	fn()
}

func GoSafe(fn func(), logger log.Logger) {
	go WithRecover(fn, logger)
}
