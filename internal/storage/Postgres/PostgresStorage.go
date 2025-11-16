package Postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	DB  *sql.DB
	Log *slog.Logger
}

func NewPostgresStorage(storagePath string, log *slog.Logger) (*PostgresStorage, error) {
	const op = "internal.storage.Postgres.NewPostgresStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		log.Error(op, ":", "Error opening database")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info(op, ":", "Successfully opened database")
	return &PostgresStorage{DB: db, Log: log}, nil
}

func (s *PostgresStorage) GetDB() *sql.DB {
	return s.DB
}
