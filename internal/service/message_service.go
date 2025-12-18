package service

import (
	"context"
	"strings"
	"time"

	"gochatbot/internal/domain"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Message struct {
	ID        string
	SessionID string
	Role      Role
	Content   string
	ToolName  string
	ToolData  map[string]any
	CreatedAt time.Time
}

type MessageRepo interface {
	IsSessionClosed(ctx context.Context, sessionID string) (bool, error)
	InsertMessage(ctx context.Context, msg Message) (Message, error)
}

type MessageService struct {
	repo MessageRepo
	now  func() time.Time
}

func NewMessageService(repo MessageRepo, now func() time.Time) *MessageService {
	if now == nil {
		now = time.Now
	}
	return &MessageService{repo: repo, now: now}
}

func (s *MessageService) Append(ctx context.Context, sessionID string, role Role, content string, toolName string, toolData map[string]any) (Message, error) {
	if !isValidRole(role) {
		return Message{}, domain.ErrInvalidRole
	}

	closed, err := s.repo.IsSessionClosed(ctx, sessionID)
	if err != nil {
		return Message{}, err
	}
	if closed {
		return Message{}, domain.ErrSessionClosed
	}

	content = strings.TrimSpace(content)

	// match your legacy vibe: strip outer quotes for non-tool messages (optional, but useful)
	if role != RoleTool {
		content = stripOuterQuotes(content)
		content = strings.TrimSpace(content)
	}

	// tool messages may have empty "content" but must have a tool name
	if role == RoleTool {
		if strings.TrimSpace(toolName) == "" {
			return Message{}, domain.ErrEmptyMessage
		}
	} else {
		if content == "" {
			return Message{}, domain.ErrEmptyMessage
		}
	}

	msg := Message{
		ID:        "", // repo assigns, or leave blank for now
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		ToolName:  strings.TrimSpace(toolName),
		ToolData:  toolData,
		CreatedAt: s.now(),
	}

	return s.repo.InsertMessage(ctx, msg)
}

func isValidRole(r Role) bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem, RoleTool:
		return true
	default:
		return false
	}
}

func stripOuterQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
