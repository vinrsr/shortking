package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"shortking-api/internal/cache"
	"shortking-api/internal/models"
	"shortking-api/internal/repository"
	"shortking-api/internal/shortcode"
)

var (
	ErrLinkNotFound = errors.New("service: link not found")
	ErrLinkExpired  = errors.New("service: link expired")
	ErrAliasTaken   = errors.New("service: alias already in use")
	ErrForbidden    = errors.New("service: not the owner of this link")
)

// AnonymousLinkTTL is the fixed, non-configurable expiry for links created
// without an account, short on purpose, to make the value of signing up
// (custom aliases, custom/no expiry, QR codes, analytics) obvious.
const AnonymousLinkTTL = 48 * time.Hour

type CreateLinkInput struct {
	UserID      *uuid.UUID
	Destination string
	CustomAlias string
	ExpiresAt   *time.Time
	MaxClicks   *int
}

type LinkService struct {
	links        repository.LinkRepository
	cache        *cache.Cache
	baseShortURL string
}

func NewLinkService(links repository.LinkRepository, c *cache.Cache, baseShortURL string) *LinkService {
	return &LinkService{links: links, cache: c, baseShortURL: baseShortURL}
}

func (s *LinkService) Create(ctx context.Context, in CreateLinkInput) (*models.Link, error) {
	code := in.CustomAlias
	if code != "" {
		if err := shortcode.ValidateAlias(code); err != nil {
			return nil, err
		}
		return s.createWithCode(ctx, in, code)
	}

	var lastErr error
	for attempt := 0; attempt < shortcode.MaxGenAttempts; attempt++ {
		generated, err := shortcode.Generate()
		if err != nil {
			return nil, err
		}
		link, err := s.createWithCode(ctx, in, generated)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, ErrAliasTaken) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (s *LinkService) createWithCode(ctx context.Context, in CreateLinkInput, code string) (*models.Link, error) {
	link := &models.Link{
		UserID:      in.UserID,
		ShortCode:   code,
		Destination: in.Destination,
		ExpiresAt:   in.ExpiresAt,
		MaxClicks:   in.MaxClicks,
		IsActive:    true,
	}

	if err := s.links.Create(ctx, link); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrAliasTaken
		}
		return nil, err
	}

	return link, nil
}

// CreateAnonymous creates a link with no owner: auto-generated code only (no
// custom alias), a fixed short expiry, and no max-clicks. Used by the public
// landing-page shorten flow, which requires no login.
func (s *LinkService) CreateAnonymous(ctx context.Context, destination string) (*models.Link, error) {
	expiresAt := time.Now().Add(AnonymousLinkTTL)
	return s.Create(ctx, CreateLinkInput{
		Destination: destination,
		ExpiresAt:   &expiresAt,
	})
}

func (s *LinkService) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Link, error) {
	links, err := s.links.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range links {
		s.applyLiveClickCount(ctx, &links[i])
	}
	return links, nil
}

// applyLiveClickCount overrides a link's DB-persisted click_count with the
// live Redis counter, when one exists. The counter is incremented
// synchronously on every redirect (see ResolveRedirect), while click_count
// only catches up a couple seconds later via the batched click writer, so
// the live counter is always the fresher of the two.
func (s *LinkService) applyLiveClickCount(ctx context.Context, link *models.Link) {
	live, err := s.cache.GetClicks(ctx, link.ShortCode)
	if err != nil {
		return
	}
	if int(live) > link.ClickCount {
		link.ClickCount = int(live)
	}
}

func (s *LinkService) CountAll(ctx context.Context) (int64, error) {
	return s.links.CountAll(ctx)
}

func (s *LinkService) CountActive(ctx context.Context) (int64, error) {
	return s.links.CountActive(ctx)
}

func (s *LinkService) GetOwned(ctx context.Context, userID, linkID uuid.UUID) (*models.Link, error) {
	link, err := s.findOwned(ctx, userID, linkID)
	if err != nil {
		return nil, err
	}
	s.applyLiveClickCount(ctx, link)
	return link, nil
}

// findOwned fetches a link by id and checks ownership, without merging in
// the live click count. Callers that are about to persist the link (Update)
// must use this instead of GetOwned, saving a live-count-inflated struct
// back to Postgres would stomp on the batched click writer's own increments.
func (s *LinkService) findOwned(ctx context.Context, userID, linkID uuid.UUID) (*models.Link, error) {
	link, err := s.links.FindByID(ctx, linkID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	if link.UserID == nil || *link.UserID != userID {
		return nil, ErrForbidden
	}
	return link, nil
}

type UpdateLinkInput struct {
	Destination string
	ExpiresAt   *time.Time
	MaxClicks   *int
	IsActive    bool
}

// Update changes a link's destination, expiry, max-clicks limit, and active
// flag. The short code itself is not editable, it's the link's identity,
// and changing it out from under anyone who already has the URL would break
// it, so this only ever needs to invalidate (not repopulate) the cache entry.
func (s *LinkService) Update(ctx context.Context, userID, linkID uuid.UUID, in UpdateLinkInput) (*models.Link, error) {
	link, err := s.findOwned(ctx, userID, linkID)
	if err != nil {
		return nil, err
	}

	link.Destination = in.Destination
	link.ExpiresAt = in.ExpiresAt
	link.MaxClicks = in.MaxClicks
	link.IsActive = in.IsActive

	if err := s.links.Update(ctx, link); err != nil {
		return nil, err
	}
	if err := s.cache.InvalidateLink(ctx, link.ShortCode); err != nil {
		return nil, err
	}

	s.applyLiveClickCount(ctx, link)
	return link, nil
}

func (s *LinkService) Delete(ctx context.Context, userID, linkID uuid.UUID) error {
	link, err := s.findOwned(ctx, userID, linkID)
	if err != nil {
		return err
	}
	if err := s.links.Delete(ctx, link.ID); err != nil {
		return err
	}
	return s.cache.InvalidateLink(ctx, link.ShortCode)
}

func (s *LinkService) ShortURL(code string) string {
	return s.baseShortURL + "/" + code
}

// MarkQRGenerated records that a QR code has been generated for this link,
// so the dashboard keeps showing it after a refresh instead of re-prompting
// "Generate QR code". Idempotent: firstTime is false on repeat calls, so the
// caller can count the landing-page stat only once per link.
func (s *LinkService) MarkQRGenerated(ctx context.Context, userID, linkID uuid.UUID) (firstTime bool, err error) {
	link, err := s.findOwned(ctx, userID, linkID)
	if err != nil {
		return false, err
	}
	if link.QRGeneratedAt != nil {
		return false, nil
	}

	now := time.Now()
	link.QRGeneratedAt = &now
	if err := s.links.Update(ctx, link); err != nil {
		return false, err
	}
	return true, nil
}

type ResolvedLink struct {
	LinkID      uuid.UUID
	Destination string
}

// ResolveRedirect resolves a short code to its destination (and link id, for
// click recording) for the public redirect endpoint. It checks Redis first
// (cache-aside); on a miss it falls back to Postgres and repopulates the
// cache. Expiration (by date or click count, checked against the live Redis
// counter) evicts the cache entry and returns ErrLinkExpired so the redirect
// handler can respond 410.
func (s *LinkService) ResolveRedirect(ctx context.Context, code string) (ResolvedLink, error) {
	entry, err := s.cache.GetLink(ctx, code)
	if errors.Is(err, cache.ErrCacheMiss) {
		entry, err = s.populateCache(ctx, code)
	}
	if err != nil {
		return ResolvedLink{}, err
	}

	if !entry.IsActive {
		return ResolvedLink{}, ErrLinkNotFound
	}

	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		_ = s.cache.InvalidateLink(ctx, code)
		return ResolvedLink{}, ErrLinkExpired
	}

	// Always incremented (not just for max-click links) so the live counter
	// can back an up-to-date clickCount for readers like the dashboard,
	// ahead of the batched Postgres write.
	count, err := s.cache.IncrClicks(ctx, code)
	if err != nil {
		return ResolvedLink{}, err
	}
	if entry.MaxClicks != nil && int(count) > *entry.MaxClicks {
		_ = s.cache.InvalidateLink(ctx, code)
		return ResolvedLink{}, ErrLinkExpired
	}

	linkID, err := uuid.Parse(entry.LinkID)
	if err != nil {
		return ResolvedLink{}, err
	}

	return ResolvedLink{LinkID: linkID, Destination: entry.Destination}, nil
}

func (s *LinkService) populateCache(ctx context.Context, code string) (*cache.LinkCacheEntry, error) {
	link, err := s.links.FindByShortCode(ctx, code)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}

	entry := cache.LinkCacheEntry{
		LinkID:      link.ID.String(),
		Destination: link.Destination,
		ExpiresAt:   link.ExpiresAt,
		MaxClicks:   link.MaxClicks,
		IsActive:    link.IsActive,
	}
	if err := s.cache.SetLink(ctx, code, entry); err != nil {
		return nil, err
	}
	if err := s.cache.SeedClicks(ctx, code, link.ClickCount); err != nil {
		return nil, err
	}
	return &entry, nil
}
