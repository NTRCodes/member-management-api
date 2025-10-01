package httpx

import (
	"NTRCodes/member-api/internal/app"
	"NTRCodes/member-api/internal/database/metrics"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreatePerformanceMetricRequest represents the request body for creating a performance metric
type CreatePerformanceMetricRequest struct {
	BatchJobID  int     `json:"batch_job_id"`
	MetricType  string  `json:"metric_type"`
	MetricValue float64 `json:"metric_value"`
	Unit        string  `json:"unit"`
}

// CreatePerformanceMetricResponse represents the response after creating a performance metric
type CreatePerformanceMetricResponse struct {
	ID          int       `json:"id"`
	BatchJobID  int       `json:"batch_job_id"`
	MetricType  string    `json:"metric_type"`
	MetricValue float64   `json:"metric_value"`
	Unit        string    `json:"unit"`
	MeasuredAt  time.Time `json:"measured_at"`
}

// CreateValueMetricRequest represents the request body for creating a value metric
type CreateValueMetricRequest struct {
	BatchJobID                int     `json:"batch_job_id"`
	ManualTimeEstimateMinutes int     `json:"manual_time_estimate_minutes"`
	ActualProcessingMinutes   float64 `json:"actual_processing_minutes"`
	TimeSavedMinutes          float64 `json:"time_saved_minutes"`
	HourlyRate                float64 `json:"hourly_rate"`
	CostSavings               float64 `json:"cost_savings"`
	ErrorReductionPercentage  float64 `json:"error_reduction_percentage"`
}

// CreateValueMetricResponse represents the response after creating a value metric
type CreateValueMetricResponse struct {
	ID                        int       `json:"id"`
	BatchJobID                int       `json:"batch_job_id"`
	ManualTimeEstimateMinutes int       `json:"manual_time_estimate_minutes"`
	ActualProcessingMinutes   float64   `json:"actual_processing_minutes"`
	TimeSavedMinutes          float64   `json:"time_saved_minutes"`
	HourlyRate                float64   `json:"hourly_rate"`
	CostSavings               float64   `json:"cost_savings"`
	ErrorReductionPercentage  float64   `json:"error_reduction_percentage"`
	CalculatedAt              time.Time `json:"calculated_at"`
}

// CreateBatchJobRequest represents the request body for creating a batch job
type CreateBatchJobRequest struct {
	JobName           string     `json:"job_name"`
	ToolUsed          string     `json:"tool_used"`
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	TotalRecords      *int       `json:"total_records,omitempty"`
	SuccessfulRecords *int       `json:"successful_records,omitempty"`
	FailedRecords     int        `json:"failed_records"`
	Status            string     `json:"status"`
	FileSizeMB        *float64   `json:"file_size_mb,omitempty"`
}

// CreateBatchJobResponse represents the response after creating a batch job
type CreateBatchJobResponse struct {
	ID                int        `json:"id"`
	JobName           string     `json:"job_name"`
	ToolUsed          string     `json:"tool_used"`
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	TotalRecords      *int       `json:"total_records,omitempty"`
	SuccessfulRecords *int       `json:"successful_records,omitempty"`
	FailedRecords     int        `json:"failed_records"`
	Status            string     `json:"status"`
	FileSizeMB        *float64   `json:"file_size_mb,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// GetBatchJobResponse represents the response for getting a batch job
type GetBatchJobResponse struct {
	ID                int        `json:"id"`
	JobName           string     `json:"job_name"`
	ToolUsed          string     `json:"tool_used"`
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	TotalRecords      *int       `json:"total_records,omitempty"`
	SuccessfulRecords *int       `json:"successful_records,omitempty"`
	FailedRecords     int        `json:"failed_records"`
	Status            string     `json:"status"`
	FileSizeMB        *float64   `json:"file_size_mb,omitempty"`
}

// validatePerformanceMetricRequest validates the incoming request
func validatePerformanceMetricRequest(req CreatePerformanceMetricRequest) error {
	if req.BatchJobID <= 0 {
		return fmt.Errorf("batch job id must be positive")
	}

	if req.MetricType == "" {
		return fmt.Errorf("metric type can not be empty")
	}

	if req.MetricValue < 0 {
		return fmt.Errorf("metric value must non-negative")
	}

	if req.Unit == "" {
		return fmt.Errorf("unit cannot be empty")
	}
	return nil
}

// validateValueMetricRequest validates the incoming request
func validateValueMetricRequest(req CreateValueMetricRequest) error {
	// Check: batch_job_id > 0, required fields not empty/negative, percentages in valid range
	if req.ErrorReductionPercentage < -1 || req.ErrorReductionPercentage > 1 {
		return fmt.Errorf("error reduction percentage must be less than 1 and greater than -1 - got:  %v", req.ErrorReductionPercentage)
	}

	if req.CostSavings <= 0 {
		return fmt.Errorf("cost savings must be positive")
	}

	if req.HourlyRate < 30 {
		return fmt.Errorf("hourly rate must be at least 30; got:  %v", req.HourlyRate)
	}

	if req.ActualProcessingMinutes < 0 {
		return fmt.Errorf("actual processing minutes can not be negative")
	}

	if req.BatchJobID <= 0 {
		return fmt.Errorf("batch job must be positive")
	}

	if req.ManualTimeEstimateMinutes <= 0 {
		return fmt.Errorf("manual time estimate minutes must be positive")
	}

	expectedTimeSaved := float64(req.ManualTimeEstimateMinutes) - req.ActualProcessingMinutes
	tolerance := 0.1 // Allow 0.1 minute difference for rounding
	if math.Abs(req.TimeSavedMinutes-expectedTimeSaved) > tolerance {
		return fmt.Errorf("time saved minutes inconsistent with manual estimate and actual processing time")
	}

	return nil
}

// validateBatchJobRequest validates the incoming request
func validateBatchJobRequest(req CreateBatchJobRequest) error {
	if req.JobName == "" {
		return fmt.Errorf("job name can not be empty")
	}

	if req.ToolUsed == "" {
		return fmt.Errorf("tool used can not be empty")
	}

	//TODO: this is set to check empty for now. check known status later
	if req.Status == "" {
		return fmt.Errorf("status can not be empty")
	}

	if req.FailedRecords < 0 {
		return fmt.Errorf("failed records must be non-negative")
	}

	if req.TotalRecords != nil {
		if *req.TotalRecords <= 0 {
			return fmt.Errorf("total records must be positive")
		}
	}

	if req.SuccessfulRecords != nil {
		if *req.SuccessfulRecords <= 0 {
			return fmt.Errorf("successful records must be positive")
		}
	}

	if req.FileSizeMB != nil {
		if *req.FileSizeMB <= 0 {
			return fmt.Errorf("file size MB must be non-negative")
		}
	}

	if req.CompletedAt != nil {
		if req.CompletedAt.Before(req.StartedAt) {
			return fmt.Errorf("completed at must be after started at")
		}
	}
	return nil
}

// PostPerformanceMetric godoc
// @Summary      Create a new performance metric
// @Description  Create a performance metric record for a batch job
// @Tags         Metrics
// @Accept       json
// @Produce      json
// @Param        metric body CreatePerformanceMetricRequest true "Performance metric data"
// @Success      201   {object}  CreatePerformanceMetricResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /metrics/performance [post]

// PostValueMetric godoc
// @Summary      Create a new value metric
// @Description  Create a value metric record for a batch job
// @Tags         Metrics
// @Accept       json
// @Produce      json
// @Param        metric body CreateValueMetricRequest true "Value metric data"
// @Success      201   {object}  CreateValueMetricResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /metrics/value [post]

// PostBatchJob godoc
// @Summary      Create a new batch job
// @Description  Create a batch job record
// @Tags         Metrics
// @Accept       json
// @Produce      json
// @Param        job body CreateBatchJobRequest true "Batch job data"
// @Success      201   {object}  CreateBatchJobResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /metrics/batch-jobs [post]

// GetBatchJob godoc
// @Summary      Get a batch job by ID
// @Description  Retrieve a batch job record by its ID
// @Tags         Metrics
// @Accept       json
// @Produce      json
// @Param        id path int true "Batch Job ID"
// @Success      200   {object}  GetBatchJobResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /metrics/batch-jobs/{id} [get]

func RegisterMetrics(mux *http.ServeMux, a *app.App, logger *slog.Logger) {
	// POST /metrics/performance - Complete pattern implementation
	performanceHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		// PATTERN 1: Parse HTTP Request → Request DTO
		var req CreatePerformanceMetricRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		// PATTERN 2: Validate Request DTO
		if err := validatePerformanceMetricRequest(req); err != nil {
			logger.Warn("request validation failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		// PATTERN 3: Transform Request DTO → Domain Model
		now := time.Now()
		domainMetric := metrics.PerformanceMetric{
			BatchJobID:  req.BatchJobID,
			MetricType:  req.MetricType,
			MetricValue: &req.MetricValue,
			Unit:        &req.Unit,
			MeasuredAt:  &now,
		}

		// PATTERN 4: Execute Business Logic (Repository Call)
		if err := a.MetricsRepo.PostPerformanceMetric(r.Context(), domainMetric); err != nil {
			logger.Error("business logic failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to create performance metric")
			return
		}

		// PATTERN 5: Transform Domain Model → Response DTO
		response := CreatePerformanceMetricResponse{
			ID:          domainMetric.ID,
			BatchJobID:  domainMetric.BatchJobID,
			MetricType:  domainMetric.MetricType,
			MetricValue: *domainMetric.MetricValue,
			Unit:        *domainMetric.Unit,
			MeasuredAt:  *domainMetric.MeasuredAt,
		}

		// PATTERN 6: Return HTTP Response
		logger.Info("performance metric created successfully",
			slog.String("request_id", requestID),
			slog.Int("metric_id", domainMetric.ID),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("response encoding failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
		}
	}

	valueHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		var req CreateValueMetricRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		if err := validateValueMetricRequest(req); err != nil {
			logger.Warn("request validation failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		now := time.Now()
		domainMetric := metrics.ValueMetric{
			BatchJobID:                req.BatchJobID,
			ManualTimeEstimateMinutes: &req.ManualTimeEstimateMinutes,
			ActualProcessingMinutes:   &req.ActualProcessingMinutes,
			TimeSavedMinutes:          &req.TimeSavedMinutes,
			HourlyRate:                &req.HourlyRate,
			CostSavings:               &req.CostSavings,
			ErrorReductionPercentage:  &req.ErrorReductionPercentage,
			CalculatedAT:              &now,
		}

		if err := a.MetricsRepo.PostValueMetric(r.Context(), domainMetric); err != nil {
			logger.Error("business logic failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to create value metric")
			return
		}

		response := CreateValueMetricResponse{
			ID:                        domainMetric.ID,
			BatchJobID:                domainMetric.BatchJobID,
			ManualTimeEstimateMinutes: *domainMetric.ManualTimeEstimateMinutes,
			ActualProcessingMinutes:   *domainMetric.ActualProcessingMinutes,
			TimeSavedMinutes:          *domainMetric.TimeSavedMinutes,
			HourlyRate:                *domainMetric.HourlyRate,
			CostSavings:               *domainMetric.CostSavings,
			ErrorReductionPercentage:  *domainMetric.ErrorReductionPercentage,
			CalculatedAt:              *domainMetric.CalculatedAT,
		}

		logger.Info("value metric created successfully",
			slog.String("request_id", requestID),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("response encoding failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
		}
	}

	batchJobHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		var req CreateBatchJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		if err := validateBatchJobRequest(req); err != nil {
			logger.Warn("request validation failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		now := time.Now()
		domainMetric := metrics.BatchJob{
			JobName:           req.JobName,
			ToolUsed:          req.ToolUsed,
			StartedAt:         now,
			CompletedAt:       req.CompletedAt,
			TotalRecords:      req.TotalRecords,
			SuccessfulRecords: req.SuccessfulRecords,
			FailedRecords:     req.FailedRecords,
			Status:            req.Status,
			FileSizeMB:        req.FileSizeMB,
		}

		if err := a.MetricsRepo.PostBatchJob(r.Context(), &domainMetric); err != nil {
			logger.Error("business logic failed",
				slog.String("request id", requestID),
				slog.String("error", err.Error()),
			)

			writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to create a batch job")
			return
		}

		response := CreateBatchJobResponse{
			ID:                domainMetric.ID,
			JobName:           domainMetric.JobName,
			ToolUsed:          domainMetric.ToolUsed,
			StartedAt:         domainMetric.StartedAt,
			CompletedAt:       domainMetric.CompletedAt,
			TotalRecords:      domainMetric.TotalRecords,
			SuccessfulRecords: domainMetric.SuccessfulRecords,
			FailedRecords:     domainMetric.FailedRecords,
			Status:            domainMetric.Status,
			FileSizeMB:        domainMetric.FileSizeMB,
			CreatedAt:         domainMetric.CreatedAt,
		}

		logger.Info("batch job created successfully",
			slog.String("request_id", requestID),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("response encoding failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
		}
	}

	getBatchJobHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Missing batch job ID")
			return
		}
		idStr := pathParts[3]

		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_id", "Batch job ID must be a positive integer")
			return
		}

		batchJob, err := a.MetricsRepo.GetBatchJob(r.Context(), id)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}

		if batchJob == nil {
			writeErrorResponse(w, http.StatusNotFound, "not_found", "Batch job not found")
			return
		}

		response := GetBatchJobResponse{
			ID:                batchJob.ID,
			JobName:           batchJob.JobName,
			ToolUsed:          batchJob.ToolUsed,
			StartedAt:         batchJob.StartedAt,
			CompletedAt:       batchJob.CompletedAt,
			TotalRecords:      batchJob.TotalRecords,
			SuccessfulRecords: batchJob.SuccessfulRecords,
			FailedRecords:     batchJob.FailedRecords,
			Status:            batchJob.Status,
			FileSizeMB:        batchJob.FileSizeMB,
		}

		logger.Info("batch job retrieved successfully",
			slog.String("request_id", requestID),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("response encoding failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
		}
	}

	// Register endpoints with middleware
	mux.HandleFunc("POST /metrics/performance",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(performanceHandler)))

	mux.HandleFunc("POST /metrics/value",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(valueHandler)))

	mux.HandleFunc("POST /metrics/batch-jobs",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(batchJobHandler)))

	mux.HandleFunc("GET /metrics/batch-jobs/",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(getBatchJobHandler)))
}
