package models

type Reassign struct {
	PR            PullRequest `json:"PR"`
	NewReviewerID string      `json:"NewReviewerId"`
}
