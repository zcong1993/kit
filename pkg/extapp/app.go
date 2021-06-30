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

type App struct {
	App          string
	Logger       *log.Logger
	Reg          prometheus.Registerer
	Registry     *prometheus.Registry
	Cmd          *cobra.Command
	G            *run.Group
	HttpProber   *prober.HTTPProbe
	StatusProber prober.Probe // grpc or http

	loggerOption         *log.Option
	shedderFactory       shedder.Factory
	breakerOptionFactory breaker.OptionFactory

	innerHttpOptions *innerHttpOptions
	innerHttpFactory innerHttpFactory
	innerHttpServer  *exthttp.MuxServer

	// user http server
	httpServer *exthttp.HttpServer
	// user grpc server
	grpcServer *extgrpc.Server
}

func NewApp() *App {
	// logger before init
	logger, err := log.DefaultLogger()
	FatalOnErrorf(err, "init logger")

	return &App{
		G:                &run.Group{},
		innerHttpOptions: &innerHttpOptions{},
		HttpProber:       prober.NewHTTP(),
		loggerOption:     &log.Option{},
		Logger:           logger,
	}
}

func (a *App) InitFromCmd(cmd *cobra.Command, name string) {
	a.App = name
	a.Cmd = cmd
	// 初始化日志
	logger, err := a.loggerOption.CreateLogger()
	FatalOnErrorf(err, "init logger")
	a.Logger = logger
	log.SyncOnClose(a.G, logger)

	// 初始化 metrics
	me := metrics.InitMetrics()
	a.Registry = me
	a.Reg = prometheus.WrapRegistererWith(prometheus.Labels{"app": name}, me)

	// 初始化 tracer
	err = oteltracing.InitTracerFromEnv(a.Logger, a.App)
	FatalOnError(err)

	a.StatusProber = prober.Combine(a.HttpProber, prober.NewInstrumentation(a.App, logger, a.Registry))

	a.innerHttpServer = a.innerHttpFactory()
}

// GetInnerHttpServer should call after InitFromCmd.
func (a *App) GetInnerHttpServer() *exthttp.MuxServer {
	return a.innerHttpServer
}

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

func (a *App) GrpcServer(opts ...extgrpc.Option) *extgrpc.Server {
	grpcProber := prober.NewGRPC()
	a.StatusProber = prober.Combine(a.StatusProber, grpcProber)

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

func (a *App) Start() error {
	// handle exit
	extrun.HandleSignal(a.G)

	// 1. start common
	if !a.innerHttpOptions.disable {
		a.innerHttpServer.Run(a.G, a.StatusProber)
	}

	// check if has http server
	if a.httpServer != nil {
		a.httpServer.Run(a.G, a.StatusProber)
	}

	// check if has grpc server
	if a.grpcServer != nil {
		a.grpcServer.Run(a.G, a.StatusProber)
	}

	return a.G.Run()
}

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

func FatalOnError(err error) {
	if err != nil {
		Fatal(err)
	}
}

func FatalOnErrorf(err error, format string, args ...interface{}) {
	if err != nil {
		Fatal(errors.Wrapf(err, format, args...))
	}
}

func Fatal(msgs ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, msgs...)
	os.Exit(1)
}
