package httpx

import (
	"net/http"

	"NTRCodes/member-api/internal/app"
)

// RegisterHealth mounts /healthz and /readyz.
func RegisterHealth(mux *http.ServeMux, a *app.App) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ok, reason := a.Ready(r.Context())
		if !ok {
			http.Error(w, reason, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
}
