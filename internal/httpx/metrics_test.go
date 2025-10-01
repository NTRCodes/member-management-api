package httpx

import (
	"strings"
	"testing"
)

func TestValidatePerformanceMetricRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     CreatePerformanceMetricRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  1,
				MetricType:  "processing_time",
				MetricValue: 45.5,
				Unit:        "seconds",
			},
			expectError: false,
		},
		{
			name: "zero batch_job_id",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  0,
				MetricType:  "processing_time",
				MetricValue: 45.5,
				Unit:        "seconds",
			},
			expectError: true,
			errorMsg:    "batch job id must be positive",
		},
		{
			name: "negative batch_job_id",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  -1,
				MetricType:  "processing_time",
				MetricValue: 45.5,
				Unit:        "seconds",
			},
			expectError: true,
			errorMsg:    "batch job id must be positive",
		},
		{
			name: "empty metric_type",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  1,
				MetricType:  "",
				MetricValue: 45.5,
				Unit:        "seconds",
			},
			expectError: true,
			errorMsg:    "metric type can not be empty",
		},
		{
			name: "negative metric_value",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  1,
				MetricType:  "processing_time",
				MetricValue: -1.0,
				Unit:        "seconds",
			},
			expectError: true,
			errorMsg:    "metric value must non-negative",
		},
		{
			name: "zero metric_value (should be valid)",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  1,
				MetricType:  "processing_time",
				MetricValue: 0.0,
				Unit:        "seconds",
			},
			expectError: false,
		},
		{
			name: "empty unit",
			request: CreatePerformanceMetricRequest{
				BatchJobID:  1,
				MetricType:  "processing_time",
				MetricValue: 45.5,
				Unit:        "",
			},
			expectError: true,
			errorMsg:    "unit cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePerformanceMetricRequest(tc.request)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectError && err != nil && !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tc.errorMsg, err.Error())
			}
		})
	}
}

func TestValidateValueMetricRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateValueMetricRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: false,
		},
		{
			name: "zero batch_job_id",
			request: CreateValueMetricRequest{
				BatchJobID:                0,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "batch job must be positive",
		},
		{
			name: "zero manual time estimate",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 0,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "manual time estimate minutes must be positive",
		},
		{
			name: "negative actual processing minutes",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   -5.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "actual processing minutes can not be negative",
		},
		{
			name: "hourly rate too low",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                25.0,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "hourly rate must be at least 30",
		},
		{
			name: "negative cost savings",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               -10.0,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "cost savings must be positive",
		},
		{
			name: "error reduction percentage too high",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  1.5,
			},
			expectError: true,
			errorMsg:    "error reduction percentage must be less than 1",
		},
		{
			name: "error reduction percentage too low",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.0,
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  -1.5,
			},
			expectError: true,
			errorMsg:    "error reduction percentage must be less than 1 and greater than -1",
		},
		{
			name: "inconsistent time saved minutes",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          30.0, // Should be 50.0
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: true,
			errorMsg:    "time saved minutes inconsistent",
		},
		{
			name: "time saved within tolerance",
			request: CreateValueMetricRequest{
				BatchJobID:                1,
				ManualTimeEstimateMinutes: 60,
				ActualProcessingMinutes:   10.0,
				TimeSavedMinutes:          50.05, // Within 0.1 tolerance
				HourlyRate:                31.25,
				CostSavings:               26.04,
				ErrorReductionPercentage:  0.85,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateValueMetricRequest(tc.request)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectError && err != nil && !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tc.errorMsg, err.Error())
			}
		})
	}
}

func TestValidateBatchJobRequest(t *testing.T) {
	// TODO: Implement when validateBatchJobRequest is completed
	t.Skip("Batch job validation not yet implemented")
}
