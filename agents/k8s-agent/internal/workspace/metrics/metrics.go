/*
Package metrics surfaces Workspace level Prometheus metrics.
*/
package metrics

import (
	"sync"

	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	workspacePhase = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: "aegis",
		Subsystem: "workspace",
		Name:      "phase",
		Help:      "Current workspace phase by project and queue",
	}, []string{"project", "queue", "phase"})

	registerMetrics sync.Once
)

// Collector exposes Workspace metrics helpers.
type Collector struct{}

// NewCollector returns a Collector instance and registers shared metrics.
func NewCollector() *Collector {
	registerMetrics.Do(func() {
		prom.MustRegister(workspacePhase)
	})
	return &Collector{}
}

// ObservePhase sets the phase gauge for the provided project/queue/phase combination.
func (c *Collector) ObservePhase(project, queue, phase string) {
	if c == nil {
		return
	}
	workspacePhase.WithLabelValues(project, queue, phase).Set(1)
}
