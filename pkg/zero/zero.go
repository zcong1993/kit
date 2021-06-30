package zero

import (
	"sync"

	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stat"
)

var (
	// Metrics is global go-zero metrics.
	Metrics *stat.Metrics
	once    sync.Once
)

// SetupMetrics 创建一个全局 metrics.
// SetupMetrics create a global metrics.
func SetupMetrics() {
	once.Do(func() {
		setupZero()
		Metrics = stat.NewMetrics("default")
	})
}

// 处理外部依赖, go-zero 部分需要先初始化.
func setupZero() {
	stat.SetReporter(nil)
	logx.Disable()
}
