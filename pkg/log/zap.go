package log

import (
	"context"

	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Logger is type alias for zap.Logger.
type Logger = zap.Logger

// Option is Logger options.
type Option struct {
	// LogLevel is zap logger level.
	LogLevel string
	// LogEncoding is zap LogEncoding.
	LogEncoding string
	// Caller control if add caller to log.
	Caller bool
	// Prod if true use zap production preset.
	Prod bool
	// level store zap AtomicLevel for register http control router.
	level zap.AtomicLevel
}

// DefaultLogger create a default logger, if helpful before logger init.
func DefaultLogger() (*Logger, error) {
	defaultOption := &Option{
		LogLevel:    "info",
		LogEncoding: "json",
		Caller:      false,
		Prod:        true,
	}
	return NewLogger(defaultOption)
}

// NewNopLogger is alias for zap.NewNop.
var NewNopLogger = zap.NewNop

// NewLogger create a zap logger by option.
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

	err := o.level.UnmarshalText([]byte(o.LogLevel))
	config.Level = o.level

	if err != nil {
		return nil, errors.Wrap(err, "invalid log.level")
	}

	if !o.Caller {
		config.DisableCaller = true
	}

	config.Encoding = o.LogEncoding

	return config.Build(opts...)
}

// Register register flags to cobra global flag set.
func (opt *Option) Register(app *cobra.Command) {
	f := app.PersistentFlags()

	f.StringVar(&opt.LogLevel, "log.level", "info", "Log level, one of zap level")
	f.StringVar(&opt.LogEncoding, "log.encoding", "", "Zap log encoding, console json or others you registered")
	f.BoolVar(&opt.Caller, "log.caller", false, "If with caller field")
	f.BoolVar(&opt.Prod, "log.prod", false, "If use zap log production preset")
}

// CreateLogger create zap logger from self option.
func (opt *Option) CreateLogger(opts ...zap.Option) (*Logger, error) {
	return NewLogger(opt, opts...)
}

// GetLevel get inner zap AtomicLevel.
func (opt *Option) GetLevel() zap.AtomicLevel {
	return opt.level
}

// SyncOnClose call log.Sync before run group exit.
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
