package Postgres

import (
	"avitoTestTask/internal/models"
	"fmt"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (s *PostgresStorage) CreateTeam(team *models.Team) error {
	const op = "storage.postgres.CreateTeam"

	if team.Name == "" {
		return fmt.Errorf("%s: %w", op, models.ErrEmptyTeamName)
	}

	teamStmt, err := s.DB.Prepare("INSERT INTO teams(team_name) VALUES($1)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer teamStmt.Close()

	_, err = teamStmt.Exec(team.Name)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, models.ErrTeamExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	userStmt, err := s.DB.Prepare("INSERT INTO users(user_id, username, team_name, is_active) VALUES($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer userStmt.Close()

	for _, member := range team.Members {
		_, err = userStmt.Exec(member.UserId, member.Username, team.Name, member.IsActive)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				updateStmt, err := s.DB.Prepare("UPDATE users SET team_name = $1 WHERE user_id = $2")
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				defer updateStmt.Close()

				_, err = updateStmt.Exec(team.Name, member)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			} else {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
	return nil
}

func (s *PostgresStorage) GetTeam(teamName string) (*models.Team, error) {
	const op = "storage.postgres.GetTeam"

	if teamName == "" {
		return nil, fmt.Errorf("%s: %w", op, models.ErrEmptyTeamName)
	}

	exists, err := s.TeamExists(teamName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, models.ErrTeamNotFound)
	}

	stmt, err := s.DB.Prepare("SELECT * FROM users WHERE team_name = $1 AND is_active = true")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(teamName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)

		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		members = append(members, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (s *PostgresStorage) TeamExists(teamName string) (bool, error) {
	const op = "storage.postgres.TeamExists"

	if teamName == "" {
		return false, fmt.Errorf("%s: %w", op, models.ErrEmptyTeamName)
	}

	stmt, err := s.DB.Prepare("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

//func (s *PostgresStorage) GetTeamMembers(teamName string) ([]models.User, error) {
//	const op = "storage.postgres.GetTeamMembers"
//
//	stmt, err := s.DB.Prepare("SELECT user_id, username, is_active FROM users WHERE team_name = $1")
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query(teamName)
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//	defer rows.Close()
//
//	var users []models.User
//	for rows.Next() {
//		var user models.User
//		err = rows.Scan(&user.UserId, &user.Username, &user.IsActive)
//		if err != nil {
//			return nil, fmt.Errorf("%s: %w", op, err)
//		}
//		users = append(users, user)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//
//	return users, nil
//}

//
//func (s *PostgresStorage) UpdateUserTeam(userID, teamName string) error {
//	const op = "storage.postgres.UpdateUserTeam"
//
//	stmt, err := s.DB.Prepare("UPDATE users SET team_name = $1 WHERE user_id = $2")
//	if err != nil {
//		return fmt.Errorf("%s: %w", op, err)
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(teamName, userID)
//	if err != nil {
//		return fmt.Errorf("%s: %w", op, err)
//	}
//
//	return nil
//}
