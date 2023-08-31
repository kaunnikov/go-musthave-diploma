package app

import (
	"encoding/json"
	"kaunnikov/internal/logging"
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
	if err != nil {
		logging.Errorf("cannot encode response: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	_, err = w.Write(resp)
	if err != nil {
		logging.Errorf("cannot write response to the client: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
