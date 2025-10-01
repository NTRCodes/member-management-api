package httpx

import (
	"NTRCodes/member-api/internal/app"
	"NTRCodes/member-api/internal/database/members"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Member Request/Response DTOs

// CreateMemberRequest represents the request body for creating a new member
type CreateMemberRequest struct {
	UBCID       string `json:"ubcid" validate:"required"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number"`
	SSN         string `json:"ssn" validate:"required"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
	Phone       string `db:"phone1"`
	Phone2      string `db:"phone2"`
	Local       string `db:"local"`
	Status      string `db:"status"`
	Class       string `db:"class"`
	DOB         string `db:"dob"`
	Gender      string `db:"gender"`
	InitDate    string `db:"init_date"`
	Delegate    string `db:"delegate"`
	Language    string `db:"lang"`
	Veteran     string `db:"veteran"`
	LastUpdate  string `db:"stamp"`
}

// CreateMemberResponse represents the response after creating a member
type CreateMemberResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	UBCID     string    `json:"ubcid"`
	Local     string    `json:"local"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetMemberResponse represents the response for retrieving a member
type GetMemberResponse struct {
	ID          int    `json:"id"`
	UBCID       string `json:"ubcid" validate:"required"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number"`
	SSN         string `json:"ssn" validate:"required"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
	Local       string `db:"local"`
	Status      string `db:"status"`
	Class       string `db:"class"`
	DOB         string `db:"dob"`
	Gender      string `db:"gender"`
	InitDate    string `db:"init_date"`
	Delegate    string `db:"delegate"`
	Language    string `db:"lang"`
	Veteran     string `db:"veteran"`
	LastUpdate  string `db:"stamp"`
}

// BatchUpsertMembersRequest represents a batch upsert request
type BatchUpsertMembersRequest struct {
	Members []CreateMemberRequest `json:"members"`
}

// BatchUpsertMembersResponse represents the response from a batch upsert
type BatchUpsertMembersResponse struct {
	TotalProcessed int                    `json:"total_processed"`
	Created        int                    `json:"created"`
	Updated        int                    `json:"updated"`
	Failed         int                    `json:"failed"`
	Errors         []BatchErrorResponse   `json:"errors,omitempty"`
	ProcessingTime string                 `json:"processing_time_ms"`
}

// BatchErrorResponse represents an error in batch processing
type BatchErrorResponse struct {
	Index    int    `json:"index"`
	MemberID string `json:"member_id"`
	Error    string `json:"error"`
}

// Validation Functions

// validateCreateMemberRequest validates the create member request
func validateCreateMemberRequest(req CreateMemberRequest) error {
	if req.UBCID == "" {
		return fmt.Errorf("UBC ID cannot be empty")
	}

	if req.FirstName == "" {
		return fmt.Errorf("first name cannot be empty")
	}

	if req.LastName == "" {
		return fmt.Errorf("last name cannot be empty")
	}

	if req.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if req.SSN == "" {
		return fmt.Errorf("SSN cannot be empty")
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("email must be a valid email address")
	}

	return nil
}

// CreateMember godoc
// @Summary      Create a new member
// @Description  Create a new member record in the system
// @Tags         Members
// @Accept       json
// @Produce      json
// @Param        member body CreateMemberRequest true "Member data"
// @Success      201   {object}  CreateMemberResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /members [post]

// GetMember godoc
// @Summary      Get member by ID
// @Description  Retrieve member information by UBC ID or SSN
// @Tags         Members
// @Accept       json
// @Produce      json
// @Param        id    path     string  true  "Member ID (UBC ID or SSN)"
// @Param        fullSSN query  boolean false "Return full SSN (default: last 4 digits)"
// @Success      200   {object}  GetMemberResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /members/{id} [get]

// UpdateMember godoc
// @Summary      Update an existing member
// @Description  Update member information by UBC ID or SSN
// @Tags         Members
// @Accept       json
// @Produce      json
// @Param        id     path     string  true  "Member ID (UBC ID or SSN)"
// @Param        member body     UpdateMemberRequest true "Updated member data"
// @Success      200   {object}  UpdateMemberResponse
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /members/{id} [put]

// DeleteMember godoc
// @Summary      Delete a member
// @Description  Delete member record by UBC ID or SSN
// @Tags         Members
// @Accept       json
// @Produce      json
// @Param        id    path     string  true  "Member ID (UBC ID or SSN)"
// @Success      204   "Member deleted successfully"
// @Failure      400   {object}  httpx.ErrorResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Security     BearerAuth
// @Router       /members/{id} [delete]

// RegisterMemberRoutes registers all member-related HTTP routes
func RegisterMemberRoutes(mux *http.ServeMux, a *app.App, logger *slog.Logger) {

	createMemberHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		var req CreateMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		if err := validateCreateMemberRequest(req); err != nil {
			logger.Warn("request validation failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		now := time.Now().String()
		member := members.Member{
			UBCID:      req.UBCID,
			SSN:        &req.SSN,
			FirstName:  &req.FirstName,
			LastName:   &req.LastName,
			Address:    &req.Address,
			City:       &req.City,
			State:      &req.State,
			Zip:        &req.ZipCode,
			Phone:      &req.PhoneNumber,
			Email:      &req.Email,
			Local:      &req.Local,
			Status:     &req.Status,
			Class:      &req.Class,
			DOB:        &req.DOB,
			Gender:     &req.Gender,
			InitDate:   &req.InitDate,
			Delegate:   &req.Delegate,
			Language:   &req.Language,
			Veteran:    &req.Veteran,
			LastUpdate: &now,
		}

		if err := a.MemberRepo.CreateMember(r.Context(), &member); err != nil {
			logger.Error("business logic failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to create member")
			return
		}

		// Create and return CreateMemberResponse with 201 status
		response := CreateMemberResponse{
			ID:        member.ID,
			FirstName: *member.FirstName,
			LastName:  *member.LastName,
			UBCID:     member.UBCID,
			Local:     *member.Local,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		logger.Info("member created successfully",
			slog.String("request_id", requestID),
			slog.String("ubcid", member.UBCID),
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

	getMemberHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Missing memNum")
			return
		}

		fullSSN := false
		if len(pathParts) >= 4 {
			if strings.ToLower(pathParts[3]) == "true" {
				fullSSN = true
			}
		}

		memNum := pathParts[2]
		if err := members.CheckValidMemberNumber(memNum); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_id", "member id or num was invalid")
			return
		}

		member, err := a.MemberRepo.GetMember(r.Context(), memNum, fullSSN)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		if member == nil {
			writeErrorResponse(w, http.StatusNotFound, "not_found", "member not found ")
			return
		}

		response := GetMemberResponse{
			ID:          member.ID,
			UBCID:       member.UBCID,
			FirstName:   *member.FirstName,
			LastName:    *member.LastName,
			Email:       *member.Email,
			PhoneNumber: *member.Phone,
			SSN:         *member.SSN,
			Address:     *member.Address,
			City:        *member.City,
			State:       *member.State,
			ZipCode:     *member.Zip,
			Local:       *member.Local,
			Status:      *member.Status,
			Class:       *member.Class,
			DOB:         *member.DOB,
			Gender:      *member.Gender,
			Delegate:    *member.Delegate,
			Language:    *member.Language,
			Veteran:     *member.Veteran,
			LastUpdate:  *member.LastUpdate,
		}

		logger.Info("member retrieved successfully",
			slog.String("request_id", requestID),
			slog.String("member_num", memNum),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("response encoding failed",
				slog.String("request_id", requestID),
				slog.String("member_num", memNum),
				slog.String("error", err.Error()),
			)
		}
	}

	updateMemberHandler := func(w http.ResponseWriter, r *http.Request) {
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

		var req members.UpdateMemberRequest // this is a name for the shape of the data
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		if err := a.MemberRepo.UpdateMember(r.Context(), idStr, req); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}

		logger.Info("member updated successfully",
			slog.String("request_id", requestID),
			slog.String("member_id", idStr),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}

	deleteMemberHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		// Extract ID from path (same as GET pattern)
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_path", "Missing member ID")
			return
		}

		memberID := pathParts[2]

		// Validate member ID format
		if err := members.CheckValidMemberNumber(memberID); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid member ID format")
			return
		}

		// Delete the member
		if err := a.MemberRepo.DeleteMember(r.Context(), memberID); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeErrorResponse(w, http.StatusNotFound, "not_found", "Member not found")
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "database_error", "Failed to delete member")
			}
			return
		}

		logger.Info("member deleted successfully",
			slog.String("request_id", requestID),
			slog.String("member_id", memberID),
		)

		// 204 No Content for successful deletion
		w.WriteHeader(http.StatusNoContent)
	}

	// Register endpoints with middleware
	mux.HandleFunc("POST /members",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(createMemberHandler)))

	mux.HandleFunc("GET /members/",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(getMemberHandler)))

	mux.HandleFunc("PUT /members/",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(updateMemberHandler)))

	mux.HandleFunc("DELETE /members/",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(deleteMemberHandler)))

	// High-performance batch upsert endpoint
	batchUpsertHandler := func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		startTime := time.Now()

		var req BatchUpsertMembersRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("batch upsert request parsing failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
			return
		}

		if len(req.Members) == 0 {
			writeErrorResponse(w, http.StatusBadRequest, "empty_batch", "Members array cannot be empty")
			return
		}

		// Convert HTTP request to domain models
		domainMembers := make([]*members.Member, len(req.Members))
		for i, memberReq := range req.Members {
			domainMembers[i] = &members.Member{
				UBCID:      memberReq.UBCID,
				SSN:        &memberReq.SSN,
				FirstName:  &memberReq.FirstName,
				LastName:   &memberReq.LastName,
				Address:    &memberReq.Address,
				City:       &memberReq.City,
				State:      &memberReq.State,
				Zip:        &memberReq.ZipCode,
				Phone:      &memberReq.PhoneNumber,
				Email:      &memberReq.Email,
				Local:      &memberReq.Local,
				Status:     &memberReq.Status,
				Class:      &memberReq.Class,
				DOB:        &memberReq.DOB,
				Gender:     &memberReq.Gender,
				InitDate:   &memberReq.InitDate,
				Delegate:   &memberReq.Delegate,
				Language:   &memberReq.Language,
				Veteran:    &memberReq.Veteran,
				LastUpdate: &memberReq.LastUpdate,
			}
		}

		// Execute batch upsert
		result, err := a.MemberRepo.BatchUpsertMembers(r.Context(), domainMembers)
		if err != nil {
			logger.Error("batch upsert failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
				slog.Int("batch_size", len(req.Members)),
			)
			writeErrorResponse(w, http.StatusInternalServerError, "batch_upsert_failed", "Failed to process batch upsert")
			return
		}

		// Convert batch errors to response format
		responseErrors := make([]BatchErrorResponse, len(result.Errors))
		for i, batchErr := range result.Errors {
			responseErrors[i] = BatchErrorResponse{
				Index:    batchErr.Index,
				MemberID: batchErr.MemberID,
				Error:    batchErr.Error,
			}
		}

		processingTime := time.Since(startTime)
		response := BatchUpsertMembersResponse{
			TotalProcessed: result.TotalProcessed,
			Created:        result.Created,
			Updated:        result.Updated,
			Failed:         result.Failed,
			Errors:         responseErrors,
			ProcessingTime: fmt.Sprintf("%.2f", float64(processingTime.Nanoseconds())/1e6),
		}

		logger.Info("batch upsert completed",
			slog.String("request_id", requestID),
			slog.Int("total_processed", result.TotalProcessed),
			slog.Int("created", result.Created),
			slog.Int("updated", result.Updated),
			slog.Int("failed", result.Failed),
			slog.String("processing_time_ms", response.ProcessingTime),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	mux.HandleFunc("POST /members/batch",
		LoggingMiddleware(logger)(
			APIKeyAuthMiddleware(batchUpsertHandler)))
}
