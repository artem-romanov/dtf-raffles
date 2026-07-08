package domain

import "errors"

// Authentication Errors
var (
	ErrInvalidCredentials = errors.New("credentials invalid")
)

// Session Errors
var (
	ErrUserSessionNotFound = errors.New("usersession not found")
)

// Telegram Errors
var (
	ErrTelegramUserNotFound = errors.New("telegram user not found")
	ErrTelegramUserExists   = errors.New("telegram user already exists")
)
