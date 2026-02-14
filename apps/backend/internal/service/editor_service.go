package service

import (
	"context"
	"fmt"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/coverletter"
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
