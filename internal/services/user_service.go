package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/models"
)
import "github.com/google/uuid"

const salt = "fvojdn2oip4jfv&#jvldksnvp%^&*(Wnjcd"

func CreateUser(login string, pwd string) (*models.User, error) {
	u := models.User{
		UUID:         uuid.NewString(),
		Login:        login,
		PasswordHash: GenerateHash(pwd),
	}

	err := createUser(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func GenerateHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func LoginIfExists(login string) (bool, error) {
	var isExists bool
	query := "SELECT EXISTS(SELECT * FROM \"user\" WHERE \"login\"=$1)"
	res := db.Storage.Connect.QueryRowContext(context.Background(), query, login)
	err := res.Scan(&isExists)
	if err != nil {
		logging.Errorf("Don't checked login if exists: %s", err)
		return false, err
	}

	return isExists, nil
}

func UUIDIfExists(UUID any) (bool, error) {
	var isExists bool
	query := "SELECT EXISTS(SELECT * FROM \"user\" WHERE \"uuid\"=$1)"
	res := db.Storage.Connect.QueryRowContext(context.Background(), query, UUID)
	err := res.Scan(&isExists)
	if err != nil {
		logging.Errorf("Don't checked UUID if exists: %s", err)
		return false, err
	}

	return isExists, nil
}

func createUser(u *models.User) error {
	_, err := db.Storage.Connect.ExecContext(context.Background(), "INSERT INTO \"user\" (uuid, login, password_hash) VALUES ($1, $2, $3);",
		u.UUID, u.Login, u.PasswordHash)
	if err != nil {
		logging.Errorf("Don't create(insert) user: %s", err)
		return err
	}
	return nil
}

func FindUserByLoginAndPasswordHash(login string, hash string) (string, error) {
	var u string
	query := "SELECT uuid FROM \"user\" WHERE \"login\"=$1 AND \"password_hash\"=$2"
	res := db.Storage.Connect.QueryRowContext(context.Background(), query, login, hash)
	err := res.Scan(&u)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		logging.Errorf("Don't checked is can login: %s", err)
		return "", err
	}

	return u, nil
}
