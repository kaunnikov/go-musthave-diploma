package app

import (
	"encoding/json"
	"fmt"
	"io"
	"kaunnikov/internal/auth"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
	"strings"
)

type registrationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (m *registrationRequest) validate() error {
	m.Login = strings.TrimSpace(m.Login)
	m.Password = strings.TrimSpace(m.Password)
	if m.Login == "" || m.Password == "" {
		logging.Infof("empty login or password %s:%s", m.Login, m.Password)
		return fmt.Errorf(fmt.Sprintf("empty login or password %s", ""))
	}

	return nil
}

func (m *app) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
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

	var t registrationRequest
	err = json.Unmarshal(body, &t)
	if err != nil {
		logging.Errorf("cannot decode request body to `JSON`: %s", err)
		http.Error(w, fmt.Sprintf("cannot decode request body to `JSON`: %s", err), http.StatusBadRequest)
		return
	}

	if err := t.validate(); err != nil {
		http.Error(w, fmt.Sprintf("empty login or password %s", ""), http.StatusBadRequest)
		return
	}

	// Проверяем логин на уникальность
	ifExists, err := services.LoginIfExists(t.Login)
	if err != nil {
		logging.Errorf("cannot check exists user for login: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	// Если логин есть в БД - возвращаем 409
	if ifExists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Создаём нового пользователя
	u, err := services.CreateUser(t.Login, t.Password)
	if err != nil {
		logging.Errorf("cannot insert User in DB: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	// Создаём запись в таблице Баланса
	if err := services.CreateBalanceAccountByUUID(u.UUID); err != nil {
		logging.Errorf("Don't create balance account: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	// Записываем в куку JWT
	if err = auth.SetCookie(w, u.UUID); err != nil {
		logging.Errorf("Don't create cookie: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
