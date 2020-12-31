package log

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"

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

// DefaultLogger return a simple logger, usually use before command logger init
func DefaultLogger() klog.Logger {
	logger := klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	logger = klog.With(logger, "ts", timestampFormat, "caller", klog.DefaultCaller)
	return logger
}

func Register(flagSet *pflag.FlagSet) {
	flagSet.String("log.level", "info", "Log level")
	flagSet.String("log.format", "logfmt", "Log format")
	flagSet.Bool("log.caller", false, "If with caller field")
}

func NewLogger(cmd *cobra.Command) (klog.Logger, error) {
	format, err := cmd.Flags().GetString("log.format")
	if err != nil {
		return nil, err
	}

	if !contains(format, []string{"logfmt", "json"}) {
		return nil, errors.New("invalid log format")
	}

	var l klog.Logger
	if format == "json" {
		l = klog.NewJSONLogger(klog.NewSyncWriter(os.Stderr))
	} else {
		l = klog.NewLogfmtLogger(klog.NewSyncWriter(os.Stderr))
	}

	lvStr, err := cmd.Flags().GetString("log.level")
	if err != nil {
		return nil, err
	}

	var lv level.Option
	switch lvStr {
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

	caller, err := cmd.Flags().GetBool("log.caller")
	if err != nil {
		return nil, err
	}

	l = level.NewFilter(l, lv)
	l = klog.With(l, "ts", timestampFormat)

	if caller {
		l = klog.With(l, "caller", klog.DefaultCaller)
	}
	return l, nil
}

func MustNewLogger(cmd *cobra.Command) klog.Logger {
	logger, err := NewLogger(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed init logger:  %s\n", err)
		os.Exit(2)
	}
	return logger
}

func contains(val string, enums []string) bool {
	for _, v := range enums {
		if val == v {
			return true
		}
	}
	return false
}
