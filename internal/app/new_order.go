package app

import (
	"fmt"
	"github.com/theplant/luhn"
	"io"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
	"strconv"
)

func (m *app) NewOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		logging.Infof("Invalid Content Type: %s", r.Header.Get("Content-Type"))
		http.Error(w, "Invalid Content Type!", http.StatusBadRequest)
		return
	}

	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		logging.Infof("cannot read request body: %s", err)
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
		return
	}

	if string(responseData) == "" {
		logging.Infof("Empty POST request body! %s", r.URL)
		http.Error(w, "Empty POST request body!", http.StatusBadRequest)
		return
	}

	// Првоеряем на корректность Луна
	number, err := strconv.Atoi(string(responseData))
	if err != nil {
		logging.Errorf("Не удалось преобразовать в число: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if number == 0 || luhn.Valid(number) == false {
		logging.Infof("Номер %s не прошёл валидацию Луна", string(responseData))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Проверяем на наличие в БД заказа с таким номером
	order, err := services.GetOrderByNumber(number)
	if err != nil {
		logging.Infof("Ошибка получения заказа из БД: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	UUID := r.Context().Value("uuid")
	if UUID == "" {
		logging.Errorf("Потеряли UUID")
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}
	if order != nil {
		// Если есть и загружен этим пользователем - ответ 200
		if order.UUID == UUID {
			logging.Infof("Заказ %d зарегистрирован этим пользователем ранее", number)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Если есть и загружен другим пользователем - ответ 409
		logging.Infof("Заказ %d зарегистрирован другим пользователем ранее", number)
		w.WriteHeader(http.StatusConflict)
		return
	}

	if err := services.CreateOrder(number, UUID); err != nil {
		logging.Infof("Ошибка создания заказа", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
