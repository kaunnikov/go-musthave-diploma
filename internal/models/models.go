package models

import (
	"database/sql"
	"time"
)

type User struct {
	UUID         string `json:"uuid"`
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}

type Order struct {
	Id         int
	UUID       string        `json:"uuid"`
	Number     int           `json:"number"`
	Accrual    sql.NullInt64 `json:"accrual,omitempty"`
	Status     int           `json:"status"`
	UploadedAt time.Time     `json:"uploaded_at"`
}

type Withdrawal struct {
	Number      string    `json:"order"`
	Amount      float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
