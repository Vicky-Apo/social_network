package utils

import (
	"encoding/json"
	"net/http"

	"social-network/backend/pkg/logger"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RespondWithE	rror writes a JSON error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := APIResponse{
		Success: false,
		Error:   message,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.NewDefault(false).Debug("failed to encode error response", logger.F("error", err.Error()))
	}
}

// RespondWithSuccess writes a JSON success response
func RespondWithSuccess(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := APIResponse{
		Success: true,
		Data:    payload,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.NewDefault(false).Debug("failed to encode success response", logger.F("error", err.Error()))
	}
}
