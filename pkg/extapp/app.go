package extapp

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/metrics"
	"github.com/zcong1993/x/pkg/shedder"
	"github.com/zcong1993/x/pkg/tracing/register"
)

type App struct {
	App      string
	Logger   log.Logger
	Reg      prometheus.Registerer
	Registry *prometheus.Registry
	Tracer   opentracing.Tracer
	Cmd      *cobra.Command
	G        *run.Group
}

func NewFromCmd(cmd *cobra.Command) *App {
	name, err := cmd.Flags().GetString("app")
	FatalOnError(err)

	// 初始化日志
	logger := log2.MustNewLogger(cmd)
	logger = log.WithPrefix(logger, "app", name)

	me := metrics.InitMetrics()
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"app": name}, me)

	var g run.Group

	// 初始化 tracer
	tracer := register.MustInitTracer(&g, cmd, logger, me)

	return &App{
		App:      name,
		Logger:   logger,
		Reg:      reg,
		Registry: me,
		Tracer:   tracer,
		Cmd:      cmd,
		G:        &g,
	}
}

func RunDefaultServerApp(app *cobra.Command) {
	// 注册 app name flag
	app.PersistentFlags().String("app", "", "App name")
	cobra.MarkFlagRequired(app.PersistentFlags(), "app")
	// 注册日志相关 flag
	log2.Register(app.PersistentFlags())
	// 注册 tracing flag
	register.RegisterFlags(app.PersistentFlags())

	// 注册 shedder flag
	shedder.Register(app.PersistentFlags())

	// 注册 inner app server
	RegisterInnerHttpServerFlags(app.PersistentFlags())

	FatalOnError(app.Execute())
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
	fmt.Fprintln(os.Stderr, msgs...)
	os.Exit(1)
}
