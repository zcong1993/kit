package extapp

import (
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/pflag"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/exthttp"
)

// 内部 http 服务, 一般用来暴露存活, 健康检查路由
// 也可以选择 metrics 和 pprof

var (
	withPprof   = "inner.with-pprof"
	withMetrics = "inner.with-metrics"
	addr        = "inner.addr"
)

// RegisterInnerHttpServerFlags 注册相关 flags.
func RegisterInnerHttpServerFlags(flagSet *pflag.FlagSet) {
	flagSet.Bool(withPprof, false, "If enable pprof routes.")
	flagSet.Bool(withMetrics, true, "If expose metrics router.")
	flagSet.String(addr, ":6060", "Inner http server addr.")
}

// StartInnerHttpServer 启动服务, 会根据 flags 参数调整功能.
func StartInnerHttpServer(app *App, httpProber *prober.HTTPProbe, opts ...exthttp.OptionFunc) *exthttp.MuxServer {
	addrVar, err := app.Cmd.Flags().GetString(addr)
	FatalOnErrorf(err, "get %s error", addr)

	withPprofVar, err := app.Cmd.Flags().GetBool(withPprof)
	FatalOnErrorf(err, "get %s error", withPprof)

	withMetricsVar, err := app.Cmd.Flags().GetBool(withMetrics)
	FatalOnErrorf(err, "get %s error", withMetrics)

	opts = append([]exthttp.OptionFunc{
		exthttp.WithListen(addrVar),
		exthttp.WithServiceName("metrics/profiler"),
	}, opts...)

	profileServer := exthttp.NewMuxServer(app.Logger, opts...)

	if withPprofVar {
		level.Info(app.Logger).Log("msg", "register pprof routes")
		profileServer.RegisterProfiler()
	}

	if withMetricsVar {
		profileServer.RegisterMetrics(app.Registry)
	}

	profileServer.RegisterProber(httpProber)

	profileServer.RunGroup(app.G)

	return profileServer
}
