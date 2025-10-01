package metrics

import (
	"time"
)

type PerformanceMetric struct {
	ID          int        `db:"id" json:"id"`
	BatchJobID  int        `db:"batch_job_id" json:"batch_job_id"`
	MetricType  string     `db:"metric_type" json:"metric_type"`
	MetricValue *float64   `db:"metric_value" json:"metric_value"`
	Unit        *string    `db:"unit" json:"unit"`
	MeasuredAt  *time.Time `db:"measured_at" json:"measured_at"`
}
