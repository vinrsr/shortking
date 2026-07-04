// Package testsupport holds in-memory fakes for the repository interfaces
// and a miniredis-backed cache constructor, shared by service- and
// handler-level tests so neither needs a live Postgres or Redis.
package testsupport

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"shortking-api/internal/cache"
	"shortking-api/internal/models"
	"shortking-api/internal/repository"
)

// NewTestCache spins up an in-memory Redis (miniredis) and wraps it in a
// real *cache.Cache, so tests exercise the actual cache-aside logic instead
// of a hand-rolled fake.
func NewTestCache(t testing.TB) *cache.Cache {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.New("redis://" + mr.Addr())
	require.NoError(t, err)
	return c
}

// FakeLinkRepository is an in-memory stand-in for repository.LinkRepository.
type FakeLinkRepository struct {
	mu   sync.Mutex
	byID map[uuid.UUID]*models.Link
}

func NewFakeLinkRepository() *FakeLinkRepository {
	return &FakeLinkRepository{byID: map[uuid.UUID]*models.Link{}}
}

func (f *FakeLinkRepository) Create(ctx context.Context, link *models.Link) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, l := range f.byID {
		if l.ShortCode == link.ShortCode {
			return repository.ErrConflict
		}
	}
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	link.CreatedAt = time.Now()
	cp := *link
	f.byID[link.ID] = &cp
	return nil
}

func (f *FakeLinkRepository) FindByShortCode(ctx context.Context, code string) (*models.Link, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, l := range f.byID {
		if l.ShortCode == code {
			cp := *l
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *FakeLinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *l
	return &cp, nil
}

func (f *FakeLinkRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Link, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Link
	for _, l := range f.byID {
		if l.UserID != nil && *l.UserID == userID {
			out = append(out, *l)
		}
	}
	return out, nil
}

func (f *FakeLinkRepository) Update(ctx context.Context, link *models.Link) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[link.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *link
	f.byID[link.ID] = &cp
	return nil
}

func (f *FakeLinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, id)
	return nil
}

func (f *FakeLinkRepository) IncrementClickCount(ctx context.Context, id uuid.UUID, by int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.byID[id]
	if !ok {
		return repository.ErrNotFound
	}
	l.ClickCount += by
	return nil
}

func (f *FakeLinkRepository) CountAll(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.byID)), nil
}

func (f *FakeLinkRepository) CountActive(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	var n int64
	for _, l := range f.byID {
		if !l.IsActive {
			continue
		}
		if l.ExpiresAt != nil && l.ExpiresAt.Before(now) {
			continue
		}
		if l.MaxClicks != nil && l.ClickCount >= *l.MaxClicks {
			continue
		}
		n++
	}
	return n, nil
}

// FakeUserRepository is an in-memory stand-in for repository.UserRepository.
type FakeUserRepository struct {
	mu   sync.Mutex
	byID map[uuid.UUID]*models.User
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{byID: map[uuid.UUID]*models.User{}}
}

func (f *FakeUserRepository) Create(ctx context.Context, user *models.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Email == user.Email {
			return repository.ErrConflict
		}
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	cp := *user
	f.byID[user.ID] = &cp
	return nil
}

func (f *FakeUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.byID {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *FakeUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *FakeUserRepository) Update(ctx context.Context, user *models.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[user.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *user
	f.byID[user.ID] = &cp
	return nil
}

func (f *FakeUserRepository) CountAll(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.byID)), nil
}

// FakeClickRepository is an in-memory stand-in for repository.ClickRepository.
type FakeClickRepository struct {
	mu     sync.Mutex
	events []models.ClickEvent
	batchN int
}

func NewFakeClickRepository() *FakeClickRepository {
	return &FakeClickRepository{}
}

func (f *FakeClickRepository) CreateBatch(ctx context.Context, events []models.ClickEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, events...)
	f.batchN++
	return nil
}

func (f *FakeClickRepository) ListByLink(ctx context.Context, linkID uuid.UUID, limit int) ([]models.ClickEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.ClickEvent
	for _, e := range f.events {
		if e.LinkID == linkID {
			out = append(out, e)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (f *FakeClickRepository) CountAll(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.events)), nil
}

// Count returns the number of click events recorded so far. Safe for
// concurrent use while a ClickRecorder is flushing in the background.
func (f *FakeClickRepository) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

// Events returns a snapshot of the recorded click events.
func (f *FakeClickRepository) Events() []models.ClickEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.ClickEvent, len(f.events))
	copy(out, f.events)
	return out
}

// BatchCount returns how many CreateBatch calls have landed, i.e. how many
// separate flushes have happened.
func (f *FakeClickRepository) BatchCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.batchN
}

// FakeStatsRepository is an in-memory stand-in for repository.StatsRepository.
type FakeStatsRepository struct {
	mu    sync.Mutex
	count int64
}

func NewFakeStatsRepository() *FakeStatsRepository {
	return &FakeStatsRepository{}
}

func (f *FakeStatsRepository) IncrementQRGenerations(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.count++
	return nil
}

func (f *FakeStatsRepository) TotalQRGenerations(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.count, nil
}
