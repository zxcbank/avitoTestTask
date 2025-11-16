package service

import (
	"avitoTestTask/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
)

type userStorage interface {
	GetDB() *sql.DB

	SetUserActive(userID string, isActive bool) (*models.User, error)
	GetUserReviewPRs(userID string) ([]*models.PullRequest, error)
}

type UserService struct {
	storage userStorage
	log     *slog.Logger
}

func CreateUserService(storage userStorage, log *slog.Logger) UserService {
	return UserService{storage: storage, log: log}
}

func (s *UserService) SetUserActive(userId string, isActive bool) (*models.User, error) {
	const op = "internal.service.SetUserActive"

	if userId == "" {
		s.log.Error(op, " : ", "User ID is empty")
		return nil, models.ErrEmptyUserId
	}

	db := s.storage.GetDB()
	tx, err := db.Begin()
	if err != nil {
		s.log.Error(op, " : ", "Error starting transaction")
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	user, err := s.storage.SetUserActive(userId, isActive)
	if err != nil {
		s.log.Error(op, " : ", "Error setting user active: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return nil, err
	}

	s.log.Info(op, " : ", "User activity updated", "user_id", userId, "is_active", isActive)
	return user, nil
}

func (s *UserService) GetUserReviewPRs(userId string) ([]*models.PullRequest, error) {
	const op = "internal.service.GetUserReviewPRs"

	if userId == "" {
		s.log.Error(op, " : ", "User ID is empty")
		return nil, models.ErrEmptyUserId
	}

	db := s.storage.GetDB()
	tx, err := db.Begin()
	if err != nil {
		s.log.Error(op, " : ", "Error starting transaction")
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED")
	if err != nil {
		tx.Rollback()
		s.log.Error(op, " : ", "Error setting transaction isolation level: ", err)
		return nil, fmt.Errorf("set isolation level: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	prs, err := s.storage.GetUserReviewPRs(userId)
	if err != nil {
		s.log.Error(op, " : ", "Error getting user review PRs: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return nil, err
	}

	s.log.Info(op, " : ", "Retrieved review PRs for user", "user_id", userId, "prs_count", len(prs))
	return prs, nil
}
