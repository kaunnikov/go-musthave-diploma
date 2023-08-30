package app

import (
	"encoding/json"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
)

func (m *app) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {

	UUID := r.Context().Value("uuid")
	orders, err := services.GetOrdersByUUID(UUID)
	if err != nil {
		logging.Infof("Ошибка получения заказов пользователя: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		logging.Infof("cannot encode response: %s", err)
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
	return
}
