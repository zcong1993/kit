package exthttp

import (
	"context"
	"net/http"
	"time"

	"github.com/zcong1993/x/pkg/log"
	"go.uber.org/zap"

	"github.com/oklog/run"
	"github.com/zcong1993/x/pkg/prober"
)

type HttpServer struct {
	handler http.Handler
	srv     *http.Server
	logger  *log.Logger
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

func NewHttpServer(handler http.Handler, logger *log.Logger, opts ...OptionFunc) *HttpServer {
	option := &Option{
		service: "http/server",
	}
	for _, f := range opts {
		f(option)
	}
	return &HttpServer{handler: handler, logger: logger.With(log.Service(option.service)), option: option}
}

func (hs *HttpServer) Start() error {
	srv := &http.Server{
		Addr:    hs.option.listen,
		Handler: hs.handler,
	}
	hs.srv = srv
	hs.logger.Sugar().Infow("listening for requests", "address", hs.option.listen)
	return srv.ListenAndServe()
}

func (hs *HttpServer) Shutdown(err error) {
	hs.logger.Info("internal server is shutting down", log.ErrorMsg(err))
	if err == http.ErrServerClosed {
		hs.logger.Warn("internal server closed unexpectedly")
		return
	}

	if hs.option.gracePeriod == 0 {
		hs.srv.Close()
		hs.logger.Info("internal server is shutdown", log.ErrorMsg(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), hs.option.gracePeriod)
	defer cancel()

	if err := hs.srv.Shutdown(ctx); err != nil {
		hs.logger.Error("internal server shut down failed", zap.Error(err))
		return
	}

	hs.logger.Info("internal server is shutdown gracefully", log.ErrorMsg(err))
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
