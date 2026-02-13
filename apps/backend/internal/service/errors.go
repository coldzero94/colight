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
)
