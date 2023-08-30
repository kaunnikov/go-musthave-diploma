package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"kaunnikov/internal/logging"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Token string
	UUID  string
}

func SetCookie(w http.ResponseWriter, UUID string) error {
	tokenCookie, err := generateCookie(UUID)
	if err != nil {
		return fmt.Errorf("don't create token: %s", err)
	}
	http.SetCookie(w, tokenCookie)
	return nil
}

func generateCookie(UUID string) (*http.Cookie, error) {
	token, err := generateJWTString(UUID)
	if err != nil {
		logging.Infof("Don't create JWT: %s", err)
		return nil, err
	}

	return &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	}, nil
}

func generateJWTString(UUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(31 * 24 * time.Hour)),
		},
		Token: uuid.NewString(),
		UUID:  UUID,
	})
	logging.Infof("Кладём в куку UUID: %s", UUID)

	return token.SignedString([]byte(SecretKey))
}
