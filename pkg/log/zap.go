package log

import (
	"context"

	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Logger = zap.Logger

type Option struct {
	LogLevel    string
	LogEncoding string
	Caller      bool
	Prod        bool
}

func DefaultLogger() (*Logger, error) {
	defaultOption := &Option{
		LogLevel:    "info",
		LogEncoding: "json",
		Caller:      false,
		Prod:        true,
	}
	return NewLogger(defaultOption)
}

var NewNopLogger = zap.NewNop

func NewLogger(o *Option, opts ...zap.Option) (*Logger, error) {
	var config zap.Config

	isEmptyEncoding := len(o.LogEncoding) == 0

	if o.Prod {
		config = zap.NewProductionConfig()
		if isEmptyEncoding {
			o.LogEncoding = "json"
		}
	} else {
		config = zap.NewDevelopmentConfig()
		if isEmptyEncoding {
			o.LogEncoding = "console"
		}
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	var level zap.AtomicLevel
	err := level.UnmarshalText([]byte(o.LogLevel))

	if err != nil {
		return nil, errors.Wrap(err, "invalid log.level")
	}

	if !o.Caller {
		config.DisableCaller = true
	}

	config.Encoding = o.LogEncoding

	return config.Build(opts...)
}

func (opt *Option) Register(app *cobra.Command) {
	f := app.PersistentFlags()

	f.StringVar(&opt.LogLevel, "log.level", "info", "Log level, one of debug info warn error")
	f.StringVar(&opt.LogEncoding, "log.encoding", "", "Zap log encoding, console json or others you registered")
	f.BoolVar(&opt.Caller, "log.caller", false, "If with caller field")
	f.BoolVar(&opt.Prod, "log.prod", false, "If use zap log production preset")
}

func (opt *Option) CreateLogger(opts ...zap.Option) (*Logger, error) {
	return NewLogger(opt, opts...)
}

func SyncOnClose(g *run.Group, logger *Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(err error) {
		_ = logger.Sync()
		logger.Debug("sync logger before close")
		cancel()
	})
}
