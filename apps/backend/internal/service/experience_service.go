package service

import (
	"context"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/experiencetag"
	"github.com/coby/colight/apps/backend/ent/experienceusage"
	"github.com/coby/colight/apps/backend/ent/experienceweapon"
	"github.com/google/uuid"
)

type ExperienceService struct {
	db *ent.Client
}

type CreateExperienceInput struct {
	Title         string
	Category      string
	PeriodStart   *time.Time
	PeriodEnd     *time.Time
	Role          string
	Content       string
	Result        string
	StarSituation string
	StarTask      string
	StarAction    string
	StarResult    string
	Keywords      []string
	Source        string
}

type UpdateExperienceInput struct {
	Title         *string
	Category      *string
	PeriodStart   *time.Time
	PeriodEnd     *time.Time
	Role          *string
	Content       *string
	Result        *string
	StarSituation *string
	StarTask      *string
	StarAction    *string
	StarResult    *string
	Keywords      []string
}

type ExperienceListParams struct {
	Sort       string // "latest" (default), "oldest", "title"
	Category   string
	WeaponCode string // Filter by weapon code (W01-W07)
}

func NewExperienceService(db *ent.Client) *ExperienceService {
	return &ExperienceService{db: db}
}

func (s *ExperienceService) CreateExperience(ctx context.Context, userID uuid.UUID, input CreateExperienceInput) (*ent.Experience, error) {
	if err := validateTitle(input.Title); err != nil {
		return nil, err
	}
	if err := validatePeriod(input.PeriodStart, input.PeriodEnd); err != nil {
		return nil, err
	}

	builder := s.db.Experience.Create().
		SetUserID(userID).
		SetTitle(input.Title).
		SetContent(input.Content)

	if input.Category != "" {
		builder = builder.SetCategory(input.Category)
	}
	if input.PeriodStart != nil {
		builder = builder.SetPeriodStart(*input.PeriodStart)
	}
	if input.PeriodEnd != nil {
		builder = builder.SetPeriodEnd(*input.PeriodEnd)
	}
	if input.Role != "" {
		builder = builder.SetRole(input.Role)
	}
	if input.Result != "" {
		builder = builder.SetResult(input.Result)
	}
	if input.StarSituation != "" {
		builder = builder.SetStarSituation(input.StarSituation)
	}
	if input.StarTask != "" {
		builder = builder.SetStarTask(input.StarTask)
	}
	if input.StarAction != "" {
		builder = builder.SetStarAction(input.StarAction)
	}
	if input.StarResult != "" {
		builder = builder.SetStarResult(input.StarResult)
	}
	if len(input.Keywords) > 0 {
		builder = builder.SetKeywords(input.Keywords)
	}
	if input.Source != "" {
		builder = builder.SetSource(input.Source)
	}

	return builder.Save(ctx)
}

func (s *ExperienceService) GetExperiences(ctx context.Context, userID uuid.UUID, params ExperienceListParams) ([]*ent.Experience, error) {
	query := s.db.Experience.Query().
		Where(experience.UserIDEQ(userID)).
		WithWeapons()

	if params.Category != "" {
		query = query.Where(experience.CategoryEQ(params.Category))
	}

	// Filter by weapon code if specified
	if params.WeaponCode != "" {
		query = query.Where(experience.HasWeaponsWith(experienceweapon.WeaponCodeEQ(params.WeaponCode)))
	}

	switch params.Sort {
	case "oldest":
		query = query.Order(ent.Asc(experience.FieldCreatedAt))
	case "title":
		query = query.Order(ent.Asc(experience.FieldTitle))
	default: // "latest" or empty
		query = query.Order(ent.Desc(experience.FieldCreatedAt))
	}

	return query.All(ctx)
}

func (s *ExperienceService) GetExperience(ctx context.Context, userID uuid.UUID, experienceID uuid.UUID) (*ent.Experience, error) {
	exp, err := s.db.Experience.Query().
		Where(experience.IDEQ(experienceID)).
		WithWeapons().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrExperienceNotFound
		}
		return nil, err
	}

	if exp.UserID != userID {
		return nil, ErrExperienceForbidden
	}

	return exp, nil
}

func (s *ExperienceService) UpdateExperience(ctx context.Context, userID uuid.UUID, experienceID uuid.UUID, input UpdateExperienceInput) (*ent.Experience, error) {
	exp, err := s.GetExperience(ctx, userID, experienceID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		if err := validateTitle(*input.Title); err != nil {
			return nil, err
		}
	}
	if err := validatePeriod(input.PeriodStart, input.PeriodEnd); err != nil {
		return nil, err
	}

	builder := s.db.Experience.UpdateOneID(exp.ID)

	if input.Title != nil {
		builder = builder.SetTitle(*input.Title)
	}
	if input.Category != nil {
		builder = builder.SetCategory(*input.Category)
	}
	if input.PeriodStart != nil {
		builder = builder.SetPeriodStart(*input.PeriodStart)
	}
	if input.PeriodEnd != nil {
		builder = builder.SetPeriodEnd(*input.PeriodEnd)
	}
	if input.Role != nil {
		builder = builder.SetRole(*input.Role)
	}
	if input.Content != nil {
		builder = builder.SetContent(*input.Content)
	}
	if input.Result != nil {
		builder = builder.SetResult(*input.Result)
	}
	if input.StarSituation != nil {
		builder = builder.SetStarSituation(*input.StarSituation)
	}
	if input.StarTask != nil {
		builder = builder.SetStarTask(*input.StarTask)
	}
	if input.StarAction != nil {
		builder = builder.SetStarAction(*input.StarAction)
	}
	if input.StarResult != nil {
		builder = builder.SetStarResult(*input.StarResult)
	}
	if input.Keywords != nil {
		builder = builder.SetKeywords(input.Keywords)
	}

	return builder.Save(ctx)
}

func (s *ExperienceService) DeleteExperience(ctx context.Context, userID uuid.UUID, experienceID uuid.UUID) error {
	exp, err := s.GetExperience(ctx, userID, experienceID)
	if err != nil {
		return err
	}

	// Delete related records first (cascade manually for test DB compatibility)
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return err
	}
	// Delete weapons
	_, _ = tx.ExperienceWeapon.Delete().Where(
		experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID)),
	).Exec(ctx)
	// Delete tags
	_, _ = tx.ExperienceTag.Delete().Where(
		experiencetag.HasExperienceWith(experience.IDEQ(exp.ID)),
	).Exec(ctx)
	// Delete usages
	_, _ = tx.ExperienceUsage.Delete().Where(
		experienceusage.HasExperienceWith(experience.IDEQ(exp.ID)),
	).Exec(ctx)
	// Delete the experience
	if err := tx.Experience.DeleteOneID(exp.ID).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func validateTitle(title string) error {
	if title == "" || len(title) > 200 {
		return ErrInvalidTitle
	}
	return nil
}

func validatePeriod(start, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return ErrInvalidPeriod
	}
	return nil
}
