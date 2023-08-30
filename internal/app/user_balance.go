package app

import (
	"encoding/json"
	"kaunnikov/internal/services"
	"net/http"
)

func (m *app) UserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	UUID := r.Context().Value("uuid")
	balance, err := services.GetBalanceAccountByUUID(UUID)
	if err != nil {
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(balance)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(resp)
	w.WriteHeader(http.StatusOK)
	return
}
