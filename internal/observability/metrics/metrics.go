package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Label constants для избежания ошибок в runtime
const (
	labelMethod    = "method"
	labelPath      = "path"
	labelStatus    = "status"
	labelResult    = "result"
	labelOperation = "operation"
)

// Metrics holds all prometheus metric collectors for the application.
// Uses modern prometheus patterns with promauto for automatic registration.
type Metrics struct {
	// HTTP metrics
	HTTPRequestTotal     *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Trip create metrics
	TripCreateTotal    *prometheus.CounterVec
	TripCreateDuration *prometheus.HistogramVec

	// Trip get by id metrics
	TripGetByIDTotal    *prometheus.CounterVec
	TripGetByIDDuration *prometheus.HistogramVec

	// Trip publish metrics
	TripDraftToPublishTotal    *prometheus.CounterVec
	TripDraftToPublishDuration *prometheus.HistogramVec

	// Repository metrics
	RepositoryQueryTotal    *prometheus.CounterVec
	RepositoryQueryDuration *prometheus.HistogramVec
}

// New creates and registers all metrics using the provided registry.
// Uses promauto for automatic registration to reduce boilerplate and deprecation warnings.
func New(reg prometheus.Registerer) *Metrics {
	// Set default registry for promauto
	factory := promauto.With(reg)

	m := &Metrics{
		// HTTP metrics
		HTTPRequestTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sharetrip",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed",
			},
			[]string{labelMethod, labelPath, labelStatus},
		),

		HTTPRequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sharetrip",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{labelMethod, labelPath, labelStatus},
		),

		HTTPRequestsInFlight: factory.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "sharetrip",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),

		// Trip create metrics
		TripCreateTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "create_total",
				Help:      "Total number of trip creation attempts",
			},
			[]string{labelResult},
		),

		TripCreateDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "create_duration_seconds",
				Help:      "Trip creation operation latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{labelResult},
		),

		// Trip get by id metrics
		TripGetByIDTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "get_by_id_total",
				Help:      "Total number of get trip by ID attempts",
			},
			[]string{labelResult},
		),

		TripGetByIDDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "get_by_id_duration_seconds",
				Help:      "Get trip by ID operation latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{labelResult},
		),

		// Trip publish metrics
		TripDraftToPublishTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "publish_total",
				Help:      "Total number of trip publication attempts",
			},
			[]string{labelResult},
		),

		TripDraftToPublishDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sharetrip",
				Subsystem: "trip",
				Name:      "publish_duration_seconds",
				Help:      "Trip publication operation latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{labelResult},
		),

		// Repository metrics
		RepositoryQueryTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sharetrip",
				Subsystem: "repository",
				Name:      "query_total",
				Help:      "Total number of database queries",
			},
			[]string{labelOperation, labelResult},
		),

		RepositoryQueryDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sharetrip",
				Subsystem: "repository",
				Name:      "query_duration_seconds",
				Help:      "Database query latency in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{labelOperation, labelResult},
		),
	}

	// Register standard Go and process collectors
	// Note: GoCollector and ProcessCollector are registered separately as they're not factory methods
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		// prometheus.NewGoCollector(),
		//prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	return m
}
