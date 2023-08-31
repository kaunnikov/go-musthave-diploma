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

const (
	StatusNew        = 10
	StatusProcessing = 20
	StatusInQueue    = 25 // Промежуточный статус, чтобы повторно не забрать в очередь
	StatusInvalid    = 30
	StatusProcessed  = 40
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
		logging.Errorf("Don't checked is can login: %s", err)
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
		logging.Errorf("Don't create(insert) order: %s", err)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func GetNotProcessedOrderNumbers() ([]int64, error) {
	// Выставляем промежуточный статус, чтобы не забрать повторно раньше времени
	query := "UPDATE \"order\" SET status = $1 WHERE \"status\" IN($2,$3)"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, StatusInQueue, StatusNew, StatusProcessing)
	if err != nil {
		return nil, err
	}

	items := make([]int64, 0)
	query = "SELECT number FROM \"order\" WHERE \"status\" = $1"
	rows, err := db.Storage.Connect.QueryContext(context.Background(), query, StatusInQueue)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var n int64
		err = rows.Scan(&n)
		if err != nil {
			return nil, err
		}

		items = append(items, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func SetProcessingStatusByNumber(number int64) error {
	query := "UPDATE \"order\" SET status = $1 WHERE \"number\" = $2"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, StatusProcessing, number)
	if err != nil {
		return err
	}
	return nil
}

func SetInvalidStatusByNumber(number int64) error {
	query := "UPDATE \"order\" SET status = $1 WHERE \"number\" = $2"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, StatusInvalid, number)
	if err != nil {
		return err
	}
	return nil
}

func SetProcessedStatusByNumber(number int64, accrual int) error {
	query := "UPDATE \"order\" SET status = $1, accrual = $2 WHERE \"number\" = $3"
	_, err := db.Storage.Connect.ExecContext(context.Background(), query, StatusProcessed, accrual, number)
	if err != nil {
		return err
	}
	return nil
}

func getStatusLabelById(statusID int) string {
	switch statusID {
	case StatusNew:
		return "NEW"
	case StatusProcessing:
		return "PROCESSING"
	case StatusInQueue:
		return "PROCESSING"
	case StatusInvalid:
		return "INVALID"
	case StatusProcessed:
		return "PROCESSED"
	}
	return "UNDEFINED"
}
