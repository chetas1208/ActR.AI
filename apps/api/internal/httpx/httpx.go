package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// WriteJSON encodes v as JSON and writes it to w with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// WriteValidationError writes a 422 validation error response.
func WriteValidationError(w http.ResponseWriter, field, message string) {
	WriteJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
		"error":  "validation_error",
		"field":  field,
		"detail": message,
	})
}

// DecodeJSON decodes r.Body into dst, closing the body.
func DecodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// RequireMethod returns false and writes 405 if r.Method != method.
func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		w.Header().Set("Allow", method)
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return false
	}
	return true
}

// WithCORS adds permissive CORS headers for the API.
func WithCORS(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Webhook-Secret")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ApplyCORS writes CORS headers directly onto w (for Vercel function handlers without middleware chain).
func ApplyCORS(w http.ResponseWriter, r *http.Request, allowedOrigins string) bool {
	origin := r.Header.Get("Origin")
	if isOriginAllowed(origin, allowedOrigins) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Webhook-Secret")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

// ReadPathParam extracts a path segment by position (0-based) from r.URL.Path
// after stripping the given prefix.
func ReadPathParam(r *http.Request, prefix string, pos int) string {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", -1)
	if pos < len(parts) {
		return parts[pos]
	}
	return ""
}

// ValidateUUID parses s as UUID, writing 400 and returning false on failure.
func ValidateUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid UUID: "+s)
		return uuid.UUID{}, false
	}
	return id, true
}

// ProviderNotConfiguredError writes a 503 with a helpful message about the missing env var.
func ProviderNotConfiguredError(w http.ResponseWriter, provider, envVar string) {
	WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
		"error":    "provider_not_configured",
		"provider": provider,
		"detail":   fmt.Sprintf("Provider not configured: add %s to enable %s.", envVar, provider),
	})
}

func isOriginAllowed(origin, allowed string) bool {
	if origin == "" || allowed == "" {
		return false
	}
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}
