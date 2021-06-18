package log

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	klog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// This timestamp format differs from RFC3339Nano by using .000 instead
// of .999999999 which changes the timestamp from 9 variable to 3 fixed
// decimals (.130 instead of .130987456).
var timestampFormat = klog.TimestampFormat(
	func() time.Time { return time.Now().UTC() },
	"2006-01-02T15:04:05.000Z07:00",
)

// DefaultLogger return a simple logger, usually use before command logger init.
func DefaultLogger() klog.Logger {
	logger := klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	logger = klog.With(logger, "ts", timestampFormat, "caller", klog.DefaultCaller)
	return logger
}

type Option struct {
	LogLevel  string
	LogFormat string
	Caller    bool
}

func NewLogger(opt *Option) (klog.Logger, error) {
	if !contains(opt.LogFormat, []string{"logfmt", "json"}) {
		return nil, errors.New("invalid log format")
	}

	var l klog.Logger
	if opt.LogFormat == "json" {
		l = klog.NewJSONLogger(klog.NewSyncWriter(os.Stderr))
	} else {
		l = klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	}

	var lv level.Option
	switch opt.LogLevel {
	case "debug":
		lv = level.AllowDebug()
	case "info":
		lv = level.AllowInfo()
	case "warn":
		lv = level.AllowWarn()
	case "error":
		lv = level.AllowError()
	default:
		return nil, errors.New("invalid log level")
	}

	l = level.NewFilter(l, lv)
	l = klog.With(l, "ts", timestampFormat)

	if opt.Caller {
		l = klog.With(l, "caller", klog.DefaultCaller)
	}

	return l, nil
}

func (opt *Option) Register(app *cobra.Command) {
	f := app.PersistentFlags()

	f.StringVar(&opt.LogLevel, "log.level", "info", "Log level")
	f.StringVar(&opt.LogFormat, "log.format", "logfmt", "Log format")
	f.BoolVar(&opt.Caller, "log.caller", false, "If with caller field")
}

func (opt *Option) CreateLogger() (klog.Logger, error) {
	return NewLogger(opt)
}

func contains(val string, enums []string) bool {
	for _, v := range enums {
		if val == v {
			return true
		}
	}
	return false
}
