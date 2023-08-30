package app

import (
	"encoding/json"
	"fmt"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
)

func (m *app) GetWithdrawsHandler(w http.ResponseWriter, r *http.Request) {
	UUID := r.Context().Value("uuid")

	// Получим отсортированный по времени список вывода средств пользователя
	result, err := services.GetWithdrawByUUID(UUID)
	if err != nil {
		logging.Errorf("Не удалось получить списания пользователя: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если нет списаний - 204
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(result)
	if err != nil {
		logging.Errorf("cannot encode response: %s", err)
		http.Error(w, fmt.Sprintf("cannot encode response: %s", err), http.StatusBadRequest)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		logging.Errorf("cannot write response to the client: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
