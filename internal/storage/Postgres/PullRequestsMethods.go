package Postgres

import (
	"avitoTestTask/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func (s *PostgresStorage) CreatePullRequest(PullRequestId, PullRequestName, AuthorID string) (models.PullRequest, error) {
	const op = "storage.postgres.CreatePullRequest"

	if PullRequestId == "" {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrEmptyPullRequestId)
	}
	if PullRequestName == "" {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrEmptyPullRequestName)
	}
	if AuthorID == "" {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrEmptyAuthorId)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	authorExists, err := s.UserExists(AuthorID)
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	if !authorExists {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
	}

	prExists, err := s.PRExists(PullRequestId)
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	if prExists {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrPRExists)
	}

	prStmt, err := tx.Prepare(`
        INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id, status) 
        VALUES($1, $2, $3, 'OPEN') 
        RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
    `)
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	defer prStmt.Close()

	var pr models.PullRequest
	var createdAt, mergedAt sql.NullTime
	err = prStmt.QueryRow(PullRequestId, PullRequestName, AuthorID).Scan(
		&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &createdAt, &mergedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.PullRequest{}, fmt.Errorf("%s: %w", op, models.ErrPRExists)
		}
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}

	if createdAt.Valid {
		pr.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if mergedAt.Valid {
		pr.MergedAt = mergedAt.Time.Format(time.RFC3339)
	}

	reviewers, err := s.assignReviewers(tx, AuthorID, PullRequestId)
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	pr.AssignedReviewers = reviewers

	if err = tx.Commit(); err != nil {
		return models.PullRequest{}, fmt.Errorf("%s: %w", op, err)
	}

	return pr, nil
}

func (s *PostgresStorage) GetPullRequest(PullRequestID string) (*models.PullRequest, error) {
	const op = "storage.postgres.GetPullRequest"

	if PullRequestID == "" {
		return nil, fmt.Errorf("%s: %w", op, models.ErrEmptyPullRequestId)
	}

	exists, err := s.PRExists(PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, models.ErrPRNotFound)
	}

	stmt, err := s.DB.Prepare(`
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests 
        WHERE pull_request_id = $1
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var pr models.PullRequest
	var createdAt, mergedAt sql.NullTime
	err = stmt.QueryRow(PullRequestID).Scan(
		&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &createdAt, &mergedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, models.ErrPRNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if createdAt.Valid {
		pr.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if mergedAt.Valid {
		pr.MergedAt = mergedAt.Time.Format(time.RFC3339)
	}

	reviewers, err := s.getPRReviewers(PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (s *PostgresStorage) MergePullRequest(PullRequestID string) (*models.PullRequest, error) {
	const op = "storage.postgres.MergePullRequest"

	if PullRequestID == "" {
		return nil, fmt.Errorf("%s: %w", op, models.ErrEmptyPullRequestId)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	exists, err := s.PRExists(PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, models.ErrPRNotFound)
	}

	stmt, err := tx.Prepare(`
        UPDATE pull_requests 
        SET status = 'MERGED', merged_at = CURRENT_TIMESTAMP 
        WHERE pull_request_id = $1 AND status != 'MERGED'
        RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var pr models.PullRequest
	var createdAt, mergedAt sql.NullTime
	err = stmt.QueryRow(PullRequestID).Scan(
		&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &createdAt, &mergedAt,
	)

	if err == sql.ErrNoRows {
		getStmt, err := tx.Prepare(`
            SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
            FROM pull_requests 
            WHERE pull_request_id = $1
        `)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		defer getStmt.Close()

		err = getStmt.QueryRow(PullRequestID).Scan(
			&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &createdAt, &mergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if createdAt.Valid {
		pr.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if mergedAt.Valid {
		pr.MergedAt = mergedAt.Time.Format(time.RFC3339)
	}

	reviewers, err := s.getPRReviewers(PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	pr.AssignedReviewers = reviewers

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &pr, nil
}

func (s *PostgresStorage) ReassignReviewer(PullRequestID, OldUserId string) (models.Reassign, error) {
	const op = "storage.postgres.ReassignReviewer"

	if PullRequestID == "" {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrEmptyPullRequestId)
	}
	if OldUserId == "" {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrEmptyUserId)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	pr, err := s.GetPullRequest(PullRequestID)
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	if pr.Status == "MERGED" {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrPRMerged)
	}

	isAssigned := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == OldUserId {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrNotAssigned)
	}

	var oldUserTeam string
	teamStmt, err := tx.Prepare("SELECT team_name FROM users WHERE user_id = $1")
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}
	defer teamStmt.Close()

	err = teamStmt.QueryRow(OldUserId).Scan(&oldUserTeam)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
		}
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	newReviewerStmt, err := tx.Prepare(`
        SELECT u.user_id 
        FROM users u
        WHERE u.team_name = $1 
        AND u.is_active = true 
        AND u.user_id != $2 
        AND u.user_id != $3 
        AND u.user_id NOT IN (
            SELECT prr.user_id 
            FROM pull_request_reviewers prr 
            WHERE prr.pull_request_id = $4
        )
        ORDER BY RANDOM()
        LIMIT 1
    `)
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}
	defer newReviewerStmt.Close()

	var newReviewerID string
	err = newReviewerStmt.QueryRow(oldUserTeam, OldUserId, pr.AuthorId, PullRequestID).Scan(&newReviewerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Reassign{}, fmt.Errorf("%s: %w", op, models.ErrNoCandidate)
		}
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	deleteStmt, err := tx.Prepare("DELETE FROM pull_request_reviewers WHERE pull_request_id = $1 AND user_id = $2")
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}
	defer deleteStmt.Close()

	_, err = deleteStmt.Exec(PullRequestID, OldUserId)
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	insertStmt, err := tx.Prepare("INSERT INTO pull_request_reviewers(pull_request_id, user_id) VALUES($1, $2)")
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(PullRequestID, newReviewerID)
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	updatedPR, err := s.GetPullRequest(PullRequestID)
	if err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return models.Reassign{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Reassign{
		PR:            *updatedPR,
		NewReviewerID: newReviewerID,
	}, nil
}

func (s *PostgresStorage) PRExists(prID string) (bool, error) {
	const op = "storage.postgres.PRExists"

	stmt, err := s.DB.Prepare("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *PostgresStorage) assignReviewers(tx *sql.Tx, authorID, prID string) ([]string, error) {
	const op = "storage.postgres.assignReviewers"

	var authorTeam string
	teamStmt, err := tx.Prepare("SELECT team_name FROM users WHERE user_id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer teamStmt.Close()

	err = teamStmt.QueryRow(authorID).Scan(&authorTeam)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reviewerStmt, err := tx.Prepare(`
        SELECT user_id 
        FROM users 
        WHERE team_name = $1 
        AND is_active = true 
        AND user_id != $2 
        ORDER BY RANDOM() 
        LIMIT 2
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer reviewerStmt.Close()

	rows, err := reviewerStmt.Query(authorTeam, authorID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		err = rows.Scan(&reviewerID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	insertStmt, err := tx.Prepare("INSERT INTO pull_request_reviewers(pull_request_id, user_id) VALUES($1, $2)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer insertStmt.Close()

	for _, reviewer := range reviewers {
		_, err = insertStmt.Exec(prID, reviewer)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return reviewers, nil
}
