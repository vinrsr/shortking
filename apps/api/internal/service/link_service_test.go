package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortking-api/internal/cache"
	"shortking-api/internal/models"
	"shortking-api/internal/repository"
	"shortking-api/internal/testsupport"
)

func newTestLinkService(t *testing.T) (*LinkService, *testsupport.FakeLinkRepository, *cache.Cache) {
	t.Helper()
	repo := testsupport.NewFakeLinkRepository()
	c := testsupport.NewTestCache(t)
	return NewLinkService(repo, c, "http://sk.io"), repo, c
}

func TestCreate_GeneratesRandomCodeWhenNoAliasGiven(t *testing.T) {
	svc, _, _ := newTestLinkService(t)

	link, err := svc.Create(context.Background(), CreateLinkInput{Destination: "https://example.com"})

	require.NoError(t, err)
	assert.Len(t, link.ShortCode, 7)
}

func TestCreate_CustomAliasIsUsedVerbatim(t *testing.T) {
	svc, _, _ := newTestLinkService(t)

	link, err := svc.Create(context.Background(), CreateLinkInput{
		Destination: "https://example.com",
		CustomAlias: "my-link",
	})

	require.NoError(t, err)
	assert.Equal(t, "my-link", link.ShortCode)
}

func TestCreate_InvalidCustomAliasIsRejected(t *testing.T) {
	svc, _, _ := newTestLinkService(t)

	_, err := svc.Create(context.Background(), CreateLinkInput{
		Destination: "https://example.com",
		CustomAlias: "a", // shorter than shortcode.MinAliasLen
	})

	assert.ErrorIs(t, err, ErrInvalidAlias)
	assert.NotErrorIs(t, err, ErrAliasTaken) // format error, not a conflict
}

func TestCreate_TakenCustomAliasReturnsErrAliasTaken(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	require.NoError(t, repo.Create(ctx, &models.Link{
		ShortCode: "taken", Destination: "https://other.com", IsActive: true,
	}))

	_, err := svc.Create(ctx, CreateLinkInput{Destination: "https://example.com", CustomAlias: "taken"})

	assert.ErrorIs(t, err, ErrAliasTaken)
}

func TestResolveRedirect_CacheMissPopulatesFromRepository(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	link := &models.Link{ShortCode: "abc1234", Destination: "https://example.com", IsActive: true}
	require.NoError(t, repo.Create(ctx, link))

	resolved, err := svc.ResolveRedirect(ctx, "abc1234")

	require.NoError(t, err)
	assert.Equal(t, link.ID, resolved.LinkID)
	assert.Equal(t, "https://example.com", resolved.Destination)
}

func TestResolveRedirect_UnknownCodeReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestLinkService(t)

	_, err := svc.ResolveRedirect(context.Background(), "nope")

	assert.ErrorIs(t, err, ErrLinkNotFound)
}

func TestResolveRedirect_InactiveLinkReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	require.NoError(t, repo.Create(ctx, &models.Link{
		ShortCode: "abc1234", Destination: "https://example.com", IsActive: false,
	}))

	_, err := svc.ResolveRedirect(ctx, "abc1234")

	assert.ErrorIs(t, err, ErrLinkNotFound)
}

func TestResolveRedirect_PastExpiryReturnsExpired(t *testing.T) {
	ctx := context.Background()
	svc, repo, c := newTestLinkService(t)
	past := time.Now().Add(-time.Hour)
	require.NoError(t, repo.Create(ctx, &models.Link{
		ShortCode: "abc1234", Destination: "https://example.com", IsActive: true, ExpiresAt: &past,
	}))

	_, err := svc.ResolveRedirect(ctx, "abc1234")

	assert.ErrorIs(t, err, ErrLinkExpired)
	_, err = c.GetLink(ctx, "abc1234")
	assert.ErrorIs(t, err, cache.ErrCacheMiss, "expired link should be evicted from cache")
}

func TestResolveRedirect_MaxClicksEnforced(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	max := 2
	require.NoError(t, repo.Create(ctx, &models.Link{
		ShortCode: "abc1234", Destination: "https://example.com", IsActive: true, MaxClicks: &max,
	}))

	_, err := svc.ResolveRedirect(ctx, "abc1234")
	require.NoError(t, err, "1st click should be within the limit")

	_, err = svc.ResolveRedirect(ctx, "abc1234")
	require.NoError(t, err, "2nd click should be within the limit")

	_, err = svc.ResolveRedirect(ctx, "abc1234")
	assert.ErrorIs(t, err, ErrLinkExpired, "3rd click should exceed maxClicks=2")
}

func TestResolveRedirect_ClicksAreCountedEvenWithoutMaxClicks(t *testing.T) {
	// Regression: the live click counter must be incremented on every
	// redirect, not just for links that have a maxClicks limit, so readers
	// like the dashboard can see an up-to-date count ahead of the batched
	// Postgres write.
	ctx := context.Background()
	svc, repo, c := newTestLinkService(t)
	require.NoError(t, repo.Create(ctx, &models.Link{
		ShortCode: "abc1234", Destination: "https://example.com", IsActive: true,
	}))

	_, err := svc.ResolveRedirect(ctx, "abc1234")
	require.NoError(t, err)
	_, err = svc.ResolveRedirect(ctx, "abc1234")
	require.NoError(t, err)

	live, err := c.GetClicks(ctx, "abc1234")
	require.NoError(t, err)
	assert.EqualValues(t, 2, live)
}

func TestListByUser_PrefersLiveRedisCountOverStaleDBCount(t *testing.T) {
	// Regression for the dashboard "clicks don't show up" fix: click_count in
	// Postgres only catches up a couple seconds after a redirect (see
	// ClickRecorder), so ListByUser must prefer the live Redis counter.
	ctx := context.Background()
	svc, repo, c := newTestLinkService(t)
	userID := uuid.New()
	link := &models.Link{
		UserID: &userID, ShortCode: "abc1234", Destination: "https://example.com",
		IsActive: true, ClickCount: 2, // stale value, as if the DB flush hasn't run yet
	}
	require.NoError(t, repo.Create(ctx, link))
	require.NoError(t, c.SeedClicks(ctx, link.ShortCode, 2))
	live, err := c.IncrClicks(ctx, link.ShortCode)
	require.NoError(t, err)
	require.EqualValues(t, 3, live)

	links, err := svc.ListByUser(ctx, userID)

	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, 3, links[0].ClickCount)
}

func TestListByUser_FallsBackToDBCountWhenNeverClicked(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	userID := uuid.New()
	require.NoError(t, repo.Create(ctx, &models.Link{
		UserID: &userID, ShortCode: "abc1234", Destination: "https://example.com",
		IsActive: true, ClickCount: 7,
	}))

	links, err := svc.ListByUser(ctx, userID)

	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, 7, links[0].ClickCount)
}

func TestGetOwned_ForbidsAccessByAnotherUser(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	owner := uuid.New()
	other := uuid.New()
	link := &models.Link{UserID: &owner, ShortCode: "abc1234", Destination: "https://example.com", IsActive: true}
	require.NoError(t, repo.Create(ctx, link))

	_, err := svc.GetOwned(ctx, other, link.ID)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpdate_ChangesEditableFields(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	owner := uuid.New()
	link := &models.Link{
		UserID: &owner, ShortCode: "abc1234", Destination: "https://old.example.com", IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, link))
	newExpiry := time.Now().Add(24 * time.Hour)
	maxClicks := 10

	updated, err := svc.Update(ctx, owner, link.ID, UpdateLinkInput{
		Destination: "https://new.example.com",
		ExpiresAt:   &newExpiry,
		MaxClicks:   &maxClicks,
		IsActive:    false,
	})

	require.NoError(t, err)
	assert.Equal(t, "https://new.example.com", updated.Destination)
	assert.Equal(t, "abc1234", updated.ShortCode, "short code must not change on update")
	assert.False(t, updated.IsActive)
	require.NotNil(t, updated.MaxClicks)
	assert.Equal(t, 10, *updated.MaxClicks)

	stored, err := repo.FindByID(ctx, link.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://new.example.com", stored.Destination)
	assert.False(t, stored.IsActive)
}

func TestUpdate_DoesNotClobberInFlightClickCount(t *testing.T) {
	// Regression: Update must not save a live-Redis-inflated click count
	// back over the DB value, or it would race with the batched click
	// writer's own atomic increment and could lose counted clicks.
	ctx := context.Background()
	svc, repo, c := newTestLinkService(t)
	owner := uuid.New()
	link := &models.Link{
		UserID: &owner, ShortCode: "abc1234", Destination: "https://example.com",
		IsActive: true, ClickCount: 2,
	}
	require.NoError(t, repo.Create(ctx, link))
	require.NoError(t, c.SeedClicks(ctx, link.ShortCode, 2))
	_, err := c.IncrClicks(ctx, link.ShortCode) // live counter now ahead of the DB value
	require.NoError(t, err)

	_, err = svc.Update(ctx, owner, link.ID, UpdateLinkInput{Destination: "https://example.com", IsActive: true})
	require.NoError(t, err)

	stored, err := repo.FindByID(ctx, link.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, stored.ClickCount, "the DB click_count must be untouched by Update")
}

func TestUpdate_ForbidsAnotherUsersLink(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	owner := uuid.New()
	other := uuid.New()
	link := &models.Link{UserID: &owner, ShortCode: "abc1234", Destination: "https://example.com", IsActive: true}
	require.NoError(t, repo.Create(ctx, link))

	_, err := svc.Update(ctx, other, link.ID, UpdateLinkInput{Destination: "https://evil.example.com", IsActive: true})

	assert.ErrorIs(t, err, ErrForbidden)
	stored, ferr := repo.FindByID(ctx, link.ID)
	require.NoError(t, ferr)
	assert.Equal(t, "https://example.com", stored.Destination, "the link must be unchanged")
}

func TestUpdate_InvalidatesCacheSoStaleEntriesArentServed(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newTestLinkService(t)
	owner := uuid.New()
	link := &models.Link{UserID: &owner, ShortCode: "abc1234", Destination: "https://old.example.com", IsActive: true}
	require.NoError(t, repo.Create(ctx, link))
	_, err := svc.ResolveRedirect(ctx, link.ShortCode) // populate the cache with the old destination
	require.NoError(t, err)

	_, err = svc.Update(ctx, owner, link.ID, UpdateLinkInput{Destination: "https://new.example.com", IsActive: true})
	require.NoError(t, err)

	resolved, err := svc.ResolveRedirect(ctx, link.ShortCode)
	require.NoError(t, err)
	assert.Equal(t, "https://new.example.com", resolved.Destination)
}

func TestDelete_RemovesLinkAndInvalidatesCache(t *testing.T) {
	ctx := context.Background()
	svc, repo, c := newTestLinkService(t)
	owner := uuid.New()
	link := &models.Link{UserID: &owner, ShortCode: "abc1234", Destination: "https://example.com", IsActive: true}
	require.NoError(t, repo.Create(ctx, link))
	_, err := svc.ResolveRedirect(ctx, link.ShortCode) // populate the cache
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, owner, link.ID))

	_, err = repo.FindByID(ctx, link.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	_, err = c.GetLink(ctx, link.ShortCode)
	assert.ErrorIs(t, err, cache.ErrCacheMiss)
}
