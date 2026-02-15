package service

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/coverletter"
	"github.com/coby/colight/apps/backend/ent/coverletterversion"
	"github.com/google/uuid"
)

// EditorService handles cover letter editing
type EditorService struct {
	entClient *ent.Client
}

// NewEditorService creates a new editor service
func NewEditorService(entClient *ent.Client) *EditorService {
	return &EditorService{
		entClient: entClient,
	}
}

// GetCoverLetter retrieves a cover letter with latest version
func (s *EditorService) GetCoverLetter(ctx context.Context, userID uuid.UUID, coverLetterID uuid.UUID) (*ent.CoverLetter, error) {
	cl, err := s.entClient.CoverLetter.Query().
		Where(coverletter.IDEQ(coverLetterID)).
		WithApplication().
		WithVersions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("cover letter not found")
		}
		return nil, err
	}

	if cl.UserID != userID {
		return nil, fmt.Errorf("forbidden: not the owner")
	}

	return cl, nil
}

// UpdateContent updates cover letter content (auto-save)
func (s *EditorService) UpdateContent(ctx context.Context, userID uuid.UUID, coverLetterID uuid.UUID, content string) error {
	cl, err := s.GetCoverLetter(ctx, userID, coverLetterID)
	if err != nil {
		return err
	}

	_, err = cl.Update().
		SetCurrentContent(content).
		Save(ctx)

	return err
}

// CreateVersion creates a new version with optional change summary
func (s *EditorService) CreateVersion(ctx context.Context, userID uuid.UUID, coverLetterID uuid.UUID, content string, changeSummary string) (*ent.CoverLetterVersion, error) {
	cl, err := s.GetCoverLetter(ctx, userID, coverLetterID)
	if err != nil {
		return nil, err
	}

	// Get next version number
	versions, _ := s.entClient.CoverLetterVersion.Query().
		Where(coverletterversion.HasCoverLetterWith(coverletter.IDEQ(coverLetterID))).
		All(ctx)

	nextVersion := len(versions) + 1
	charCount := len([]rune(content))

	builder := s.entClient.CoverLetterVersion.Create().
		SetCoverLetter(cl).
		SetVersionNumber(nextVersion).
		SetContent(content).
		SetCharCount(charCount)

	if changeSummary != "" {
		builder.SetChangeSummary(changeSummary)
	}

	version, err := builder.Save(ctx)
	return version, err
}

// GetVersions retrieves all versions for a cover letter
func (s *EditorService) GetVersions(ctx context.Context, userID uuid.UUID, coverLetterID uuid.UUID) ([]*ent.CoverLetterVersion, error) {
	// Verify ownership
	_, err := s.GetCoverLetter(ctx, userID, coverLetterID)
	if err != nil {
		return nil, err
	}

	versions, err := s.entClient.CoverLetterVersion.Query().
		Where(coverletterversion.HasCoverLetterWith(coverletter.IDEQ(coverLetterID))).
		Order(ent.Desc("version_number")).
		All(ctx)

	return versions, err
}
