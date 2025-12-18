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

type fakeMsgRepo struct {
	mu           sync.Mutex
	closed       map[string]bool
	inserted     []service.Message
	nextID       int
	forceErr     error
	forceClosedE error
}

func newFakeMsgRepo() *fakeMsgRepo {
	return &fakeMsgRepo{closed: map[string]bool{}}
}

func (r *fakeMsgRepo) IsSessionClosed(ctx context.Context, sessionID string) (bool, error) {
	if r.forceClosedE != nil {
		return false, r.forceClosedE
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed[sessionID], nil
}

func (r *fakeMsgRepo) InsertMessage(ctx context.Context, msg service.Message) (service.Message, error) {
	if r.forceErr != nil {
		return service.Message{}, r.forceErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	msg.ID = "m" + itoa(r.nextID)
	r.inserted = append(r.inserted, msg)
	return msg, nil
}

func itoa(n int) string {
	// tiny helper to avoid fmt in tests
	if n < 10 {
		return string(rune('0' + n))
	}
	return "X"
}

func TestAppend_UserMessage_StripsQuotesAndTrims(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMsgRepo()
	t0 := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)
	svc := service.NewMessageService(repo, func() time.Time { return t0 })

	got, err := svc.Append(ctx, "s1", service.RoleUser, `  "hello"  `, "", nil)
	require.NoError(t, err)
	require.Equal(t, service.RoleUser, got.Role)
	require.Equal(t, "hello", got.Content)
	require.Equal(t, t0, got.CreatedAt)
	require.Equal(t, "s1", got.SessionID)
}

func TestAppend_RejectsInvalidRole(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMsgRepo()
	svc := service.NewMessageService(repo, time.Now)

	_, err := svc.Append(ctx, "s1", service.Role("wizard"), "hi", "", nil)
	require.ErrorIs(t, err, domain.ErrInvalidRole)
}

func TestAppend_RejectsEmptyUserMessage(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMsgRepo()
	svc := service.NewMessageService(repo, time.Now)

	_, err := svc.Append(ctx, "s1", service.RoleUser, "   ", "", nil)
	require.ErrorIs(t, err, domain.ErrEmptyMessage)
}

func TestAppend_ToolMessage_AllowsEmptyContentButRequiresToolName(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMsgRepo()
	svc := service.NewMessageService(repo, time.Now)

	// ok
	_, err := svc.Append(ctx, "s1", service.RoleTool, "", "calendly", map[string]any{"x": 1})
	require.NoError(t, err)

	// not ok
	_, err = svc.Append(ctx, "s1", service.RoleTool, "", "   ", map[string]any{})
	require.ErrorIs(t, err, domain.ErrEmptyMessage)
}

func TestAppend_FailsIfSessionClosed(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMsgRepo()
	repo.closed["s1"] = true
	svc := service.NewMessageService(repo, time.Now)

	_, err := svc.Append(ctx, "s1", service.RoleUser, "hi", "", nil)
	require.ErrorIs(t, err, domain.ErrSessionClosed)
}
