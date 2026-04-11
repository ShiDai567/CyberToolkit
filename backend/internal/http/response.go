package http

import (
	"encoding/json"
	"net/http"
)

type meta map[string]any

func writeJSON(w http.ResponseWriter, status int, data any, m meta) {
	w.WriteHeader(status)
	payload := map[string]any{"data": data}
	if m != nil {
		payload["meta"] = m
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	w.WriteHeader(status)
	payload := map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": details,
		},
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
