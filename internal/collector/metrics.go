package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the Prometheus metrics for Calico flow data
type Metrics struct {
	FlowAllow *prometheus.CounterVec
	FlowDeny  *prometheus.CounterVec

	// Operational metrics for monitoring the exporter itself
	APIRequests         *prometheus.CounterVec
	APILastSuccessTime  prometheus.Gauge
	APIFlowsProcessed   prometheus.Counter
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	flowLabels := []string{
		"reporter",
		"protocol",
		"src_namespace",
		"src_pod",
		"src_port",
		"dst_namespace",
		"dst_object",
		"dst_port",
	}

	return &Metrics{
		FlowAllow: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calico_flow_allow",
				Help: "Number of allowed network flows in Calico",
			},
			flowLabels,
		),
		FlowDeny: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "calico_flow_deny",
				Help: "Number of denied network flows in Calico",
			},
			flowLabels,
		),
		APIRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "goldmane_api_requests_total",
				Help: "Total number of Goldmane API poll requests",
			},
			[]string{"result"},
		),
		APILastSuccessTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goldmane_api_last_success_timestamp_seconds",
				Help: "Unix timestamp of the last successful Goldmane API poll",
			},
		),
		APIFlowsProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "goldmane_api_flows_processed_total",
				Help: "Total number of flows processed from the Goldmane API",
			},
		),
	}
}
