package httpx

import (
	"NTRCodes/member-api/internal/database/members"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(requestIDKey); requestID != nil {
		return requestID.(string)
	}
	return ""
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// RateLimiter manages rate limiting per IP address
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    float64
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burstSize int) *RateLimiter {
	rateLimiter := RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    requestsPerSecond,
		burst:    burstSize,
	}

	return &rateLimiter
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.limit), rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter.Allow()
}

// GetClientIP grabs the client's IP from
func getClientIP(r *http.Request) string {
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For may contain multiple IPs; take the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	if xRealIp := r.Header.Get("X-Real-IP"); xRealIp != "" {
		return strings.TrimSpace(xRealIp)
	}

	if cfConnectingIp := r.Header.Get("CF-Connecting-IP"); cfConnectingIp != "" {
		return strings.TrimSpace(cfConnectingIp)
	}

	// Fallback to RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

//

// RateLimitMiddleware provides rate limiting per IP address
func RateLimitMiddleware(rateLimiter *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.Contains(authHeader, "Bearer") {
				apiKey := strings.TrimPrefix(authHeader, "Bearer ")
				if isTrustedAPIKey(apiKey) {
					next(w, r)
					return
				}
			}
			clientIP := getClientIP(r)
			if !rateLimiter.Allow(clientIP) {
				writeRateLimitResponse(w)
				return
			}
			next(w, r)
		}
	}
}

// writeRateLimitResponse writes a 429 Too Many Requests response
func writeRateLimitResponse(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "60")
	writeErrorResponse(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests. Please try again later.")
}

// isTrustedAPIKey checks if the provided API key should bypass rate limiting
func isTrustedAPIKey(apiKey string) bool {
	return apiKey == os.Getenv("API_KEY")
}

// LoggingMiddleware provides structured logging for HTTP requests
func LoggingMiddleware(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID and add to context
			requestID := generateRequestID()
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Add request ID to response headers for tracing
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w}

			// Log request start
			logger.Info("request started",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("user_agent", r.UserAgent()),
				slog.String("remote_addr", r.RemoteAddr),
			)

			// Call next handler
			next(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			logger.Info("request completed",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status_code", rw.statusCode),
				slog.Int("response_size", rw.written),
				slog.Duration("duration", duration),
			)
		}
	}
}

// ValidateMemberIDMiddleware validates the member ID parameter before passing to the handler
func ValidateMemberIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		// Extract the ID parameter
		id := r.URL.Query().Get("id")
		if id == "" {
			slog.Warn("validation failed: missing parameter",
				slog.String("request_id", requestID),
				slog.String("error", "missing id parameter"),
			)
			writeErrorResponse(w, http.StatusBadRequest, "missing_parameter", "id parameter is required")
			return
		}

		// Validate using your custom validation
		if err := members.CheckValidMemberNumber(id); err != nil {
			// Handle validation errors with proper HTTP responses
			var customErr members.InvalidMemberNumberError
			if errors.As(err, &customErr) {
				slog.Warn("validation failed: invalid input",
					slog.String("request_id", requestID),
					slog.String("input", id),
					slog.String("error", customErr.Error()),
				)
				writeErrorResponse(w, http.StatusBadRequest, "invalid_input", customErr.Error())
			}
			return
		}

		slog.Debug("validation passed",
			slog.String("request_id", requestID),
			slog.String("member_id", id),
		)

		// Validation passed - call the next handler
		next(w, r)
	}
}

// writeErrorResponse writes a standardized JSON error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorMsg, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorMsg,
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// APIKeyAuthMiddleware validates API key from Authorization header
func APIKeyAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		// Get the expected API key from environment
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			slog.Error("API key not configured",
				slog.String("request_id", requestID),
			)
			writeErrorResponse(w, http.StatusInternalServerError, "server_error", "API key not configured")
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.Warn("authentication failed: missing header",
				slog.String("request_id", requestID),
				slog.String("remote_addr", r.RemoteAddr),
			)
			writeErrorResponse(w, http.StatusUnauthorized, "missing_auth", "Authorization header required")
			return
		}

		// Check for "Bearer <key>" format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			slog.Warn("authentication failed: invalid format",
				slog.String("request_id", requestID),
				slog.String("remote_addr", r.RemoteAddr),
			)
			writeErrorResponse(w, http.StatusUnauthorized, "invalid_auth_format", "Authorization header must be 'Bearer <api-key>'")
			return
		}

		// Extract the key
		providedKey := strings.TrimPrefix(authHeader, "Bearer ")
		if providedKey != expectedKey {
			slog.Warn("authentication failed: invalid key",
				slog.String("request_id", requestID),
				slog.String("remote_addr", r.RemoteAddr),
			)
			writeErrorResponse(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key")
			return
		}

		slog.Debug("authentication successful",
			slog.String("request_id", requestID),
		)

		// Authentication passed - call next handler
		next(w, r)
	}
}
