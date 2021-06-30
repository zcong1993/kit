package extapp

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/zcong1993/kit/pkg/server/exthttp"
)

// 内部 http 服务, 一般用来暴露存活, 健康检查路由
// 也可以选择 metrics 和 pprof.

var (
	withPprof      = "inner.with-pprof"
	withMetrics    = "inner.with-metrics"
	addr           = "inner.addr"
	disable        = "inner.disable"
	gracePeriod    = "inner.grace-period"
	withLogControl = "inner.log-control"
)

type innerHTTPOptions struct {
	withPprof      bool
	withMetrics    bool
	addr           string
	disable        bool
	gracePeriod    time.Duration
	withLogControl bool
}

type innerHTTPFactory = func() *exthttp.MuxServer

func registerInnerHTTP(app *App, cmd *cobra.Command) innerHTTPFactory {
	f := cmd.PersistentFlags()

	f.BoolVar(&app.innerHTTPOptions.withPprof, withPprof, false, "If enable pprof routes, /debug/pprof/*")
	f.BoolVar(&app.innerHTTPOptions.withMetrics, withMetrics, true, "If expose metrics router, /metrics")
	f.StringVar(&app.innerHTTPOptions.addr, addr, ":6060", "Inner metrics/pprof http server addr")
	f.BoolVar(&app.innerHTTPOptions.disable, disable, false, "If disable inner http server")
	f.DurationVar(&app.innerHTTPOptions.gracePeriod, gracePeriod, 0, "Inner http exit grace period")
	f.BoolVar(&app.innerHTTPOptions.withLogControl, withLogControl, false, "If enable logger level control, GET PUT /log/level")

	return func() *exthttp.MuxServer {
		profileServer := exthttp.NewMuxServer(app.Logger, exthttp.WithListen(app.innerHTTPOptions.addr), exthttp.WithServiceName("metrics/profiler"), exthttp.WithGracePeriod(app.innerHTTPOptions.gracePeriod))

		if app.innerHTTPOptions.withPprof {
			app.Logger.Info("register pprof routes, /debug/pprof/*")
			profileServer.RegisterProfiler()
		}

		if app.innerHTTPOptions.withMetrics {
			app.Logger.Info("register metrics route, /metrics")
			profileServer.RegisterMetrics(app.registry)
		}

		if app.innerHTTPOptions.withLogControl {
			app.Logger.Info("register log control route, GET PUT /log/level")
			profileServer.RegisterLogControl(app.loggerOption.GetLevel())
		}

		profileServer.RegisterProber(app.httpProber)

		return profileServer
	}
}
