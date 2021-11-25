package extrun

import (
	"context"
	"os"
	"syscall"

	"github.com/zcong1993/kit/pkg/log"

	"github.com/oklog/run"
)

// HandleSignal add signals interrupter to run group
// default signals are SIGINT and SIGTERM.
func HandleSignal(g *run.Group, signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	e, c := run.SignalHandler(context.Background(), signals...)
	g.Add(e, c)
}

// RunUntilExit is helper function for long term tasks
// if run group is invoked, the wrapper fn should receive <- ctx.Done()
func RunUntilExit(g *run.Group, logger *log.Logger, fn func(ctx context.Context) error, componentName string) {
	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		return fn(ctx)
	}, func(err error) {
		logger.With(log.Component(componentName)).Info("shutting down", log.ErrorMsg(err))
		cancel()
	})
}
