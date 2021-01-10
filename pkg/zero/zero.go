package zero

import (
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
	"github.com/tal-tech/go-zero/core/syncx"
)

var Metrics *stat.Metrics

func SetupMetrics() {
	syncx.Once(func() {
		setupZero()
		Metrics = stat.NewMetrics("default")
	})
}

func setupZero() {
	stat.SetReporter(nil)
	logx.MustSetup(logx.LogConf{
		ServiceName:         "default",
		Mode:                "console",
		Level:               "info",
		StackCooldownMillis: 100,
	})
}
