package exthttp

import (
	"net/http"
	"net/http/pprof"

	"github.com/zcong1993/x/pkg/prober"

	"github.com/go-kit/kit/log/level"

	"github.com/felixge/fgprof"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MuxServer 简单的 http mux server
// 一般用来挂载 prometheus 和 pprof
type MuxServer struct {
	*HttpServer
	mux *http.ServeMux
}

func NewMuxServer(logger log.Logger, opts ...OptionFunc) *MuxServer {
	mux := http.NewServeMux()
	return &MuxServer{
		HttpServer: NewHttpServer(mux, logger, opts...),
		mux:        mux,
	}
}

func (ms *MuxServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	ms.mux.HandleFunc(pattern, handler)
}

func (ms *MuxServer) Handle(pattern string, handler http.Handler) {
	ms.mux.Handle(pattern, handler)
}

func (ms *MuxServer) RegisterProfiler() {
	level.Info(ms.logger).Log("msg", "register profiler")
	ms.mux.HandleFunc("/debug/pprof/", pprof.Index)
	ms.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ms.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ms.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ms.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	ms.mux.Handle("/debug/fgprof", fgprof.Handler())
}

func (ms *MuxServer) RegisterMetrics(g prometheus.Gatherer) {
	if g != nil {
		level.Info(ms.logger).Log("msg", "register prometheus gatherer")
		ms.mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	} else {
		level.Info(ms.logger).Log("msg", "register prometheus default gatherer")
		ms.mux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	}
}

func (ms *MuxServer) RegisterProber(p *prober.HTTPProbe) {
	if p != nil {
		ms.mux.Handle("/-/healthy", p.HealthyHandler(ms.logger))
		ms.mux.Handle("/-/ready", p.ReadyHandler(ms.logger))
	}
}

func (ms *MuxServer) RunGroup(g *run.Group) {
	g.Add(func() error {
		return ms.Start()
	}, func(err error) {
		ms.Shutdown(err)
	})
}
