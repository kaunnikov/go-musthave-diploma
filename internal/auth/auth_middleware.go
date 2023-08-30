package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
	"time"
)

const SecretKey = "UZo57ez$4e2V"
const CookieTokenName = "token"

func CustomAuthMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// получаем токен из куки
		tokenCookie, _ := r.Cookie(CookieTokenName)
		if tokenCookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Достаём токен из куки и расшифровываем
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenCookie.Value, claims,
			func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(SecretKey), nil
			})

		// Если кука не валидная - удаляем старую, выбрасываем 401
		if !token.Valid {
			logging.Infof("Invalid token in cookie: %s", tokenCookie)
			deleteCoolie(w)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			logging.Infof("Token don't decode: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if claims.UUID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Проверяем, что такой пользователь существует
		isExists, err := services.UUIDIfExists(claims.UUID)
		if err != nil {
			logging.Infof("UUID don't check if exists: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !isExists {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Прокидываем UUID в контекст
		ctx := context.WithValue(r.Context(), "uuid", claims.UUID)
		logging.Infof("Достали UUID: %s", claims.UUID)

		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func deleteCoolie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:    CookieTokenName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}

	http.SetCookie(w, c)
}
