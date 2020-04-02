package querier

import (
	"bytes"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestMetaFetcherMetrics(t *testing.T) {
	mainReg := prometheus.NewPedanticRegistry()

	metrics := newMetaFetcherMetrics()
	mainReg.MustRegister(metrics)

	metrics.addUserRegistry("user1", populateMetaFetcherMetrics(3))
	metrics.addUserRegistry("user2", populateMetaFetcherMetrics(5))
	metrics.addUserRegistry("user3", populateMetaFetcherMetrics(7))

	//noinspection ALL
	err := testutil.GatherAndCompare(mainReg, bytes.NewBufferString(`
		# HELP cortex_querier_blocks_meta_sync_duration_seconds Duration of the blocks metadata synchronization in seconds
		# TYPE cortex_querier_blocks_meta_sync_duration_seconds histogram
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="0.01"} 0
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="1"} 0
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="10"} 3
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="100"} 3
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="1000"} 3
		cortex_querier_blocks_meta_sync_duration_seconds_bucket{le="+Inf"} 3
		cortex_querier_blocks_meta_sync_duration_seconds_sum 9
		cortex_querier_blocks_meta_sync_duration_seconds_count 3

		# HELP cortex_querier_blocks_meta_sync_failures_total Total blocks metadata synchronization failures
		# TYPE cortex_querier_blocks_meta_sync_failures_total counter
		cortex_querier_blocks_meta_sync_failures_total 30

		# HELP cortex_querier_blocks_meta_syncs_total Total blocks metadata synchronization attempts
		# TYPE cortex_querier_blocks_meta_syncs_total counter
		cortex_querier_blocks_meta_syncs_total 15

		# HELP cortex_querier_blocks_meta_sync_consistency_delay_seconds Configured consistency delay in seconds.
		# TYPE cortex_querier_blocks_meta_sync_consistency_delay_seconds gauge
		cortex_querier_blocks_meta_sync_consistency_delay_seconds 300
`))
	require.NoError(t, err)
}

func populateMetaFetcherMetrics(base float64) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	m := newMetaFetcherMetricsMock(reg)

	m.syncs.Add(base * 1)
	m.syncFailures.Add(base * 2)
	m.syncDuration.Observe(3)
	m.syncConsistencyDelay.Set(300)

	return reg
}

type metaFetcherMetricsMock struct {
	syncs                prometheus.Counter
	syncFailures         prometheus.Counter
	syncDuration         prometheus.Histogram
	syncConsistencyDelay prometheus.Gauge
}

func newMetaFetcherMetricsMock(reg prometheus.Registerer) *metaFetcherMetricsMock {
	var m metaFetcherMetricsMock

	m.syncs = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Subsystem: "blocks_meta",
		Name:      "syncs_total",
		Help:      "Total blocks metadata synchronization attempts",
	})
	m.syncFailures = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Subsystem: "blocks_meta",
		Name:      "sync_failures_total",
		Help:      "Total blocks metadata synchronization failures",
	})
	m.syncDuration = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
		Subsystem: "blocks_meta",
		Name:      "sync_duration_seconds",
		Help:      "Duration of the blocks metadata synchronization in seconds",
		Buckets:   []float64{0.01, 1, 10, 100, 1000},
	})
	m.syncConsistencyDelay = promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "consistency_delay_seconds",
		Help: "Configured consistency delay in seconds.",
	})

	return &m
}