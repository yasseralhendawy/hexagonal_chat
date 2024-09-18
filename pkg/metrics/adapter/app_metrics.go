package appmetrics

type Metrics interface {
	LatencyWithLabelValues(duration float64, lvs ...string)
	DBCallsWithLabelValues(lvs ...string)
}
