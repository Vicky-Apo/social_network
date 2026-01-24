package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithError writes a JSON error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("failed to encode error response: %v", err)
	}
}

// RespondWithSuccess writes a JSON success response
func RespondWithSuccess(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode success response: %v", err)
	}
}