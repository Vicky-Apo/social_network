package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RespondWithError writes a JSON error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := APIResponse{
		Success: false,
		Error:   message,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode error response: %v", err)
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
		log.Printf("failed to encode success response: %v", err)
	}
}
