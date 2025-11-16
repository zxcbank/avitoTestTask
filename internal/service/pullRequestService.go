package service

import (
	"avitoTestTask/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
)

type pullRequestStorage interface {
	GetDB() *sql.DB
	CreatePullRequest(PullRequestId, PullRequestName, AuthorID string) (models.PullRequest, error)
	GetPullRequest(PullRequestName string) (*models.PullRequest, error)
	MergePullRequest(PullRequestID string) (*models.PullRequest, error)
	ReassignReviewer(PullRequestID, OldUserId string) (models.Reassign, error)
}

type PullRequestService struct {
	storage pullRequestStorage
	log     *slog.Logger
}

func CreatePullRequestService(storage pullRequestStorage, log *slog.Logger) PullRequestService {
	return PullRequestService{storage: storage, log: log}
}

func (s *PullRequestService) CreatePullRequest(PullRequestId, PullRequestName, AuthorID string) (*models.PullRequest, error) {
	const op = "internal.service.pullRequestService.CreatePullRequest"

	if PullRequestId == "" {
		s.log.Error(op, " : ", "PullRequest ID is empty")
		return &models.PullRequest{}, models.ErrEmptyPullRequestId
	}
	if PullRequestName == "" {
		s.log.Error(op, " : ", "PullRequest name is empty")
		return &models.PullRequest{}, models.ErrEmptyPullRequestName
	}
	if AuthorID == "" {
		s.log.Error(op, " : ", "Author ID is empty")
		return &models.PullRequest{}, models.ErrEmptyPullRequestAutorId
	}

	db := s.storage.GetDB()
	tx, err := db.Begin()
	if err != nil {
		s.log.Error(op, " : ", "Error starting transaction")
		return &models.PullRequest{}, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	pr, err := s.storage.CreatePullRequest(PullRequestId, PullRequestName, AuthorID)
	if err != nil {
		s.log.Error(op, " : ", "Error creating pull request: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return &models.PullRequest{}, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return &models.PullRequest{}, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return &models.PullRequest{}, err
	}

	s.log.Info(op, " : ", "Pull request created",
		"pull_request_id", PullRequestId,
		"pull_request_name", PullRequestName,
		"author_id", AuthorID)
	return &pr, nil
}

func (s *PullRequestService) GetPullRequest(PullRequestName string) (*models.PullRequest, error) {
	const op = "internal.service.pullRequestService.GetPullRequest"

	if PullRequestName == "" {
		s.log.Error(op, " : ", "PullRequest name is empty")
		return nil, models.ErrEmptyPullRequestName
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

	pr, err := s.storage.GetPullRequest(PullRequestName)
	if err != nil {
		s.log.Error(op, " : ", "Error getting pull request: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return nil, err
	}

	s.log.Info(op, " : ", "Pull request retrieved", "pull_request_name", PullRequestName)
	return pr, nil
}

func (s *PullRequestService) MergePullRequest(PullRequestID string) (*models.PullRequest, error) {
	const op = "internal.service.pullRequestService.MergePullRequest"

	if PullRequestID == "" {
		s.log.Error(op, " : ", "PullRequest ID is empty")
		return nil, models.ErrEmptyPullRequestId
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

	pr, err := s.storage.MergePullRequest(PullRequestID)
	if err != nil {
		s.log.Error(op, " : ", "Error merging pull request: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return nil, err
	}

	s.log.Info(op, " : ", "Pull request merged", "pull_request_id", PullRequestID)
	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(PullRequestID, OldUserId string) (*models.Reassign, error) {
	const op = "internal.service.pullRequestService.ReassignReviewer"

	if PullRequestID == "" {
		s.log.Error(op, " : ", "PullRequest ID is empty")
		return &models.Reassign{}, models.ErrEmptyPullRequestId
	}
	if OldUserId == "" {
		s.log.Error(op, " : ", "Old user ID is empty")
		return &models.Reassign{}, models.ErrEmptyOldUserId
	}

	db := s.storage.GetDB()
	tx, err := db.Begin()
	if err != nil {
		s.log.Error(op, " : ", "Error starting transaction")
		return &models.Reassign{}, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	reassign, err := s.storage.ReassignReviewer(PullRequestID, OldUserId)
	if err != nil {
		s.log.Error(op, " : ", "Error reassigning reviewer: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return &models.Reassign{}, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return &models.Reassign{}, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return &models.Reassign{}, err
	}

	s.log.Info(op, " : ", "Reviewer reassigned",
		"pull_request_id", PullRequestID,
		"old_user_id", OldUserId,
		"new_user_id", reassign.NewReviewerID)
	return &reassign, nil
}
