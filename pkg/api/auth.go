package api

import (
	"Yandex_final/pkg/auth"
	"encoding/json"
	"net/http"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	token, err := auth.GenerateToken(req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
