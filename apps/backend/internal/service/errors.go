package service

import "errors"

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")

	// Experience errors
	ErrExperienceNotFound  = errors.New("experience not found")
	ErrExperienceForbidden = errors.New("forbidden: not the owner of this experience")
	ErrInvalidTitle        = errors.New("title is required and must be 1-200 characters")
	ErrInvalidPeriod       = errors.New("period_end must be after period_start")

	// Matching errors
	ErrNoExperiences = errors.New("no experiences found for this user")

	// Application errors
	ErrApplicationNotFound  = errors.New("application not found")
	ErrApplicationForbidden = errors.New("forbidden: not the owner of this application")

	// Analysis errors
	ErrAnalysisNotFound  = errors.New("company analysis not found")
	ErrAnalysisForbidden = errors.New("forbidden: not the owner of this analysis")

	// Cover letter errors
	ErrCoverLetterNotFound  = errors.New("cover letter not found")
	ErrCoverLetterForbidden = errors.New("forbidden: not the owner of this cover letter")

	// Usage errors
	ErrUsageLimitExceeded = errors.New("usage limit exceeded")
	ErrInvalidFeature     = errors.New("invalid feature name")

	// Interview errors
	ErrInvalidInterviewStage = errors.New("invalid interview stage")
	ErrEmptySTARField        = errors.New("STAR extraction returned empty field")

	// Admin errors
	ErrAdminUserNotFound         = errors.New("admin: user not found")
	ErrAdminSelfRoleChange       = errors.New("admin: cannot change own role")
	ErrAdminInsufficientLevel    = errors.New("admin: insufficient role level for target role")
	ErrAdminTargetTooHigh        = errors.New("admin: target user role is at or above caller level")
	ErrAdminLastSuperAdmin       = errors.New("admin: cannot demote the last super_admin")
	ErrAdminUserAlreadySuspended = errors.New("admin: user is already suspended")
	ErrAdminEmailExists          = errors.New("admin: email already exists")
	ErrAdminDeletionNotFound     = errors.New("admin: no pending deletion request found")
	ErrAdminConfigNotFound       = errors.New("admin: config not found")
	ErrAdminPromptNotFound       = errors.New("admin: prompt template not found")
	ErrAdminUnsupportedModel     = errors.New("admin: unsupported model name")
	ErrAdminFeedbackNotFound     = errors.New("admin: feedback not found")
)
