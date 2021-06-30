package runutil

import (
	"fmt"
	"runtime/debug"

	"github.com/zcong1993/kit/pkg/log"
	"go.uber.org/zap"
)

func Recover(logger *log.Logger, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logger.Error("recover panic", zap.String("panic", fmt.Sprintf("%s\n%s", p, string(debug.Stack()))))
	}
}

func WithRecover(fn func(), logger *log.Logger) {
	defer Recover(logger)
	fn()
}

func GoSafe(fn func(), logger *log.Logger) {
	go WithRecover(fn, logger)
}
