package extrun

import (
	"context"
	"os"
	"syscall"

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
