package services

import (
	"context"
	"database/sql"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/models"
	"strconv"
	"time"
)

type UserOrdersResponseMessage struct {
	number          int
	NumberResponse  string `json:"number"`
	accrual         sql.NullInt64
	AccrualResponse float64 `json:"accrual,omitempty"`
	status          int
	StatusResponse  string    `json:"status"`
	UploadedAt      time.Time `json:"uploaded_at"`
}

func (m *UserOrdersResponseMessage) prepareData() {
	m.NumberResponse = strconv.Itoa(m.number)
	if m.accrual.Int64 > 0 {
		m.AccrualResponse = float64(m.accrual.Int64) / float64(100)
	}
	m.StatusResponse = getStatusLabelById(m.status)

}

func GetOrderByNumber(n int) (*models.Order, error) {
	var (
		id         int
		UUID       string
		number     int
		accrual    sql.NullInt64
		status     int
		uploadedAt time.Time
	)

	query := "SELECT id, uuid, number, accrual, status,uploaded_at FROM \"order\" WHERE \"number\"=$1"
	res := db.Storage.Connect.QueryRowContext(context.Background(), query, n)
	err := res.Scan(&id, &UUID, &number, &accrual, &status, &uploadedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logging.Infof("Don't checked is can login: %s", err)
		return nil, err
	}

	order := models.Order{
		Id:         id,
		UUID:       UUID,
		Number:     number,
		Accrual:    accrual,
		Status:     status,
		UploadedAt: uploadedAt,
	}
	return &order, nil
}

func CreateOrder(n int, UUID any) error {
	_, err := db.Storage.Connect.ExecContext(context.Background(), "INSERT INTO \"order\" (uuid, number) VALUES ($1, $2);",
		UUID, n)
	if err != nil {
		logging.Infof("Don't create(insert) order: %s", err)
		return err
	}
	return nil
}

func GetOrdersByUUID(UUID any) ([]UserOrdersResponseMessage, error) {

	items := make([]UserOrdersResponseMessage, 0)
	query := "SELECT number, accrual, status,uploaded_at FROM \"order\" WHERE \"uuid\"=$1"
	rows, err := db.Storage.Connect.QueryContext(context.Background(), query, UUID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var o UserOrdersResponseMessage
		err = rows.Scan(&o.number, &o.accrual, &o.status, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		o.prepareData()

		items = append(items, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func getStatusLabelById(statusID int) string {
	switch statusID {
	case 10:
		return "NEW"
	case 20:
		return "PROCESSING"
	case 30:
		return "INVALID"
	case 40:
		return "PROCESSED"
	}
	return "UNDEFINED"
}
