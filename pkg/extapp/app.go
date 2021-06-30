package extapp

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/zcong1993/kit/pkg/breaker"
	"github.com/zcong1993/kit/pkg/extrun"
	"github.com/zcong1993/kit/pkg/log"
	"github.com/zcong1993/kit/pkg/metrics"
	"github.com/zcong1993/kit/pkg/prober"
	"github.com/zcong1993/kit/pkg/server/extgrpc"
	"github.com/zcong1993/kit/pkg/server/exthttp"
	"github.com/zcong1993/kit/pkg/shedder"
	oteltracing "github.com/zcong1993/kit/pkg/tracing/otel"
)

// App is kit core application.
type App struct {
	// App app name.
	App string
	// Logger is zap logger instance.
	Logger *log.Logger
	// Reg is prometheus Registerer.
	Reg prometheus.Registerer
	// G is run.Group handle goroutines exit.
	G *run.Group

	// registry is prometheus Registry.
	registry             *prometheus.Registry
	loggerOption         *log.Option
	shedderFactory       shedder.Factory
	breakerOptionFactory breaker.OptionFactory

	httpProber *prober.HTTPProbe
	// statusProber report healthy status and readiness status to inner http routes
	// and metrics, and you can add your prober.
	statusProber     prober.Probe // grpc or http.
	innerHttpOptions *innerHttpOptions
	innerHttpFactory innerHttpFactory
	innerHttpServer  *exthttp.MuxServer

	// user http server
	httpServer *exthttp.HttpServer
	// user grpc server
	grpcServer *extgrpc.Server
}

// NewApp create new App instance.
func NewApp() *App {
	// logger before init.
	logger, err := log.DefaultLogger()
	FatalOnErrorf(err, "init logger")

	return &App{
		G:                &run.Group{},
		innerHttpOptions: &innerHttpOptions{},
		httpProber:       prober.NewHTTP(),
		loggerOption:     &log.Option{},
		Logger:           logger,
	}
}

// InitFromCmd init components which depend commandline options
// should always be called at first of cobra.Command.Run.
func (a *App) InitFromCmd(cmd *cobra.Command, name string) {
	a.App = name
	// 初始化日志
	logger, err := a.loggerOption.CreateLogger()
	FatalOnErrorf(err, "init logger")
	a.Logger = logger
	log.SyncOnClose(a.G, logger)

	// 初始化 metrics
	me := metrics.InitMetrics()
	a.registry = me
	a.Reg = prometheus.WrapRegistererWith(prometheus.Labels{"app": name}, me)

	// 初始化 tracer
	err = oteltracing.InitTracerFromEnv(a.Logger, a.App)
	FatalOnError(err)

	a.statusProber = prober.Combine(a.httpProber, prober.NewInstrumentation(a.App, logger, a.registry))

	a.innerHttpServer = a.innerHttpFactory()
}

// AddProber can combine your prober.
func (a *App) AddProber(p prober.Probe) {
	a.statusProber = prober.Combine(a.statusProber, p)
}

// GetInnerHttpServer get inner metrice pprof health http server
// should call after InitFromCmd.
func (a *App) GetInnerHttpServer() *exthttp.MuxServer {
	return a.innerHttpServer
}

// GinServer create a gin server with many components controlled by commandline options.
func (a *App) GinServer(opts ...exthttp.OptionFunc) *gin.Engine {
	r := gin.Default()
	r.Use(metrics.NewInstrumentationMiddleware(a.Reg))
	r.Use(oteltracing.GinMiddleware(a.App))
	// shedder 中间件
	shedder.RegisterGinShedder(r, a.shedderFactory(), a.Logger)
	// breaker 中间件
	breaker.RegisterGinBreaker(r, a.Logger, a.breakerOptionFactory())

	a.httpServer = exthttp.NewHttpServer(r, a.Logger, opts...)

	return r
}

// GrpcServer create a grpc server with many components controlled by commandline options.
func (a *App) GrpcServer(opts ...extgrpc.Option) *extgrpc.Server {
	grpcProber := prober.NewGRPC()
	a.statusProber = prober.Combine(a.statusProber, grpcProber)

	o := []extgrpc.Option{
		metrics.WithServerMetrics(a.Logger, a.Reg),
		extgrpc.WithOtelTracing(),
		shedder.WithGrpcShedder(a.Logger, a.shedderFactory()),
		breaker.WithGrpcServerBreaker(a.Logger, a.breakerOptionFactory()),
	}

	o = append(o, opts...)

	s := extgrpc.NewServer(a.Logger, grpcProber, o...)

	a.grpcServer = s

	return s
}

// Start start our application.
func (a *App) Start() error {
	// handle exit
	extrun.HandleSignal(a.G)

	// 1. start common
	if !a.innerHttpOptions.disable {
		a.innerHttpServer.Run(a.G, a.statusProber)
	}

	// check if has http server
	if a.httpServer != nil {
		a.httpServer.Run(a.G, a.statusProber)
	}

	// check if has grpc server
	if a.grpcServer != nil {
		a.grpcServer.Run(a.G, a.statusProber)
	}

	return a.G.Run()
}

// Run register all the commandline flags
// should be called in main function.
func (a *App) Run(cmd *cobra.Command) {
	// 注册日志相关 flag
	a.loggerOption.Register(cmd)

	// 注册 shedder flag
	a.shedderFactory = shedder.Register(cmd)

	// 注册 breaker flag
	a.breakerOptionFactory = breaker.Register(cmd)

	// 注册 inner app server
	a.innerHttpFactory = registerInnerHttp(a, cmd)

	FatalOnError(cmd.Execute())
}

// FatalOnError log fatal and exit with code 1.
func FatalOnError(err error) {
	if err != nil {
		Fatal(err)
	}
}

// FatalOnErrorf call fatal with format.
func FatalOnErrorf(err error, format string, args ...interface{}) {
	if err != nil {
		Fatal(errors.Wrapf(err, format, args...))
	}
}

// Fatal log fatal and exit with code 1.
func Fatal(msgs ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, msgs...)
	os.Exit(1)
}
