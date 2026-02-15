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

	// Cover letter errors
	ErrCoverLetterNotFound  = errors.New("cover letter not found")
	ErrCoverLetterForbidden = errors.New("forbidden: not the owner of this cover letter")

	// Usage errors
	ErrUsageLimitExceeded = errors.New("usage limit exceeded")
	ErrInvalidFeature     = errors.New("invalid feature name")

	// Interview errors
	ErrInvalidInterviewStage = errors.New("invalid interview stage")
	ErrEmptySTARField        = errors.New("STAR extraction returned empty field")
)
