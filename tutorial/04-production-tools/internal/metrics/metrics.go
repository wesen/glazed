package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	CommandExecutions *prometheus.CounterVec
	CommandDuration   *prometheus.HistogramVec
	ErrorsTotal       *prometheus.CounterVec
	FilesProcessed    prometheus.Counter
	BytesProcessed    prometheus.Counter
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		CommandExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "logprocessor_command_executions_total",
				Help: "Total number of command executions",
			},
			[]string{"command", "status"},
		),
		CommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "logprocessor_command_duration_seconds",
				Help:    "Duration of command executions",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"command"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "logprocessor_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "code"},
		),
		FilesProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "logprocessor_files_processed_total",
				Help: "Total number of files processed",
			},
		),
		BytesProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "logprocessor_bytes_processed_total",
				Help: "Total number of bytes processed",
			},
		),
	}
}

// Default metrics instance
var defaultMetrics *Metrics

// InitDefaultMetrics initializes the default metrics
func InitDefaultMetrics() {
	defaultMetrics = NewMetrics()
}

// GetDefaultMetrics returns the default metrics instance
func GetDefaultMetrics() *Metrics {
	if defaultMetrics == nil {
		InitDefaultMetrics()
	}
	return defaultMetrics
}
