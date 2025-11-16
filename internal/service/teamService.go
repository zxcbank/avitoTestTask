package service

import (
	"avitoTestTask/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

type teamStorage interface {
	GetDB() *sql.DB

	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)
}

type TeamService struct {
	storage teamStorage
	log     *slog.Logger
}

func CreateTeamService(storage teamStorage, log *slog.Logger) TeamService {
	return TeamService{storage: storage, log: log}
}

func (s *TeamService) CreateTeam(team *models.Team) error {
	const op = "internal.service.CreateTeam"
	if team == nil {
		s.log.Error(op, " : ", "Team is nil")
		return errors.New("nil team")
	}
	if team.Name == "" {
		s.log.Error(op, " : ", "Team name is empty")
		return errors.New("empty team name")
	}

	db := s.storage.GetDB()
	tx, err := db.Begin()

	if err != nil {
		s.log.Error(op, " : ", "Error starting transaction")
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	err = s.storage.CreateTeam(team)

	if err != nil {
		s.log.Error(op, " : ", "Error creating team: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return err
	}

	s.log.Info(op, " : ", "Team created", team.Name)
	return nil
}

func (s *TeamService) GetTeam(teamName string) (*models.Team, error) {
	const op = "internal.service.GetTeam"

	if teamName == "" {
		s.log.Error(op, " : ", "teamName is empty")
		return nil, models.ErrTeamNotFound
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

	team, err := s.storage.GetTeam(teamName)

	if err != nil {
		s.log.Error(op, " : ", "Error getting team: ", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%v : rollback error: %v, original error: %w", op, rbErr, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.log.Error(op, " : ", "commit error: ", err)
		return nil, err
	}

	s.log.Info(op, " : ", "Team Founded", team.Name)
	return team, nil
}
