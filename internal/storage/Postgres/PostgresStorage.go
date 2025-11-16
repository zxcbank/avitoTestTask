package Postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	DB *sql.DB
}

func NewPostgresStorage(storagePath string) (*PostgresStorage, error) {
	const op = "internal.storage.NewPostgresStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PostgresStorage{DB: db}, nil
}

func (s *PostgresStorage) GetDB() *sql.DB {
	return s.DB
}
