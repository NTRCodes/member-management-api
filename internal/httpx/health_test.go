// internal/httpx/health_test.go
package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"NTRCodes/member-api/internal/app"
)

func TestHealthz(t *testing.T) {
	t.Setenv("PORT", "8000")

	mux := http.NewServeMux()
	RegisterHealth(mux, app.New(nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if body := w.Body.String(); body != "ok" {
		t.Fatalf("got %q", body)
	}
}

func TestReadyz_OK(t *testing.T) {
	t.Setenv("PORT", "8000")

	mux := http.NewServeMux()
	RegisterHealth(mux, app.New(nil))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("want %d, got %d", http.StatusOK, w.Code)
	}
	if body := w.Body.String(); body != "ready" {
		t.Fatalf("expected body 'ready', got %q", body)
	}
}

func TestReadyz_MissingEnv(t *testing.T) {
	t.Setenv("PORT", "")

	mux := http.NewServeMux()
	RegisterHealth(mux, app.New(nil))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "missing env:") {
		t.Fatalf("expected missing env reason, got %q", w.Body.String())
	}
}
