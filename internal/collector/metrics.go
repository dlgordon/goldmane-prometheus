package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the Prometheus metrics for Calico flow data
type Metrics struct {
	FlowAllow *prometheus.CounterVec
	FlowDeny  *prometheus.CounterVec
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		FlowAllow: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calico_flow_allow",
				Help: "Number of allowed network flows in Calico",
			},
			[]string{
				"reporter",
				"protocol",
				"src_namespace",
				"src_pod",
				"src_port",
				"dst_namespace",
				"dst_object",
				"dst_port",
			},
		),
		FlowDeny: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calico_flow_deny",
				Help: "Number of denied network flows in Calico",
			},
			[]string{
				"reporter",
				"protocol",
				"src_namespace",
				"src_pod",
				"src_port",
				"dst_namespace",
				"dst_object",
				"dst_port",
			},
		),
	}
}
