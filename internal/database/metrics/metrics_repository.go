package metrics

import (
	"NTRCodes/member-api/internal/database"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PerformanceMetric and ValueMetric structs...

// MetricsRepository can write, get, and update a record.
type MetricsRepository interface {
	PostBatchJob(ctx context.Context, job *BatchJob) error
	GetBatchJob(ctx context.Context, id int) (*BatchJob, error)
	UpdateBatchJob(ctx context.Context, job *BatchJob) error
	PostPerformanceMetric(ctx context.Context, metric PerformanceMetric) error
	PostValueMetric(ctx context.Context, metric ValueMetric) error
}

type metricsRepository struct {
	db *database.DB
}

func (r *metricsRepository) PostPerformanceMetric(ctx context.Context, metric PerformanceMetric) error {
	query := `
		INSERT INTO performance_metrics
		(batch_job_id,
		 metric_type,
		 metric_value,
		 unit,
		 measured_at) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, measured_at`

	err := r.db.QueryRowContext(ctx, query,
		metric.BatchJobID,
		metric.MetricType,
		metric.MetricValue,
		metric.Unit,
		metric.MeasuredAt).Scan(&metric.ID, &metric.MeasuredAt)

	if err != nil {
		return fmt.Errorf("failed to post performance metric:  %w", err)
	}

	return nil
}

func (r *metricsRepository) PostValueMetric(ctx context.Context, metric ValueMetric) error {
	query := `
		INSERT INTO value_metrics
			(batch_job_id,
			 manual_time_estimate_minutes,
			 actual_processing_minutes,
			 time_saved_minutes,
			 hourly_rate,
			 cost_savings,
			 error_reduction_percent,
			 calculated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id, calculated_at`

	err := r.db.QueryRowContext(ctx, query,
		metric.BatchJobID,
		metric.ManualTimeEstimateMinutes,
		metric.ActualProcessingMinutes,
		metric.TimeSavedMinutes,
		metric.HourlyRate,
		metric.CostSavings,
		metric.ErrorReductionPercentage,
		metric.CalculatedAT).Scan(&metric.ID, &metric.CalculatedAT)

	if err != nil {
		return fmt.Errorf("failed to post value metric:  %w", err)
	}

	return nil
}

func NewMetricsRepository(db *database.DB) MetricsRepository {
	return &metricsRepository{db: db}
}

func (r *metricsRepository) PostBatchJob(ctx context.Context, job *BatchJob) error {
	query := `
        INSERT INTO batch_jobs (
            job_name, tool_used, started_at, completed_at, total_records,
            successful_records, failed_records, status, file_size_mb
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		job.JobName,
		job.ToolUsed,
		job.StartedAt,
		job.CompletedAt,
		job.TotalRecords,
		job.SuccessfulRecords,
		job.FailedRecords,
		job.Status,
		job.FileSizeMB,
	).Scan(&job.ID, &job.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to post batch job: %w", err)
	}

	return nil
}

func (r *metricsRepository) GetBatchJob(ctx context.Context, id int) (*BatchJob, error) {
	query := `SELECT job_name, tool_used, started_at, completed_at, total_records,
            successful_records, failed_records, status, file_size_mb FROM batch_jobs where id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	var job BatchJob
	err := row.Scan(
		&job.JobName,
		&job.ToolUsed,
		&job.StartedAt,
		&job.CompletedAt,
		&job.TotalRecords,
		&job.SuccessfulRecords,
		&job.FailedRecords,
		&job.Status,
		&job.FileSizeMB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get batch job: %v", err)
	}

	return &job, nil
}

func (r *metricsRepository) UpdateBatchJob(ctx context.Context, job *BatchJob) error {
	// check if job exists; if not return the error
	_, err := r.GetBatchJob(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("Job with id: %v  - does not exist", job.ID)
	}

	query := `
		UPDATE batch_jobs
		set job_name = $1, 
		    tool_used = $2, 
		    started_at = $3, 
		    completed_at = $4, 
		    total_records = $5,
            successful_records = $6, 
            failed_records = $7, 
            status = $8, 
            file_size_mb = $9 
		where id = $10`

	// find job in the database
	newRow, queryError := r.db.ExecContext(ctx, query,
		job.JobName,
		job.ToolUsed,
		job.StartedAt,
		job.CompletedAt,
		job.TotalRecords,
		job.SuccessfulRecords,
		job.FailedRecords,
		job.Status,
		job.FileSizeMB,
		job.ID)
	if queryError != nil {
		return fmt.Errorf("Job with: %v  - could not be updated", job.ID)
	}

	rowsAffected, err := newRow.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated for job id: %v", job.ID)
	}

	return nil
}
