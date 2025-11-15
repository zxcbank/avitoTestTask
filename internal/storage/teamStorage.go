package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type TeamStorage struct {
	db *sql.DB
}

func NewTeamStorage(storagePath string) (*TeamStorage, error) {
	const op = "internal.storage.NewTeamStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &TeamStorage{db: db}, nil
}
