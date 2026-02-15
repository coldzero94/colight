package service

import (
	"context"
	"errors"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/feedback"
	"github.com/google/uuid"
)

var ErrEmptyFeedbackContent = errors.New("feedback content is required")

// FeedbackInput represents the input for submitting feedback.
type FeedbackInput struct {
	Category  string `json:"category"`
	Content   string `json:"content"`
	PageURL   string `json:"page_url,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// FeedbackService handles feedback submissions.
type FeedbackService struct {
	db *ent.Client
}

// NewFeedbackService creates a new feedback service.
func NewFeedbackService(db *ent.Client) *FeedbackService {
	return &FeedbackService{db: db}
}

// Submit creates a new feedback record.
func (s *FeedbackService) Submit(ctx context.Context, userID uuid.UUID, input FeedbackInput) (*ent.Feedback, error) {
	if input.Content == "" {
		return nil, ErrEmptyFeedbackContent
	}

	builder := s.db.Feedback.Create().
		SetUserID(userID).
		SetCategory(feedback.Category(input.Category)).
		SetContent(input.Content)

	if input.PageURL != "" {
		builder = builder.SetPageURL(input.PageURL)
	}
	if input.UserAgent != "" {
		builder = builder.SetUserAgent(input.UserAgent)
	}

	return builder.Save(ctx)
}
