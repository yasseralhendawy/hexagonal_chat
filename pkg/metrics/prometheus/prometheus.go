package prometrics

import (
	"github.com/prometheus/client_golang/prometheus"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type prometheusMetrics struct {
	reg prometheus.Registerer

	calls   *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func CreateMetrics() (appmetrics.Metrics, error) {
	metric := &prometheusMetrics{
		reg: prometheus.NewRegistry(),
		calls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_calls_total",
				Help: "Number of database calls",
			}, []string{"type_name", "operation_name", "status"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_time",
				Help:    "Duration of HTTP requests",
				Buckets: []float64{1, 2, 5, 10, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
			}, []string{"path", "method", "status_code"}),
	}
	err := metric.register()
	if err != nil {
		return nil, err
	}
	return metric, nil
}

// CallsWithLabelValues implements Metrics.
func (metric *prometheusMetrics) DBCallsWithLabelValues(lvs ...string) {
	metric.calls.WithLabelValues(lvs...).Inc()
}

// LatencyWithLabelValues implements Metrics.
func (metric *prometheusMetrics) LatencyWithLabelValues(duration float64, lvs ...string) {
	metric.latency.WithLabelValues(lvs...).Observe(duration)
}

func (metric *prometheusMetrics) register() error {
	err := metric.reg.Register(metric.calls)
	if err != nil {
		return err
	}
	err = metric.reg.Register(metric.latency)
	if err != nil {
		return err
	}
	return nil
}
