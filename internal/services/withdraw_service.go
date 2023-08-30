package services

import (
	"context"
	"database/sql"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/models"
)

func CreateWithdraw(UUID any, number int64, sum int64) error {
	_, err := db.Storage.Connect.ExecContext(context.Background(), "INSERT INTO \"withdrawal\" (uuid, order_number, sum) VALUES ($1, $2, $3);",
		UUID, number, sum)
	if err != nil {
		logging.Infof("Don't create(insert) withdrawal: %s", err)
		return err
	}
	return nil
}

func GetWithdrawByUUID(UUID any) ([]models.Withdrawal, error) {
	items := make([]models.Withdrawal, 0)
	query := "SELECT CAST(order_number AS VARCHAR) as order, CAST(sum AS FLOAT) / CAST(100 AS FLOAT) AS sum, " +
		"processed_at  FROM \"withdrawal\" WHERE uuid = $1 ORDER BY processed_at ASC;"
	rows, err := db.Storage.Connect.QueryContext(context.Background(), query, UUID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var w models.Withdrawal
		err = rows.Scan(&w.Number, &w.Amount, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}

		items = append(items, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
