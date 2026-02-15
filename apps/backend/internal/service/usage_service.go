package service

import (
	"context"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/usagelog"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/google/uuid"
)

// UsageService handles freemium usage tracking and limit enforcement.
type UsageService struct {
	db *ent.Client
}

// NewUsageService creates a new usage service.
func NewUsageService(db *ent.Client) *UsageService {
	return &UsageService{db: db}
}

type featureLimit struct {
	LimitType string // "total" or "daily"
	Limit     int
}

// FreeLimits defines the usage limits for free-tier users.
var FreeLimits = map[string]featureLimit{
	"experience":        {LimitType: "total", Limit: 3},
	"analysis":          {LimitType: "daily", Limit: 1},
	"question_analysis": {LimitType: "daily", Limit: 1},
	"draft":             {LimitType: "daily", Limit: 1},
	"review":            {LimitType: "daily", Limit: 1},
}

// KST is the Korea Standard Time timezone.
var kst = time.FixedZone("KST", 9*60*60)

// UsageStatus represents the current usage status for a feature.
type UsageStatus struct {
	Allowed   bool `json:"allowed"`
	Used      int  `json:"used"`
	Limit     int  `json:"limit"`
	Remaining int  `json:"remaining"`
}

// CheckLimit checks whether the user can use a feature.
func (s *UsageService) CheckLimit(ctx context.Context, userID uuid.UUID, feature string) (*UsageStatus, error) {
	fl, ok := FreeLimits[feature]
	if !ok {
		return nil, ErrInvalidFeature
	}

	// Check user plan
	user, err := s.db.UserProfile.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Paid users have no limits
	if user.Plan != userprofile.PlanFree {
		return &UsageStatus{
			Allowed:   true,
			Used:      0,
			Limit:     -1, // unlimited
			Remaining: -1,
		}, nil
	}

	// Count usage
	query := s.db.UsageLog.Query().
		Where(
			usagelog.UserIDEQ(userID),
			usagelog.FeatureEQ(feature),
		)

	if fl.LimitType == "daily" {
		now := time.Now().In(kst)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, kst)
		query = query.Where(usagelog.CreatedAtGTE(startOfDay))
	}

	used, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	remaining := fl.Limit - used
	if remaining < 0 {
		remaining = 0
	}

	return &UsageStatus{
		Allowed:   used < fl.Limit,
		Used:      used,
		Limit:     fl.Limit,
		Remaining: remaining,
	}, nil
}

// TrackUsage records a feature usage event.
func (s *UsageService) TrackUsage(ctx context.Context, userID uuid.UUID, feature string, metadata map[string]interface{}) error {
	if _, ok := FreeLimits[feature]; !ok {
		return ErrInvalidFeature
	}

	builder := s.db.UsageLog.Create().
		SetUserID(userID).
		SetFeature(feature)

	if metadata != nil {
		builder = builder.SetMetadata(metadata)
	}

	_, err := builder.Save(ctx)
	return err
}

// GetAllUsage returns usage status for all features for a user.
func (s *UsageService) GetAllUsage(ctx context.Context, userID uuid.UUID) (map[string]*UsageStatus, string, error) {
	user, err := s.db.UserProfile.Get(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	plan := string(user.Plan)
	result := make(map[string]*UsageStatus)

	for feature := range FreeLimits {
		status, err := s.checkLimitForPlan(ctx, userID, feature, user.Plan)
		if err != nil {
			return nil, "", err
		}
		result[feature] = status
	}

	return result, plan, nil
}

// checkLimitForPlan is an internal helper that skips the user lookup.
func (s *UsageService) checkLimitForPlan(ctx context.Context, userID uuid.UUID, feature string, plan userprofile.Plan) (*UsageStatus, error) {
	fl := FreeLimits[feature]

	if plan != userprofile.PlanFree {
		return &UsageStatus{
			Allowed:   true,
			Used:      0,
			Limit:     -1,
			Remaining: -1,
		}, nil
	}

	query := s.db.UsageLog.Query().
		Where(
			usagelog.UserIDEQ(userID),
			usagelog.FeatureEQ(feature),
		)

	if fl.LimitType == "daily" {
		now := time.Now().In(kst)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, kst)
		query = query.Where(usagelog.CreatedAtGTE(startOfDay))
	}

	used, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	remaining := fl.Limit - used
	if remaining < 0 {
		remaining = 0
	}

	return &UsageStatus{
		Allowed:   used < fl.Limit,
		Used:      used,
		Limit:     fl.Limit,
		Remaining: remaining,
	}, nil
}
