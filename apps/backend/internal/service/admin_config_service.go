package service

import (
	"context"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/systemconfig"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// AdminConfigService handles system configuration and prompt template management.
type AdminConfigService struct {
	db *ent.Client
}

// NewAdminConfigService creates a new admin config service.
func NewAdminConfigService(db *ent.Client) *AdminConfigService {
	return &AdminConfigService{db: db}
}

// UpdateConfigInput contains the fields for updating a system config.
type UpdateConfigInput struct {
	Key      string
	Value    string
	CallerID uuid.UUID
}

// UpdatePromptInput contains the optional fields for updating a prompt template.
type UpdatePromptInput struct {
	ID                 uuid.UUID
	SystemPrompt       *string
	UserPromptTemplate *string
	ModelName          *string
	Temperature        *float64
	MaxTokens          *int
	IsActive           *bool
}

// ListConfigs returns system configuration entries with optional category filter.
func (s *AdminConfigService) ListConfigs(ctx context.Context, category string) ([]*ent.SystemConfig, error) {
	query := s.db.SystemConfig.Query()
	if category != "" {
		query = query.Where(systemconfig.CategoryEQ(category))
	}
	return query.
		Order(ent.Asc(systemconfig.FieldCategory, systemconfig.FieldConfigKey)).
		All(ctx)
}

// UpdateConfig updates a system configuration value by key.
// Returns the updated config and the old value (for audit logging).
func (s *AdminConfigService) UpdateConfig(ctx context.Context, input UpdateConfigInput) (*ent.SystemConfig, string, error) {
	cfg, err := s.db.SystemConfig.Query().
		Where(systemconfig.ConfigKeyEQ(input.Key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", ErrAdminConfigNotFound
		}
		return nil, "", err
	}

	oldValue := cfg.ConfigValue
	updated, err := s.db.SystemConfig.UpdateOne(cfg).
		SetConfigValue(input.Value).
		SetUpdatedBy(input.CallerID).
		Save(ctx)
	if err != nil {
		return nil, "", err
	}

	return updated, oldValue, nil
}

// ListPrompts returns prompt templates with optional category filter.
func (s *AdminConfigService) ListPrompts(ctx context.Context, category string) ([]*ent.PromptTemplate, error) {
	query := s.db.PromptTemplate.Query()
	if category != "" {
		query = query.Where(prompttemplate.CategoryEQ(category))
	}
	return query.
		Order(ent.Asc(prompttemplate.FieldCategory, prompttemplate.FieldSubCategory)).
		All(ctx)
}

// UpdatePrompt updates a prompt template's fields.
func (s *AdminConfigService) UpdatePrompt(ctx context.Context, input UpdatePromptInput) (*ent.PromptTemplate, error) {
	update := s.db.PromptTemplate.UpdateOneID(input.ID)

	if input.SystemPrompt != nil {
		update = update.SetSystemPrompt(*input.SystemPrompt)
	}
	if input.UserPromptTemplate != nil {
		update = update.SetUserPromptTemplate(*input.UserPromptTemplate)
	}
	if input.ModelName != nil {
		if !ai.IsSelectableModel(*input.ModelName) {
			return nil, ErrAdminUnsupportedModel
		}
		update = update.SetModel(*input.ModelName)
	}
	if input.Temperature != nil {
		update = update.SetTemperature(*input.Temperature)
	}
	if input.MaxTokens != nil {
		update = update.SetMaxTokens(*input.MaxTokens)
	}
	if input.IsActive != nil {
		update = update.SetIsActive(*input.IsActive)
	}

	prompt, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminPromptNotFound
		}
		return nil, err
	}

	return prompt, nil
}
