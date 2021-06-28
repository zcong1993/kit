package register

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zcong1993/x/pkg/log"

	"github.com/zcong1993/x/pkg/tracing/client"

	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zcong1993/x/pkg/extflag"
	"github.com/zcong1993/x/pkg/runutil"
)

const (
	tracingConfig = "tracing.config"
	helpText      = "YAML file with tracing configuration"
)

type NoopCloser struct{}

func (nc *NoopCloser) Close() error {
	return nil
}

func RegisterFlags(flagSet *pflag.FlagSet) {
	extflag.RegisterPathOrContent(flagSet, tracingConfig, helpText)
}

func NewTracer(cmd *cobra.Command, logger *log.Logger, metrics prometheus.Registerer) (opentracing.Tracer, io.Closer, error) {
	confContentYaml, err := extflag.LoadContent(cmd, tracingConfig, false)
	if err != nil {
		return nil, nil, err
	}

	if len(confContentYaml) == 0 {
		logger.Info("Tracing will be disabled")
		return client.NoopTracer(), &NoopCloser{}, nil
	}

	return client.NewTracer(context.Background(), logger, metrics, confContentYaml)
}

func MustInitTracer(g *run.Group, cmd *cobra.Command, logger *log.Logger, metrics prometheus.Registerer) opentracing.Tracer {
	tracer, closer, err := NewTracer(cmd, logger, metrics)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "init tracing failed"))
		os.Exit(1)
	}

	opentracing.SetGlobalTracer(tracer)

	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(error) {
		runutil.CloseWithLogOnErr(logger, closer, "tracer")
		cancel()
	})

	return tracer
}
