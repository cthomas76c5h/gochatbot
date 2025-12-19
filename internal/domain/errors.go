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

	// Templates
	ErrTemplateNotFound  = errors.New("template not found")
	ErrTemplateSlugTaken = errors.New("template slug taken")
	ErrTemplateImmutable = errors.New("template immutable")

	// Template Versions
	ErrVersionNotFound           = errors.New("version not found")
	ErrVersionAlreadyPublished   = errors.New("version already published")
	ErrPublishedVersionImmutable = errors.New("published version immutable")
)
