package extapp

import (
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/server/exthttp"
)

// 内部 http 服务, 一般用来暴露存活, 健康检查路由
// 也可以选择 metrics 和 pprof

var (
	withPprof   = "inner.with-pprof"
	withMetrics = "inner.with-metrics"
	addr        = "inner.addr"
	disable     = "inner.disable"
	gracePeriod = "inner.grace-period"
)

type innerHttpOptions struct {
	withPprof   bool
	withMetrics bool
	addr        string
	disable     bool
	gracePeriod time.Duration
}

type innerHttpFactory = func() *exthttp.MuxServer

func registerInnerHttp(app *App, cmd *cobra.Command) innerHttpFactory {
	f := cmd.PersistentFlags()

	f.BoolVar(&app.innerHttpOptions.withPprof, withPprof, false, "If enable pprof routes, /debug/pprof/*.")
	f.BoolVar(&app.innerHttpOptions.withMetrics, withMetrics, true, "If expose metrics router, /metrics.")
	f.StringVar(&app.innerHttpOptions.addr, addr, ":6060", "Inner metrics/pprof http server addr.")
	f.BoolVar(&app.innerHttpOptions.disable, disable, false, "If disable inner http server.")
	f.DurationVar(&app.innerHttpOptions.gracePeriod, gracePeriod, 0, "Inner http exit grace period.")

	return func() *exthttp.MuxServer {
		profileServer := exthttp.NewMuxServer(app.Logger, exthttp.WithListen(app.innerHttpOptions.addr), exthttp.WithServiceName("metrics/profiler"), exthttp.WithGracePeriod(app.innerHttpOptions.gracePeriod))

		if app.innerHttpOptions.withPprof {
			level.Info(app.Logger).Log("msg", "register pprof routes")
			profileServer.RegisterProfiler()
		}

		if app.innerHttpOptions.withMetrics {
			profileServer.RegisterMetrics(app.Registry)
		}

		profileServer.RegisterProber(app.HttpProber)

		return profileServer
	}
}
