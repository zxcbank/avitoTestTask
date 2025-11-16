package models

import "errors"

var (
	ErrTeamExists    = errors.New("team already exists")
	ErrTeamNotFound  = errors.New("team not found")
	ErrEmptyTeamName = errors.New("empty team name")

	ErrEmptyUserId  = errors.New("empty user id")
	ErrUserNotFound = errors.New("user not found")

	ErrEmptyPullRequestId      = errors.New("empty Pull Request Id")
	ErrEmptyOldUserId          = errors.New("empty old user id")
	ErrEmptyPullRequestName    = errors.New("empty Pull Request Name")
	ErrEmptyPullRequestAutorId = errors.New("empty Pull Request AutorId")
	ErrEmptyAuthorId           = errors.New("empty Author Id")

	ErrPRNotFound  = errors.New("pr not found")
	ErrPRExists    = errors.New("pr already exists")
	ErrPRMerged    = errors.New("pr merged")
	ErrNotAssigned = errors.New("not assigned")
	ErrNoCandidate = errors.New("no candidate")
)
