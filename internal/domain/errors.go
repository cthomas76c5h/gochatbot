package domain

import "errors"

var (
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPhone            = errors.New("invalid phone")
	ErrInvalidColor            = errors.New("invalid color")
	ErrInvalidSlug             = errors.New("invalid slug")
	ErrUnknownEmailProfile     = errors.New("unknown email profile")
	ErrTemplateVersionNotFound = errors.New("template version not found")
	ErrSessionNotFound         = errors.New("session not found")
	ErrInvalidRole             = errors.New("invalid role")
	ErrEmptyMessage            = errors.New("empty message")
	ErrSessionClosed           = errors.New("session closed")
	ErrInvalidCursor           = errors.New("invalid cursor")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrTenantSlugTaken         = errors.New("tenant slug taken")
)
