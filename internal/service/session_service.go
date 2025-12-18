package service

import (
	"context"
	"time"

	"gochatbot/internal/domain"
)

type Session struct {
	ID       string
	ClosedAt *time.Time
}

type Lead struct {
	ID        string
	SessionID string
}

type Repo interface {
	GetSession(ctx context.Context, sessionID string) (Session, error)
	MarkSessionClosed(ctx context.Context, sessionID string, closedAt time.Time) error

	GetLeadBySession(ctx context.Context, sessionID string) (Lead, bool, error)
	CreateLeadForSession(ctx context.Context, sessionID string) (Lead, error)
}

type Queue interface {
	Enqueue(ctx context.Context, kind string, payload map[string]any) error
}

type SessionService struct {
	repo  Repo
	queue Queue
	now   func() time.Time
}

func NewSessionService(repo Repo, queue Queue, now func() time.Time) *SessionService {
	if now == nil {
		now = time.Now
	}
	return &SessionService{repo: repo, queue: queue, now: now}
}

// CloseSession is idempotent:
// - if already closed: OK
// - else: close session, create lead if missing, enqueue export job once
func (s *SessionService) CloseSession(ctx context.Context, sessionID string) error {
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if sess.ClosedAt != nil {
		return nil
	}

	t := s.now()

	// Mark closed first (so repeated calls stop here)
	if err := s.repo.MarkSessionClosed(ctx, sessionID, t); err != nil {
		return err
	}

	lead, exists, err := s.repo.GetLeadBySession(ctx, sessionID)
	if err != nil {
		return err
	}
	if !exists {
		lead, err = s.repo.CreateLeadForSession(ctx, sessionID)
		if err != nil {
			return err
		}
	}

	// enqueue export job (idempotency enforced by lead existence + close state)
	return s.queue.Enqueue(ctx, "export_lead", map[string]any{
		"session_id": sessionID,
		"lead_id":    lead.ID,
	})
}

var _ = domain.ErrSessionNotFound
