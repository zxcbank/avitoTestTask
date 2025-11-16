package Postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"avitoTestTask/internal/models"
	Postgres "avitoTestTask/internal/storage/Postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresStorageTestSuite struct {
	suite.Suite
	db        *sql.DB
	storage   *Postgres.PostgresStorage
	container testcontainers.Container
	ctx       context.Context
	logger    *slog.Logger
}

func (suite *PostgresStorageTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	container, err := postgres.Run(suite.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
	suite.container = container

	connStr, err := container.ConnectionString(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	connStr += " sslmode=disable"

	dataBase, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = dataBase.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	suite.db = dataBase
	suite.storage = &Postgres.PostgresStorage{DB: dataBase, Log: suite.logger}

	err = suite.createTables()
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *PostgresStorageTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *PostgresStorageTestSuite) SetupTest() {
	_, err := suite.db.Exec("DELETE FROM pull_request_reviewers")
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec("DELETE FROM pull_requests")
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec("DELETE FROM users")
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec("DELETE FROM teams")
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *PostgresStorageTestSuite) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS teams (
			team_name VARCHAR(255) PRIMARY KEY
		)`,

		`CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			team_name VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS pull_requests (
			pull_request_id VARCHAR(255) PRIMARY KEY,
			pull_request_name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			merged_at TIMESTAMP NULL,
			FOREIGN KEY (author_id) REFERENCES users(user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS pull_request_reviewers (
			pull_request_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			PRIMARY KEY (pull_request_id, user_id),
			FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name)`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_user_id ON pull_request_reviewers(user_id)`,
	}

	for _, query := range queries {
		_, err := suite.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (suite *PostgresStorageTestSuite) insertTestData() error {
	_, err := suite.db.Exec("INSERT INTO teams (team_name) VALUES ('backend'), ('frontend')")
	if err != nil {
		return err
	}

	_, err = suite.db.Exec(`
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
		('user1', 'User One', 'backend', true),
		('user2', 'User Two', 'backend', true),
		('user3', 'User Three', 'backend', true),
		('user4', 'User Four', 'frontend', true),
		('user5', 'User Five', 'frontend', true)
	`)
	return err
}

func (suite *PostgresStorageTestSuite) TestCreateTeam() {
	t := suite.T()

	team := &models.Team{
		Name: "devops",
		Members: []models.User{
			{
				UserId:   "dev1",
				Username: "DevOps One",
				TeamName: "devops",
				IsActive: true,
			},
			{
				UserId:   "dev2",
				Username: "DevOps Two",
				TeamName: "devops",
				IsActive: true,
			},
		},
	}

	err := suite.storage.CreateTeam(team)

	assert.NoError(t, err)

	var teamName string
	err = suite.db.QueryRow("SELECT team_name FROM teams WHERE team_name = $1", "devops").Scan(&teamName)
	assert.NoError(t, err)
	assert.Equal(t, "devops", teamName)

	var userCount int
	err = suite.db.QueryRow("SELECT COUNT(*) FROM users WHERE team_name = $1", "devops").Scan(&userCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, userCount)
}

func (suite *PostgresStorageTestSuite) TestCreateTeam_EmptyName() {
	t := suite.T()

	team := &models.Team{
		Name:    "",
		Members: []models.User{},
	}

	err := suite.storage.CreateTeam(team)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrEmptyTeamName))
}

func (suite *PostgresStorageTestSuite) TestGetTeam() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	team, err := suite.storage.GetTeam("backend")

	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "backend", team.Name)
	assert.Len(t, team.Members, 3)

	assert.Equal(t, "user1", team.Members[0].UserId)
	assert.Equal(t, "User One", team.Members[0].Username)
	assert.True(t, team.Members[0].IsActive)
}

func (suite *PostgresStorageTestSuite) TestGetTeam_NotFound() {
	t := suite.T()

	team, err := suite.storage.GetTeam("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.True(t, errors.Is(err, models.ErrTeamNotFound))
}

func (suite *PostgresStorageTestSuite) TestCreatePullRequest() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	pr, err := suite.storage.CreatePullRequest("pr1", "Test PR", "user1")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, "pr1", pr.PullRequestId)
	assert.Equal(t, "Test PR", pr.PullRequestName)
	assert.Equal(t, "user1", pr.AuthorId)
	assert.Equal(t, "OPEN", pr.Status)
	assert.NotEmpty(t, pr.CreatedAt)
	assert.Len(t, pr.AssignedReviewers, 2)

	for _, reviewer := range pr.AssignedReviewers {
		assert.NotEqual(t, "user1", reviewer)
		assert.Contains(t, []string{"user2", "user3"}, reviewer)
	}
}

func (suite *PostgresStorageTestSuite) TestCreatePullRequest_EmptyPullRequestId() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("", "Test PR", "user1")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrEmptyPullRequestId))
}

func (suite *PostgresStorageTestSuite) TestCreatePullRequest_UserNotFound() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr1", "Test PR", "nonexistent")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, models.ErrUserNotFound))
}

func (suite *PostgresStorageTestSuite) TestGetPullRequest() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	createdPR, err := suite.storage.CreatePullRequest("pr1", "Test PR", "user1")
	assert.NoError(t, err)

	pr, err := suite.storage.GetPullRequest("pr1")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, createdPR.PullRequestId, pr.PullRequestId)
	assert.Equal(t, createdPR.PullRequestName, pr.PullRequestName)
	assert.Equal(t, createdPR.AuthorId, pr.AuthorId)
	assert.Equal(t, "OPEN", pr.Status)
	assert.Len(t, pr.AssignedReviewers, 2)
}

func (suite *PostgresStorageTestSuite) TestGetPullRequest_NotFound() {
	t := suite.T()

	pr, err := suite.storage.GetPullRequest("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.True(t, errors.Is(err, models.ErrPRNotFound))
}

func (suite *PostgresStorageTestSuite) TestMergePullRequest() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr1", "Test PR", "user1")
	assert.NoError(t, err)

	pr, err := suite.storage.MergePullRequest("pr1")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotEmpty(t, pr.MergedAt)
	assert.Len(t, pr.AssignedReviewers, 2)
}

func (suite *PostgresStorageTestSuite) TestMergePullRequest_NotFound() {
	t := suite.T()

	pr, err := suite.storage.MergePullRequest("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.True(t, errors.Is(err, models.ErrPRNotFound))
}

func (suite *PostgresStorageTestSuite) TestSetUserActive() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	user, err := suite.storage.SetUserActive("user1", false)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user1", user.UserId)
	assert.False(t, user.IsActive)

	var isActive bool
	err = suite.db.QueryRow("SELECT is_active FROM users WHERE user_id = $1", "user1").Scan(&isActive)
	assert.NoError(t, err)
	assert.False(t, isActive)
}

func (suite *PostgresStorageTestSuite) TestSetUserActive_UserNotFound() {
	t := suite.T()

	user, err := suite.storage.SetUserActive("nonexistent", false)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, models.ErrUserNotFound))
}

func (suite *PostgresStorageTestSuite) TestGetUserReviewPRs() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr1", "Test PR 1", "user1")
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr2", "Test PR 2", "user4")
	assert.NoError(t, err)

	prs, err := suite.storage.GetUserReviewPRs("user2")

	assert.NoError(t, err)
	assert.Len(t, prs, 1)
	assert.Equal(t, "pr1", prs[0].PullRequestId)
	assert.Equal(t, "Test PR 1", prs[0].PullRequestName)
	assert.Equal(t, "OPEN", prs[0].Status)
}

func (suite *PostgresStorageTestSuite) TestGetUserReviewPRs_NoPRs() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	prs, err := suite.storage.GetUserReviewPRs("user1")

	assert.NoError(t, err)
	assert.Empty(t, prs)
}

func (suite *PostgresStorageTestSuite) TestReassignReviewer_PRNotFound() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	reassign, err := suite.storage.ReassignReviewer("nonexistent", "user1")

	assert.Error(t, err)
	assert.Equal(t, models.Reassign{}, reassign)
	assert.True(t, errors.Is(err, models.ErrPRNotFound))
}

func (suite *PostgresStorageTestSuite) TestReassignReviewer_NotAssigned() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr1", "Test PR", "user1")
	assert.NoError(t, err)

	reassign, err := suite.storage.ReassignReviewer("pr1", "user4")

	assert.Error(t, err)
	assert.Equal(t, models.Reassign{}, reassign)
	assert.True(t, errors.Is(err, models.ErrNotAssigned))
}

func (suite *PostgresStorageTestSuite) TestUserExists() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	exists, err := suite.storage.UserExists("user1")

	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = suite.storage.UserExists("nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func (suite *PostgresStorageTestSuite) TestPRExists() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	_, err = suite.storage.CreatePullRequest("pr1", "Test PR", "user1")
	assert.NoError(t, err)

	exists, err := suite.storage.PRExists("pr1")

	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = suite.storage.PRExists("nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func (suite *PostgresStorageTestSuite) TestTeamExists() {
	t := suite.T()

	err := suite.insertTestData()
	assert.NoError(t, err)

	exists, err := suite.storage.TeamExists("backend")

	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = suite.storage.TeamExists("nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPostgresStorageTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresStorageTestSuite))
}
