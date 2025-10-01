package metrics

import "time"

type BatchJob struct {
	ID                int        `db:"id" json:"id"`
	JobName           string     `db:"job_name" json:"job_name"`
	ToolUsed          string     `db:"tool_used" json:"tool_used"`
	StartedAt         time.Time  `db:"started_at" json:"started_at"`
	CompletedAt       *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	TotalRecords      *int       `db:"total_records" json:"total_records,omitempty"`
	SuccessfulRecords *int       `db:"successful_records" json:"successful_records,omitempty"`
	FailedRecords     int        `db:"failed_records" json:"failed_records"`
	Status            string     `db:"status" json:"status"`
	FileSizeMB        *float64   `db:"file_size_mb" json:"file_size_mb,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}
