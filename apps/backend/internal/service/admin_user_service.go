package service

import (
	"context"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/deletionrequest"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// roleLevel returns the numeric level for a role string.
// Duplicated from middleware to avoid import cycle (service -> middleware -> service).
func roleLevel(role string) int {
	switch role {
	case "user":
		return 1
	case "manager":
		return 2
	case "admin":
		return 3
	case "super_admin":
		return 4
	default:
		return 0
	}
}

// AdminUserService handles user management operations.
type AdminUserService struct {
	db *ent.Client
}

// NewAdminUserService creates a new admin user service.
func NewAdminUserService(db *ent.Client) *AdminUserService {
	return &AdminUserService{db: db}
}

// AdminListUsersParams contains parameters for listing users.
type AdminListUsersParams struct {
	Limit      int
	Offset     int
	RoleFilter string
	Search     string
}

// AdminListUsersResult contains the paginated user list and total count.
type AdminListUsersResult struct {
	Users []*ent.UserProfile
	Total int
}

// UpdateRoleInput contains the fields for updating a user's role.
type UpdateRoleInput struct {
	TargetID        uuid.UUID
	CallerID        uuid.UUID
	CallerRoleLevel int
	CallerRoleStr   string
	NewRole         string
}

// CreateUserInput contains the fields for creating a new user.
type CreateUserInput struct {
	Email    string
	Password string
	Nickname string
	Role     string
	Plan     string
}

// AdminUserDetail contains user info with related entity counts.
type AdminUserDetail struct {
	User            *ent.UserProfile
	ExperienceCount int
	CoachingCount   int
	UsageCount      int
}

// AdminUserExport contains user profile and related data for export.
type AdminUserExport struct {
	User        *ent.UserProfile
	Experiences []*ent.Experience
}

// ListUsers returns a paginated user list with optional role/search filter.
func (s *AdminUserService) ListUsers(ctx context.Context, params AdminListUsersParams) (*AdminListUsersResult, error) {
	query := s.db.UserProfile.Query()

	if params.RoleFilter != "" {
		query = query.Where(userprofile.RoleEQ(userprofile.Role(params.RoleFilter)))
	}
	if params.Search != "" {
		query = query.Where(
			userprofile.Or(
				userprofile.NicknameContains(params.Search),
				userprofile.EmailContains(params.Search),
			),
		)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	users, err := query.
		Limit(params.Limit).
		Offset(params.Offset).
		Order(ent.Desc(userprofile.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return &AdminListUsersResult{Users: users, Total: total}, nil
}

// GetUser returns a single user by ID.
func (s *AdminUserService) GetUser(ctx context.Context, id uuid.UUID) (*ent.UserProfile, error) {
	user, err := s.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// CreateUser creates a new user with the given input.
func (s *AdminUserService) CreateUser(ctx context.Context, input CreateUserInput) (*ent.UserProfile, error) {
	exists, _ := s.db.UserProfile.Query().
		Where(userprofile.EmailEQ(input.Email)).
		Exist(ctx)
	if exists {
		return nil, ErrAdminEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, err
	}

	builder := s.db.UserProfile.Create().
		SetEmail(input.Email).
		SetPasswordHash(string(hash)).
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.Role(input.Role))
	if input.Nickname != "" {
		builder = builder.SetNickname(input.Nickname)
	}
	if input.Plan != "" {
		builder = builder.SetPlan(userprofile.Plan(input.Plan))
	}

	return builder.Save(ctx)
}

// UpdateUserRole changes a user's role with hierarchy validation.
// Returns the updated user and the previous role string.
func (s *AdminUserService) UpdateUserRole(ctx context.Context, input UpdateRoleInput) (*ent.UserProfile, string, error) {
	// Self-change protection
	if input.CallerID == input.TargetID {
		return nil, "", ErrAdminSelfRoleChange
	}

	// Caller can only assign roles strictly below their own level
	targetLevel := roleLevel(input.NewRole)
	if targetLevel >= input.CallerRoleLevel {
		return nil, "", ErrAdminInsufficientLevel
	}

	// Check the current role of the target user
	target, err := s.db.UserProfile.Get(ctx, input.TargetID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", ErrAdminUserNotFound
		}
		return nil, "", err
	}

	// Cannot modify users at or above caller's level (except super_admin can modify other super_admins)
	targetCurrentLevel := roleLevel(string(target.Role))
	if targetCurrentLevel >= input.CallerRoleLevel && !(input.CallerRoleStr == "super_admin" && string(target.Role) == "super_admin") {
		return nil, "", ErrAdminTargetTooHigh
	}

	// Last super_admin protection
	if target.Role == userprofile.RoleSuperAdmin {
		count, err := s.db.UserProfile.Query().
			Where(userprofile.RoleEQ(userprofile.RoleSuperAdmin)).
			Count(ctx)
		if err == nil && count <= 1 {
			return nil, "", ErrAdminLastSuperAdmin
		}
	}

	oldRole := string(target.Role)
	user, err := s.db.UserProfile.UpdateOneID(input.TargetID).
		SetRole(userprofile.Role(input.NewRole)).
		Save(ctx)
	if err != nil {
		return nil, "", err
	}

	return user, oldRole, nil
}

// SuspendUser suspends a user account.
func (s *AdminUserService) SuspendUser(ctx context.Context, id uuid.UUID, reason string) (*ent.UserProfile, error) {
	user, err := s.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	if user.Suspended {
		return nil, ErrAdminUserAlreadySuspended
	}

	now := time.Now()
	update := s.db.UserProfile.UpdateOne(user).
		SetSuspended(true).
		SetSuspendedAt(now)
	if reason != "" {
		update = update.SetSuspendedReason(reason)
	}

	return update.Save(ctx)
}

// UnsuspendUser lifts suspension on a user account.
func (s *AdminUserService) UnsuspendUser(ctx context.Context, id uuid.UUID) (*ent.UserProfile, error) {
	updated, err := s.db.UserProfile.UpdateOneID(id).
		SetSuspended(false).
		ClearSuspendedAt().
		ClearSuspendedReason().
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return updated, nil
}

// GetUserDetail returns detailed information about a user including related entity counts.
func (s *AdminUserService) GetUserDetail(ctx context.Context, id uuid.UUID) (*AdminUserDetail, error) {
	user, err := s.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	expCount, _ := user.QueryExperiences().Count(ctx)
	coachCount, _ := user.QueryCoachingSessions().Count(ctx)
	usageCount, _ := user.QueryUsageLogs().Count(ctx)

	return &AdminUserDetail{
		User:            user,
		ExperienceCount: expCount,
		CoachingCount:   coachCount,
		UsageCount:      usageCount,
	}, nil
}

// ForceLogout invalidates all tokens for a user by setting force_logout_at.
func (s *AdminUserService) ForceLogout(ctx context.Context, id uuid.UUID) (*ent.UserProfile, error) {
	now := time.Now()
	updated, err := s.db.UserProfile.UpdateOneID(id).
		SetForceLogoutAt(now).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return updated, nil
}

// UpdateUserPlan changes a user's subscription plan.
func (s *AdminUserService) UpdateUserPlan(ctx context.Context, id uuid.UUID, plan string) (*ent.UserProfile, error) {
	updated, err := s.db.UserProfile.UpdateOneID(id).
		SetPlan(userprofile.Plan(plan)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}
	return updated, nil
}

// ExportUserData exports all user data for PIPA compliance.
func (s *AdminUserService) ExportUserData(ctx context.Context, id uuid.UUID) (*AdminUserExport, error) {
	user, err := s.db.UserProfile.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	experiences, _ := user.QueryExperiences().All(ctx)

	return &AdminUserExport{
		User:        user,
		Experiences: experiences,
	}, nil
}

// CreateDeletionRequest creates a data deletion request for a user.
func (s *AdminUserService) CreateDeletionRequest(ctx context.Context, userID, adminID uuid.UUID, reason string) (*ent.DeletionRequest, error) {
	if _, err := s.db.UserProfile.Get(ctx, userID); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	scheduledAt := time.Now().AddDate(0, 0, 30)
	create := s.db.DeletionRequest.Create().
		SetUserID(userID).
		SetRequestedBy(adminID).
		SetScheduledAt(scheduledAt)
	if reason != "" {
		create = create.SetReason(reason)
	}

	return create.Save(ctx)
}

// CancelDeletionRequest cancels a pending deletion request for a user.
func (s *AdminUserService) CancelDeletionRequest(ctx context.Context, userID uuid.UUID) (*ent.DeletionRequest, error) {
	dr, err := s.db.DeletionRequest.Query().
		Where(
			deletionrequest.UserIDEQ(userID),
			deletionrequest.StatusEQ(deletionrequest.StatusPending),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrAdminDeletionNotFound
		}
		return nil, err
	}

	now := time.Now()
	return s.db.DeletionRequest.UpdateOne(dr).
		SetStatus(deletionrequest.StatusCancelled).
		SetCancelledAt(now).
		Save(ctx)
}

// ListDeletionQueue returns pending deletion requests.
func (s *AdminUserService) ListDeletionQueue(ctx context.Context) ([]*ent.DeletionRequest, error) {
	return s.db.DeletionRequest.Query().
		Where(deletionrequest.StatusEQ(deletionrequest.StatusPending)).
		Order(ent.Asc(deletionrequest.FieldScheduledAt)).
		All(ctx)
}
