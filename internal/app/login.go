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

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (m *loginRequest) validate() error {
	m.Login = strings.TrimSpace(m.Login)
	m.Password = strings.TrimSpace(m.Password)
	if m.Login == "" || m.Password == "" {
		logging.Infof("empty login or password %s:%s", m.Login, m.Password)
		return fmt.Errorf(fmt.Sprintf("empty login or password %s", ""))
	}

	return nil
}

func (m *app) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	var t loginRequest
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

	// Проверяем логин и пароль
	hash := services.GenerateHash(t.Password)
	uuid, err := services.FindUserByLoginAndPasswordHash(t.Login, hash)
	if err != nil {
		logging.Errorf("cannot check exists user for login: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	// Если пользователь с таким логином или паролем не найден - возвращаем 401
	if uuid == "" {
		logging.Errorf("user don't logged: %s", t.Login)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Записываем в куку JWT
	if err = auth.SetCookie(w, uuid); err != nil {
		logging.Errorf("Don't create cookie: %s", err)
		http.Error(w, "Error in server!", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
