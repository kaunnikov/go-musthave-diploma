package services

import (
	"context"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
)

type BalanceAccountResponseMessage struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func CreateBalanceAccountByUUID(UUID any) error {
	_, err := db.Storage.Connect.ExecContext(context.Background(), "INSERT INTO \"balance\" (uuid) VALUES ($1);", UUID)
	if err != nil {
		logging.Infof("Don't create(insert) balance account: %s", err)
		return err
	}
	return nil
}

func GetBalanceAccountByUUID(UUID any) (*BalanceAccountResponseMessage, error) {
	var b BalanceAccountResponseMessage
	query := "SELECT CAST(current AS FLOAT) / CAST(100 AS FLOAT) as current, CAST(withdrawn AS FLOAT) / CAST(100 AS FLOAT) as withdrawn FROM \"balance\" WHERE \"uuid\"=$1"
	res := db.Storage.Connect.QueryRowContext(context.Background(), query, UUID)
	err := res.Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		logging.Errorf("Don't get balance account: %s", err)
		return nil, err
	}

	return &b, nil
}

func AddAccrualByNumber(number int64, sum int) error {
	query := "UPDATE \"balance\" SET current = current + $1 WHERE uuid = (SELECT uuid FROM \"order\" where number = $2)"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, sum, number)
	if err != nil {
		return err
	}
	return nil
}

func CalculateBalanceWithWithdraw(UUID any, amount int) error {
	query := "UPDATE \"balance\" SET current = current - $1, withdrawn = withdrawn + $1 WHERE uuid = $2"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, amount, UUID)
	if err != nil {
		logging.Errorf("Don't calculate balance: %s", err)
		return err
	}
	return nil
}
