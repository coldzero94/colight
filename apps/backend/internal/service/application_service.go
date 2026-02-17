package service

import (
	"context"
	"errors"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/ent/companyanalysis"
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
	ID               string   `json:"id"`
	CompanyName      string   `json:"company_name"`
	Position         string   `json:"position"`
	Status           string   `json:"status"`
	Deadline         *string  `json:"deadline"`
	AppliedAt        *string  `json:"applied_at"`
	Notes            string   `json:"notes"`
	Tags             []string `json:"tags"`
	AnalysisID       *string  `json:"analysis_id,omitempty"`
	CoverLetterCount int      `json:"cover_letter_count"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// CreateApplicationInput holds params for creating a manual application.
type CreateApplicationInput struct {
	CompanyName string     `json:"company_name"`
	Position    string     `json:"position"`
	JobURL      string     `json:"job_url"`
	Deadline    *time.Time `json:"deadline"`
	Notes       string     `json:"notes"`
	Tags        []string   `json:"tags"`
}

// UpdateApplicationInput holds params for editing an application (all optional).
type UpdateApplicationInput struct {
	CompanyName *string    `json:"company_name"`
	Position    *string    `json:"position"`
	JobURL      *string    `json:"job_url"`
	Deadline    *time.Time `json:"deadline"`
	AppliedAt   *time.Time `json:"applied_at"`
	Notes       *string    `json:"notes"`
	Tags        *[]string  `json:"tags"`
}

// SearchApplicationsInput holds filter params for searching applications.
type SearchApplicationsInput struct {
	Query        string     `json:"query"`
	Tags         []string   `json:"tags"`
	DeadlineFrom *time.Time `json:"deadline_from"`
	DeadlineTo   *time.Time `json:"deadline_to"`
}

// ApplicationStats is the summary for the dashboard header.
type ApplicationStats struct {
	Total             int                 `json:"total"`
	ByStatus          map[string]int      `json:"by_status"`
	UpcomingDeadlines []ApplicationDetail `json:"upcoming_deadlines"`
}

// CreateApplication creates a new application manually with status "preparing".
func (s *ApplicationService) CreateApplication(ctx context.Context, userID uuid.UUID, input CreateApplicationInput) (*ApplicationDetail, error) {
	if input.CompanyName == "" {
		return nil, errors.New("company_name is required")
	}

	builder := s.db.Application.Create().
		SetUserID(userID).
		SetCompanyName(input.CompanyName).
		SetStatus(application.StatusPreparing)

	if input.Position != "" {
		builder = builder.SetPosition(input.Position)
	}
	if input.JobURL != "" {
		builder = builder.SetJobURL(input.JobURL)
	}
	if input.Deadline != nil {
		builder = builder.SetDeadline(*input.Deadline)
	}
	if input.Notes != "" {
		builder = builder.SetNotes(input.Notes)
	}
	if input.Tags != nil {
		builder = builder.SetTags(input.Tags)
	}

	app, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	detail := toApplicationDetail(app)
	return &detail, nil
}

// UpdateApplication updates an application's fields (not status).
func (s *ApplicationService) UpdateApplication(ctx context.Context, userID uuid.UUID, appID uuid.UUID, input UpdateApplicationInput) (*ApplicationDetail, error) {
	app, err := s.db.Application.Query().
		Where(application.IDEQ(appID)).
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

	updater := app.Update()
	if input.CompanyName != nil {
		updater = updater.SetCompanyName(*input.CompanyName)
	}
	if input.Position != nil {
		updater = updater.SetPosition(*input.Position)
	}
	if input.JobURL != nil {
		updater = updater.SetJobURL(*input.JobURL)
	}
	if input.Deadline != nil {
		updater = updater.SetDeadline(*input.Deadline)
	}
	if input.AppliedAt != nil {
		updater = updater.SetAppliedAt(*input.AppliedAt)
	}
	if input.Notes != nil {
		updater = updater.SetNotes(*input.Notes)
	}
	if input.Tags != nil {
		updater = updater.SetTags(*input.Tags)
	}

	_, err = updater.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Re-query with edges
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

// DeleteApplication hard-deletes an application. Related cover letters have FK set to NULL.
func (s *ApplicationService) DeleteApplication(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	app, err := s.db.Application.Query().
		Where(application.IDEQ(appID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrApplicationNotFound
		}
		return err
	}

	if app.UserID != userID {
		return ErrApplicationForbidden
	}

	return s.db.Application.DeleteOneID(appID).Exec(ctx)
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

// SearchApplications filters applications by text query, tags, and deadline range.
func (s *ApplicationService) SearchApplications(ctx context.Context, userID uuid.UUID, input SearchApplicationsInput) ([]ApplicationDetail, error) {
	query := s.db.Application.Query().
		Where(application.UserIDEQ(userID)).
		WithCoverLetters().
		WithAnalysis().
		Order(ent.Desc(application.FieldCreatedAt))

	if input.Query != "" {
		query = query.Where(
			application.Or(
				application.CompanyNameContainsFold(input.Query),
				application.PositionContainsFold(input.Query),
			),
		)
	}

	if input.DeadlineFrom != nil {
		query = query.Where(application.DeadlineGTE(*input.DeadlineFrom))
	}
	if input.DeadlineTo != nil {
		query = query.Where(application.DeadlineLTE(*input.DeadlineTo))
	}

	apps, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	// Tag filtering in memory (works with both SQLite and PostgreSQL)
	if len(input.Tags) > 0 {
		filtered := make([]*ent.Application, 0, len(apps))
		for _, app := range apps {
			if matchesTags(app.Tags, input.Tags) {
				filtered = append(filtered, app)
			}
		}
		apps = filtered
	}

	result := make([]ApplicationDetail, len(apps))
	for i, app := range apps {
		result[i] = toApplicationDetail(app)
	}
	return result, nil
}

// LinkAnalysis links a company analysis to an application. Verifies ownership of both.
func (s *ApplicationService) LinkAnalysis(ctx context.Context, userID uuid.UUID, appID uuid.UUID, analysisID uuid.UUID) (*ApplicationDetail, error) {
	app, err := s.db.Application.Query().
		Where(application.IDEQ(appID)).
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

	analysis, err := s.db.CompanyAnalysis.Query().
		Where(companyanalysis.IDEQ(analysisID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAnalysisNotFound
		}
		return nil, err
	}

	if analysis.UserID != userID {
		return nil, ErrAnalysisForbidden
	}

	_, err = app.Update().
		SetAnalysisID(analysisID).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Re-query with edges
	app, err = s.db.Application.Query().
		Where(application.IDEQ(appID)).
		WithCoverLetters().
		WithAnalysis().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	detail := toApplicationDetail(app)
	return &detail, nil
}

// matchesTags returns true if the app has at least one of the requested tags.
func matchesTags(appTags []string, filterTags []string) bool {
	tagSet := make(map[string]struct{}, len(filterTags))
	for _, t := range filterTags {
		tagSet[t] = struct{}{}
	}
	for _, t := range appTags {
		if _, ok := tagSet[t]; ok {
			return true
		}
	}
	return false
}

func toApplicationDetail(app *ent.Application) ApplicationDetail {
	detail := ApplicationDetail{
		ID:          app.ID.String(),
		CompanyName: app.CompanyName,
		Position:    app.Position,
		Status:      string(app.Status),
		Notes:       app.Notes,
		Tags:        app.Tags,
		CreatedAt:   app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   app.UpdatedAt.Format(time.RFC3339),
	}

	if detail.Tags == nil {
		detail.Tags = []string{}
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
	if app.Edges.Analysis != nil {
		id := app.Edges.Analysis.ID.String()
		detail.AnalysisID = &id
	}

	return detail
}
