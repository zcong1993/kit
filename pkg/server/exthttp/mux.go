package exthttp

import (
	"net/http"
	"net/http/pprof"

	"github.com/zcong1993/kit/pkg/log"

	"github.com/zcong1993/kit/pkg/prober"

	"github.com/felixge/fgprof"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MuxServer 简单的 http mux server
// 一般用来挂载 prometheus 和 pprof.
// MuxServer is a simple http mux server.
type MuxServer struct {
	*HttpServer
	mux *http.ServeMux
}

// NewMuxServer crate a new MuxServer instance.
func NewMuxServer(logger *log.Logger, opts ...OptionFunc) *MuxServer {
	mux := http.NewServeMux()
	return &MuxServer{
		HttpServer: NewHttpServer(mux, logger, opts...),
		mux:        mux,
	}
}

// HandleFunc add handle function to mux server.
func (ms *MuxServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	ms.mux.HandleFunc(pattern, handler)
}

// Handle add handler to mux server.
func (ms *MuxServer) Handle(pattern string, handler http.Handler) {
	ms.mux.Handle(pattern, handler)
}

// RegisterProfiler register pprof routes.
func (ms *MuxServer) RegisterProfiler() {
	ms.logger.Info("register profiler")
	ms.mux.HandleFunc("/debug/pprof/", pprof.Index)
	ms.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ms.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ms.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ms.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	ms.mux.Handle("/debug/fgprof", fgprof.Handler())
}

// RegisterMetrics register metrics route.
func (ms *MuxServer) RegisterMetrics(g prometheus.Gatherer) {
	if g != nil {
		ms.logger.Info("register prometheus gatherer")
		ms.mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	} else {
		ms.logger.Info("register prometheus default gatherer")
		ms.mux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	}
}

// RegisterProber register liveness(healthy) and readiness(ready) routes.
func (ms *MuxServer) RegisterProber(p *prober.HTTPProbe) {
	if p != nil {
		ms.mux.Handle("/-/healthy", p.HealthyHandler(ms.logger))
		ms.mux.Handle("/-/ready", p.ReadyHandler(ms.logger))
	}
}

// RegisterLogControl register log level control routes.
func (ms *MuxServer) RegisterLogControl(handler http.Handler) {
	ms.mux.Handle("/log/level", handler)
}

// RunGroup start server with run group.
func (ms *MuxServer) RunGroup(g *run.Group) {
	g.Add(func() error {
		return ms.Start()
	}, func(err error) {
		ms.Shutdown(err)
	})
}
