package service

import "errors"

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
)
