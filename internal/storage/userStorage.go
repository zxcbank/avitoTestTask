package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(storagePath string) (*UserStorage, error) {
	const op = "internal.storage.NewUserStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &UserStorage{db: db}, nil
}
