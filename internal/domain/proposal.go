package domain

import (
	"errors"
	"time"
)

type ProposalStatus string

const (
	StatusDraft     ProposalStatus = "draft"
	StatusReview    ProposalStatus = "review"
	StatusApproved  ProposalStatus = "approved"
	StatusDelivered ProposalStatus = "delivered"
	StatusWithdrawn ProposalStatus = "withdrawn"
	StatusArchived  ProposalStatus = "archived"
)

type ReviewComment struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func TransitionProposal(current, next ProposalStatus) error {
	allowed := map[ProposalStatus]map[ProposalStatus]bool{
		StatusDraft:     {StatusReview: true, StatusWithdrawn: true, StatusArchived: true},
		StatusReview:    {StatusApproved: true, StatusWithdrawn: true, StatusDraft: true},
		StatusApproved:  {StatusDelivered: true, StatusDraft: true},
		StatusDelivered: {},
		StatusWithdrawn: {StatusDraft: true, StatusArchived: true},
		StatusArchived:  {},
	}
	if !allowed[current][next] {
		return errors.New("invalid proposal transition")
	}
	return nil
}
