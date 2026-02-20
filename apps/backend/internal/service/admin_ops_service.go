package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/adminauditlog"
	"github.com/coby/colight/apps/backend/ent/feedback"
	"github.com/google/uuid"
)

// AdminOperationsService handles audit logging and feedback management.
type AdminOperationsService struct {
	db *ent.Client
}

// NewAdminOperationsService creates a new admin operations service.
func NewAdminOperationsService(db *ent.Client) *AdminOperationsService {
	return &AdminOperationsService{db: db}
}

// RecordAuditLog saves an admin action to the audit log.
// Fire-and-forget: logs errors but does not propagate them.
func (s *AdminOperationsService) RecordAuditLog(ctx context.Context, adminID uuid.UUID, action, targetType, targetID string, oldValue, newValue *string) {
	create := s.db.AdminAuditLog.Create().
		SetAdminID(adminID).
		SetAction(action).
		SetTargetType(targetType).
		SetTargetID(targetID)
	if oldValue != nil {
		create = create.SetOldValue(*oldValue)
	}
	if newValue != nil {
		create = create.SetNewValue(*newValue)
	}
	if err := create.Exec(ctx); err != nil {
		slog.Error("record audit log failed", "error", err, "action", action)
	}
}

// ListAuditLogs returns paginated audit log entries with optional action filter.
func (s *AdminOperationsService) ListAuditLogs(ctx context.Context, limit, offset int, actionFilter string) ([]*ent.AdminAuditLog, int, error) {
	query := s.db.AdminAuditLog.Query()
	if actionFilter != "" {
		query = query.Where(adminauditlog.ActionEQ(actionFilter))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	logs, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(adminauditlog.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListFeedbacks returns paginated feedback entries with optional category filter.
func (s *AdminOperationsService) ListFeedbacks(ctx context.Context, limit, offset int, categoryFilter string) ([]*ent.Feedback, int, error) {
	query := s.db.Feedback.Query()
	if categoryFilter != "" {
		query = query.Where(feedback.CategoryEQ(feedback.Category(categoryFilter)))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	feedbacks, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(feedback.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

// UpdateFeedbackStatus updates the admin review status of a feedback entry.
func (s *AdminOperationsService) UpdateFeedbackStatus(ctx context.Context, id, callerID uuid.UUID, status, note string) (*ent.Feedback, error) {
	now := time.Now()

	update := s.db.Feedback.UpdateOneID(id).
		SetAdminStatus(feedback.AdminStatus(status)).
		SetReviewedAt(now).
		SetReviewedBy(callerID)

	if note != "" {
		update = update.SetAdminNote(note)
	}

	fb, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminFeedbackNotFound
		}
		return nil, err
	}

	return fb, nil
}
