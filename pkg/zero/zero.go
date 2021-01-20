package zero

import (
	"sync"

	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
)

var (
	Metrics *stat.Metrics
	once    sync.Once
)

// SetupMetrics 创建一个全局 metrics.
func SetupMetrics() {
	once.Do(func() {
		setupZero()
		Metrics = stat.NewMetrics("default")
	})
}

// 处理外部依赖, go-zero 部分需要先初始化.
func setupZero() {
	stat.SetReporter(nil)
	//logx.MustSetup(logx.LogConf{
	//	ServiceName:         "default",
	//	Mode:                "console",
	//	Level:               "info",
	//	StackCooldownMillis: 100,
	//})
	logx.Disable()
}
