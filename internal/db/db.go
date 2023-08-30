package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"kaunnikov/internal/config"
	"kaunnikov/internal/logging"
	"strings"
)

var Storage DataBaseStorage
var structureDB = []table{
	{
		tableName: "public.user",
		columns: []column{
			{"uuid", "VARCHAR(36) NOT NULL UNIQUE"},
			{"login", "VARCHAR(64) NOT NULL UNIQUE"},
			{"password_hash", "VARCHAR(255) NOT NULL"},
			{"created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"},
		},
	},
	{
		tableName: "public.order",
		columns: []column{
			{"id", "SERIAL PRIMARY KEY"},
			{"uuid", "VARCHAR(36) NOT NULL"},
			{"number", "BIGINT NOT NULL UNIQUE"},
			{"accrual", "INT"},
			{"status", "INT DEFAULT 10 NOT NULL"},
			{"uploaded_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"},
		},
	},
	{
		tableName: "public.balance",
		columns: []column{
			{"id", "SERIAL PRIMARY KEY"},
			{"uuid", "VARCHAR(36) NOT NULL UNIQUE"},
			{"current", "BIGINT DEFAULT 0"},
			{"withdrawn", "BIGINT DEFAULT 0"},
		},
	},
	{
		tableName: "public.withdrawal",
		columns: []column{
			{"id", "SERIAL PRIMARY KEY"},
			{"uuid", "VARCHAR(36) NOT NULL"},
			{"order_number", "BIGINT NOT NULL"},
			{"sum", "BIGINT NOT NULL"},
			{"processed_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"},
		},
	},
}

type DataBaseStorage struct {
	Connect *sql.DB
}

type table struct {
	tableName string
	columns   []column
}
type column struct {
	columnName string
	columnType string
}

// Возвращает строку "<columnName>+<columnType>" для создания таблицы в БД
func (m *table) columnsToString() string {
	var res []string
	for _, c := range m.columns {
		res = append(res, fmt.Sprintf("%s %s", c.columnName, c.columnType))
	}

	return strings.Join(res, ", ")
}

func Init(cfg *config.AppConfig) error {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		logging.Infof("DB don't open: %s", err)
		return err
	}
	Storage = DataBaseStorage{
		Connect: db,
	}

	err = checkTables()
	if err != nil {
		logging.Infof("Don't check tables: %s", err)
		return err
	}
	return nil
}

// checkTables Создаёт таблицы при инициализации БД, если небыли созданы ранее
func checkTables() error {
	var query string
	for _, t := range structureDB {
		query = "CREATE TABLE IF NOT EXISTS " + t.tableName + " (" + t.columnsToString() + ");"

		_, err := Storage.Connect.ExecContext(context.Background(), query)
		if err != nil {
			logging.Infof("table "+t.tableName+" don't created: %s", err)
			return fmt.Errorf("table "+t.tableName+" don't created: %w", err)
		}
	}

	return nil
}
