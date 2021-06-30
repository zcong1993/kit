// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package prober

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zcong1993/kit/pkg/log"
	"go.uber.org/zap"
)

const (
	ready   = "ready"
	healthy = "healthy"
)

// InstrumentationProbe stores instrumentation state of Probe.
// This is created with an intention to combine with other Probe's using prober.Combine.
type InstrumentationProbe struct {
	component string
	logger    *log.Logger

	status *prometheus.GaugeVec
}

// NewInstrumentation returns InstrumentationProbe records readiness and healthiness for given component.
func NewInstrumentation(component string, logger *log.Logger, reg prometheus.Registerer) *InstrumentationProbe {
	p := InstrumentationProbe{
		component: component,
		logger:    logger.With(log.Component(component)),
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "status",
			Help:        "Represents status (0 indicates failure, 1 indicates success) of the component.",
			ConstLabels: map[string]string{"component": component},
		},
			[]string{"check"},
		),
	}
	reg.MustRegister(p.status)
	return &p
}

// Ready records the component status when Ready is called, if combined with other Probes.
func (p *InstrumentationProbe) Ready() {
	p.status.WithLabelValues(ready).Set(1)
	p.logger.Info("changing probe status", zap.String("status", "ready"))
}

// NotReady records the component status when NotReady is called, if combined with other Probes.
func (p *InstrumentationProbe) NotReady(err error) {
	p.status.WithLabelValues(ready).Set(0)
	p.logger.Warn("changing probe status", zap.String("status", "not-ready"), log.ErrorMsg(err))
}

// Healthy records the component status when Healthy is called, if combined with other Probes.
func (p *InstrumentationProbe) Healthy() {
	p.status.WithLabelValues(healthy).Set(1)
	p.logger.Info("changing probe status", zap.String("status", "healthy"))
}

// NotHealthy records the component status when NotHealthy is called, if combined with other Probes.
func (p *InstrumentationProbe) NotHealthy(err error) {
	p.status.WithLabelValues(healthy).Set(0)
	p.logger.Warn("changing probe status", zap.String("status", "not-healthy"), log.ErrorMsg(err))
}
