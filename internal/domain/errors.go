package domain

import "errors"

// Authentication Errors
var (
	ErrInvalidCredentials = errors.New("Credentials invalid")
)

// Session Errors
var (
	ErrUserSessionNotFound = errors.New("UserSession not found")
)
