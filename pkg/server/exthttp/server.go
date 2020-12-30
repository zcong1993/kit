package exthttp

import (
	"context"
	"net/http"
	"time"

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
	GracePeriod time.Duration
	Listen      string
}

func NewHttpServer(handler http.Handler, logger log.Logger, option *Option) *HttpServer {
	return &HttpServer{handler: handler, logger: log.With(logger, "service", "http/server"), option: option}
}

func (hs *HttpServer) Start() error {
	srv := &http.Server{
		Addr:    hs.option.Listen,
		Handler: hs.handler,
	}
	hs.srv = srv
	level.Info(hs.logger).Log("msg", "listening for requests", "address", hs.option.Listen)
	return srv.ListenAndServe()
}

func (hs *HttpServer) Shutdown(err error) {
	level.Info(hs.logger).Log("msg", "internal server is shutting down", "err", err)
	if err == http.ErrServerClosed {
		level.Warn(hs.logger).Log("msg", "internal server closed unexpectedly")
		return
	}

	if hs.option.GracePeriod == 0 {
		hs.srv.Close()
		level.Info(hs.logger).Log("msg", "internal server is shutdown", "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), hs.option.GracePeriod)
	defer cancel()

	if err := hs.srv.Shutdown(ctx); err != nil {
		level.Error(hs.logger).Log("msg", "internal server shut down failed", "err", err)
		return
	}
	level.Info(hs.logger).Log("msg", "internal server is shutdown gracefully", "err", err)
}
