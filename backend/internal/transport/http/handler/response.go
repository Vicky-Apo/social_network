package handler

import (
	"net/http"

	"social-network/backend/internal/transport/http/utils"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	utils.RespondWithSuccess(w, status, payload)
}
