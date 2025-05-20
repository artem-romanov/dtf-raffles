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

// Telegram Errors
var (
	ErrTelegramUserNotFound = errors.New("Telegram User not found")
	ErrTelegramUserExists   = errors.New("Telegram User already exists")
)
