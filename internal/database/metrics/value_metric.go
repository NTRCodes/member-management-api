package metrics

import (
	"time"
)

type ValueMetric struct {
	ID                        int        `db:"id" json:"id"`
	BatchJobID                int        `db:"batch_job_id" json:"batch_job_id"`
	ManualTimeEstimateMinutes *int       `db:"manual_time_estimate_minutes" json:"manual_time_estimate_minutes"`
	ActualProcessingMinutes   *float64   `db:"actual_processing_minutes" json:"actual_processing_minutes"`
	TimeSavedMinutes          *float64   `db:"time_saved_minutes" json:"time_saved_minutes"`
	HourlyRate                *float64   `db:"hourly_rate" json:"hourly_rate"`
	CostSavings               *float64   `db:"cost_savings" json:"cost_savings"`
	ErrorReductionPercentage  *float64   `db:"error_reduction_percent" json:"error_reduction_percent"`
	CalculatedAT              *time.Time `db:"calculated_at" json:"calculated_at"`
}
