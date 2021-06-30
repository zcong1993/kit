package extgrpc

import (
	"context"
	"net"

	"github.com/zcong1993/kit/pkg/log"
	"go.uber.org/zap"

	"github.com/oklog/run"

	"github.com/pkg/errors"
	"github.com/zcong1993/kit/pkg/prober"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	logger *log.Logger

	srv      *grpc.Server
	listener net.Listener

	opts Options
}

func NewServer(logger *log.Logger, probe *prober.GRPCProbe, opts ...Option) *Server {
	logger = logger.With(log.Service("gRPC/server"))
	options := Options{
		network: "tcp",
	}

	for _, o := range opts {
		o(&options)
	}

	if options.tlsConfig != nil {
		options.grpcOpts = append(options.grpcOpts, grpc.Creds(credentials.NewTLS(options.tlsConfig)))
	}

	if len(options.grpcUnaryServerInterceptors) > 0 {
		options.grpcOpts = append(options.grpcOpts, grpc.ChainUnaryInterceptor(options.grpcUnaryServerInterceptors...))
	}

	if len(options.grpcStreamServerInterceptors) > 0 {
		options.grpcOpts = append(options.grpcOpts, grpc.ChainStreamInterceptor(options.grpcStreamServerInterceptors...))
	}

	s := grpc.NewServer(options.grpcOpts...)

	// Register all configured servers.
	for _, f := range options.registerServerFuncs {
		f(s)
	}

	if probe != nil {
		grpc_health.RegisterHealthServer(s, probe.HealthServer())
	}

	reflection.Register(s)

	return &Server{
		logger: logger,
		srv:    s,
		opts:   options,
	}
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen(s.opts.network, s.opts.listen)
	if err != nil {
		return errors.Wrapf(err, "listen gRPC on address %s", s.opts.listen)
	}
	s.listener = l
	s.logger.Info("listening for serving gRPC", zap.String("address", s.opts.listen))
	return errors.Wrap(s.srv.Serve(s.listener), "serve gRPC")
}

// Shutdown gracefully shuts down the server by waiting,
// for specified amount of time (by gracePeriod) for connections to return to idle and then shut down.
func (s *Server) Shutdown(err error) {
	s.logger.Info("internal server is shutting down", log.ErrorMsg(err))

	if s.opts.gracePeriod == 0 {
		s.srv.Stop()
		s.logger.Info("internal server is shutdown", log.ErrorMsg(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.opts.gracePeriod)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		s.logger.Info("gracefully stopping internal server")
		s.srv.GracefulStop() // Also closes s.listener.
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("grace period exceeded enforcing shutdown")
		s.srv.Stop()
		return
	case <-stopped:
		cancel()
	}
	s.logger.Info("internal server is shutdown gracefully", log.ErrorMsg(err))
}

func (s *Server) Run(g *run.Group, statusProber prober.Probe) {
	g.Add(func() error {
		statusProber.Healthy()
		return s.ListenAndServe()
	}, func(err error) {
		statusProber.NotReady(err)
		s.Shutdown(err)
		statusProber.NotHealthy(err)
	})
}
