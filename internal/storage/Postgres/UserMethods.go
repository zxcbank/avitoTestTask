package Postgres

import (
	"avitoTestTask/internal/models"
	"database/sql"
	"fmt"
)

func (s *PostgresStorage) SetUserActive(userID string, isActive bool) (*models.User, error) {
	const op = "storage.postgres.SetUserActive"

	if userID == "" {
		return nil, fmt.Errorf("%s: %w", op, models.ErrEmptyUserId)
	}

	userExists, err := s.UserExists(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !userExists {
		return nil, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
	}

	stmt, err := s.DB.Prepare("UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, team_name, is_active")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var user models.User
	err = stmt.QueryRow(isActive, userID).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *PostgresStorage) GetUserReviewPRs(userID string) ([]*models.PullRequest, error) {
	const op = "storage.postgres.GetUserReviewPRs"

	if userID == "" {
		return nil, fmt.Errorf("%s: %w", op, models.ErrEmptyUserId)
	}

	userExists, err := s.UserExists(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !userExists {
		return nil, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
	}

	stmt, err := s.DB.Prepare(`
        SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.user_id = $1 AND pr.status = 'OPEN'
        ORDER BY pr.created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var pullRequests []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var createdAt, mergedAt sql.NullTime

		err = rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &createdAt, &mergedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if createdAt.Valid {
			pr.CreatedAt = (&createdAt).Time.GoString()
		}
		if mergedAt.Valid {
			pr.MergedAt = (&mergedAt).Time.GoString()
		}

		reviewers, err := s.getPRReviewers(pr.PullRequestId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pr.AssignedReviewers = reviewers

		pullRequests = append(pullRequests, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pullRequests, nil
}

func (s *PostgresStorage) getPRReviewers(prID string) ([]string, error) {
	const op = "storage.postgres.getPRReviewers"

	stmt, err := s.DB.Prepare(`
        SELECT user_id FROM pull_request_reviewers WHERE pull_request_id = $1
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(prID)
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return reviewers, nil
}

func (s *PostgresStorage) UserExists(userID string) (bool, error) {
	const op = "storage.postgres.UserExists"

	stmt, err := s.DB.Prepare("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
