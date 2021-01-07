package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 10},
		},
		[]string{"code", "handler", "method", "path"},
	)
	requestSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"code", "handler", "method", "path"},
	)
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"code", "handler", "method", "path"},
	)
	responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"code", "handler", "method", "path"},
	)
)

func NewInstrumentationMiddleware(reg prometheus.Registerer) gin.HandlerFunc {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	reg.MustRegister(requestDuration, requestSize, requestsTotal, responseSize)

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		labels := prometheus.Labels{
			"code":    strconv.Itoa(c.Writer.Status()),
			"handler": c.HandlerName(),
			"method":  c.Request.Method,
			"path":    c.FullPath(),
		}

		requestDuration.With(labels).Observe(time.Since(start).Seconds())
		requestsTotal.With(labels).Inc()
		reqSize := computeApproximateRequestSize(c.Request)
		requestSize.With(labels).Observe(float64(reqSize))
		responseSize.With(labels).Observe(float64(c.Writer.Size()))
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s += len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
