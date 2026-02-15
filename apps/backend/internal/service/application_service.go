package service

import (
	"context"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/google/uuid"
)

// ApplicationService handles application listing, status updates, and stats for the dashboard.
type ApplicationService struct {
	db *ent.Client
}

// NewApplicationService creates a new application service.
func NewApplicationService(db *ent.Client) *ApplicationService {
	return &ApplicationService{db: db}
}

// ApplicationDetail is the enriched response for dashboard kanban cards.
type ApplicationDetail struct {
	ID               string  `json:"id"`
	CompanyName      string  `json:"company_name"`
	Position         string  `json:"position"`
	Status           string  `json:"status"`
	Deadline         *string `json:"deadline"`
	AppliedAt        *string `json:"applied_at"`
	Notes            string  `json:"notes"`
	CoverLetterCount int     `json:"cover_letter_count"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// ApplicationStats is the summary for the dashboard header.
type ApplicationStats struct {
	Total             int                `json:"total"`
	ByStatus          map[string]int     `json:"by_status"`
	UpcomingDeadlines []ApplicationDetail `json:"upcoming_deadlines"`
}

// ListApplications returns the user's applications with cover letter counts.
func (s *ApplicationService) ListApplications(ctx context.Context, userID uuid.UUID) ([]ApplicationDetail, error) {
	apps, err := s.db.Application.Query().
		Where(application.UserIDEQ(userID)).
		WithCoverLetters().
		Order(ent.Desc(application.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ApplicationDetail, len(apps))
	for i, app := range apps {
		result[i] = toApplicationDetail(app)
	}
	return result, nil
}

// UpdateStatus updates an application's status.
func (s *ApplicationService) UpdateStatus(ctx context.Context, userID uuid.UUID, appID uuid.UUID, newStatus string) (*ApplicationDetail, error) {
	app, err := s.db.Application.Query().
		Where(application.IDEQ(appID)).
		WithCoverLetters().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}

	if app.UserID != userID {
		return nil, ErrApplicationForbidden
	}

	_, err = app.Update().
		SetStatus(application.Status(newStatus)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Re-query with edges for accurate cover letter count
	app, err = s.db.Application.Query().
		Where(application.IDEQ(appID)).
		WithCoverLetters().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	detail := toApplicationDetail(app)
	return &detail, nil
}

// GetStats returns application statistics for the dashboard summary.
func (s *ApplicationService) GetStats(ctx context.Context, userID uuid.UUID) (*ApplicationStats, error) {
	apps, err := s.db.Application.Query().
		Where(application.UserIDEQ(userID)).
		WithCoverLetters().
		All(ctx)
	if err != nil {
		return nil, err
	}

	byStatus := make(map[string]int)
	var upcoming []ApplicationDetail
	now := time.Now()
	threeDaysLater := now.Add(3 * 24 * time.Hour)

	for _, app := range apps {
		byStatus[string(app.Status)]++

		if app.Deadline != nil && !app.Deadline.Before(now) && app.Deadline.Before(threeDaysLater) {
			upcoming = append(upcoming, toApplicationDetail(app))
		}
	}

	return &ApplicationStats{
		Total:             len(apps),
		ByStatus:          byStatus,
		UpcomingDeadlines: upcoming,
	}, nil
}

func toApplicationDetail(app *ent.Application) ApplicationDetail {
	detail := ApplicationDetail{
		ID:          app.ID.String(),
		CompanyName: app.CompanyName,
		Position:    app.Position,
		Status:      string(app.Status),
		Notes:       app.Notes,
		CreatedAt:   app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   app.UpdatedAt.Format(time.RFC3339),
	}

	if app.Deadline != nil {
		d := app.Deadline.Format(time.RFC3339)
		detail.Deadline = &d
	}
	if app.AppliedAt != nil {
		a := app.AppliedAt.Format(time.RFC3339)
		detail.AppliedAt = &a
	}
	if app.Edges.CoverLetters != nil {
		detail.CoverLetterCount = len(app.Edges.CoverLetters)
	}

	return detail
}
