package models

import "errors"

var (
	ErrTeamExists    = errors.New("team already exists")
	ErrTeamNotFound  = errors.New("team not found")
	ErrEmptyTeamName = errors.New("empty team name")
	ErrUserNotFound  = errors.New("user not found")
	ErrPRNotFound    = errors.New("pr not found")
	ErrPRExists      = errors.New("pr already exists")
	ErrPRMerged      = errors.New("pr merged")
	ErrNotAssigned   = errors.New("not assigned")
	ErrNoCandidate   = errors.New("no candidate")
)
