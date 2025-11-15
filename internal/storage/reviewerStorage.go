package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type ReviewerStorage struct {
	db *sql.DB
}

func NewReviewerStorage(storagePath string) (*ReviewerStorage, error) {
	const op = "internal.storage.NewReviewerStorage"

	db, err := sql.Open("postgres", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &ReviewerStorage{db: db}, nil
}
