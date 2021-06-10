package exthttp

import (
	"context"
	"net/http"
	"time"

	"github.com/oklog/run"
	"github.com/zcong1993/x/pkg/prober"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type HttpServer struct {
	handler http.Handler
	srv     *http.Server
	logger  log.Logger
	option  *Option
}

type Option struct {
	gracePeriod time.Duration
	listen      string
	service     string
}

type OptionFunc func(option *Option)

func WithGracePeriod(gracePeriod time.Duration) OptionFunc {
	return func(o *Option) {
		o.gracePeriod = gracePeriod
	}
}

func WithListen(listen string) OptionFunc {
	return func(o *Option) {
		o.listen = listen
	}
}

func WithServiceName(name string) OptionFunc {
	return func(o *Option) {
		o.service = name
	}
}

func NewHttpServer(handler http.Handler, logger log.Logger, opts ...OptionFunc) *HttpServer {
	option := &Option{
		service: "http/server",
	}
	for _, f := range opts {
		f(option)
	}
	return &HttpServer{handler: handler, logger: log.With(logger, "service", option.service), option: option}
}

func (hs *HttpServer) Start() error {
	srv := &http.Server{
		Addr:    hs.option.listen,
		Handler: hs.handler,
	}
	hs.srv = srv
	level.Info(hs.logger).Log("msg", "listening for requests", "address", hs.option.listen)
	return srv.ListenAndServe()
}

func (hs *HttpServer) Shutdown(err error) {
	level.Info(hs.logger).Log("msg", "internal server is shutting down", "err", err)
	if err == http.ErrServerClosed {
		level.Warn(hs.logger).Log("msg", "internal server closed unexpectedly")
		return
	}

	if hs.option.gracePeriod == 0 {
		hs.srv.Close()
		level.Info(hs.logger).Log("msg", "internal server is shutdown", "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), hs.option.gracePeriod)
	defer cancel()

	if err := hs.srv.Shutdown(ctx); err != nil {
		level.Error(hs.logger).Log("msg", "internal server shut down failed", "err", err)
		return
	}
	level.Info(hs.logger).Log("msg", "internal server is shutdown gracefully", "err", err)
}

func (hs *HttpServer) Run(g *run.Group, statusProber prober.Probe) {
	g.Add(func() error {
		statusProber.Healthy()
		return hs.Start()
	}, func(err error) {
		statusProber.NotReady(err)
		hs.Shutdown(err)
		statusProber.NotHealthy(err)
	})
}
