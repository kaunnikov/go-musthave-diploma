package app

import (
	"encoding/json"
	"fmt"
	"github.com/theplant/luhn"
	"io"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
	"strconv"
)

type withdrawRequest struct {
	Number string  `json:"order"`
	Sum    float32 `json:"sum"`
}

func (m *app) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	UUID := r.Context().Value("uuid")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid Content Type!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logging.Errorf("cannot read request body: %s", err)
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, fmt.Sprintf("Empty request body %s", body), http.StatusBadRequest)
		return
	}

	var wr withdrawRequest
	err = json.Unmarshal(body, &wr)
	if err != nil {
		logging.Errorf("cannot decode request body to `JSON`: %s", err)
		http.Error(w, fmt.Sprintf("cannot decode request body to `JSON`: %s", err), http.StatusBadRequest)
		return
	}

	number, err := strconv.Atoi(wr.Number)
	if err != nil {
		logging.Errorf("Не удалось преобразовать в число: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if number == 0 {
		logging.Infof("Неверный номер заказа: %s", wr.Number)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if !luhn.Valid(number) {
		logging.Infof("Невалидный номер заказа: %s", number)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Получим баланс пользователя
	balanceAccount, err := services.GetBalanceAccountByUUID(UUID)
	if err != nil {
		logging.Errorf("Не удалось получить баланс пользователя: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Проверим достаточно ли баллов
	if balanceAccount.Current < wr.Sum {
		logging.Infof("Недостаточно баллов. У пользователя %g, стоимость %g", balanceAccount.Current, wr.Sum)
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	tx, err := db.Storage.Connect.BeginTx(r.Context(), nil)
	if err != nil {
		logging.Infof("Не смогли создать транзакцию: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Записываем в БД списание
	sum := int64(wr.Sum * 100)
	services.CreateWithdraw(UUID, int64(number), sum)

	// Вычитаем из баланса, добавляем в сумму потраченных баллов
	if err := services.CalculateBalanceWithWithdraw(UUID, int(wr.Sum*100)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		logging.Errorf("Ошибка выполнения commit: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
