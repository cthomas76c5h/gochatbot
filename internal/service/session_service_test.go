package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/service"
)

type fakeRepo struct {
	mu sync.Mutex

	sessions map[string]service.Session
	leads    map[string]service.Lead // key: sessionID

	createLeadCount int
	closeCount      int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		sessions: make(map[string]service.Session),
		leads:    make(map[string]service.Lead),
	}
}

func (r *fakeRepo) GetSession(ctx context.Context, sessionID string) (service.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[sessionID]
	if !ok {
		return service.Session{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *fakeRepo) MarkSessionClosed(ctx context.Context, sessionID string, closedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[sessionID]
	if !ok {
		return domain.ErrSessionNotFound
	}
	s.ClosedAt = &closedAt
	r.sessions[sessionID] = s
	r.closeCount++
	return nil
}

func (r *fakeRepo) GetLeadBySession(ctx context.Context, sessionID string) (service.Lead, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lead, ok := r.leads[sessionID]
	return lead, ok, nil
}

func (r *fakeRepo) CreateLeadForSession(ctx context.Context, sessionID string) (service.Lead, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lead := service.Lead{ID: "lead-" + sessionID, SessionID: sessionID}
	r.leads[sessionID] = lead
	r.createLeadCount++
	return lead, nil
}

type fakeQueue struct {
	mu   sync.Mutex
	jobs []job
}

type job struct {
	kind    string
	payload map[string]any
}

func (q *fakeQueue) Enqueue(ctx context.Context, kind string, payload map[string]any) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job{kind: kind, payload: payload})
	return nil
}

func TestCloseSession_CreatesLeadAndEnqueuesOnce(t *testing.T) {
	ctx := context.Background()

	repo := newFakeRepo()
	q := &fakeQueue{}
	t0 := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)

	repo.sessions["s1"] = service.Session{ID: "s1", ClosedAt: nil}

	svc := service.NewSessionService(repo, q, func() time.Time { return t0 })

	err := svc.CloseSession(ctx, "s1")
	require.NoError(t, err)

	require.Equal(t, 1, repo.closeCount)
	require.Equal(t, 1, repo.createLeadCount)

	require.Len(t, q.jobs, 1)
	require.Equal(t, "export_lead", q.jobs[0].kind)
	require.Equal(t, "s1", q.jobs[0].payload["session_id"])
	require.Equal(t, "lead-s1", q.jobs[0].payload["lead_id"])
}

func TestCloseSession_IdempotentSecondCallDoesNothing(t *testing.T) {
	ctx := context.Background()

	repo := newFakeRepo()
	q := &fakeQueue{}
	t0 := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)

	repo.sessions["s1"] = service.Session{ID: "s1", ClosedAt: nil}

	svc := service.NewSessionService(repo, q, func() time.Time { return t0 })

	require.NoError(t, svc.CloseSession(ctx, "s1"))
	require.NoError(t, svc.CloseSession(ctx, "s1"))

	// Still only closed once, lead created once, job enqueued once
	require.Equal(t, 1, repo.closeCount)
	require.Equal(t, 1, repo.createLeadCount)
	require.Len(t, q.jobs, 1)
}

func TestCloseSession_NotFound(t *testing.T) {
	ctx := context.Background()

	repo := newFakeRepo()
	q := &fakeQueue{}

	svc := service.NewSessionService(repo, q, func() time.Time { return time.Now() })

	err := svc.CloseSession(ctx, "missing")
	require.ErrorIs(t, err, domain.ErrSessionNotFound)
}
