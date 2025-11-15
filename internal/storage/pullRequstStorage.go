package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PullRequestStorage struct {
	db *sql.DB
}

func NewPullRequestStorage(storagePath string) (*PullRequestStorage, error) {
	const op = "internal.storage.NewPullRequestStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PullRequestStorage{db: db}, nil
}
