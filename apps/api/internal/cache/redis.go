package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const linkCacheTTL = 12 * time.Hour

var ErrCacheMiss = errors.New("cache: miss")

type LinkCacheEntry struct {
	LinkID      string     `json:"link_id"`
	Destination string     `json:"destination"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	MaxClicks   *int       `json:"max_clicks,omitempty"`
	IsActive    bool       `json:"is_active"`
}

type Cache struct {
	rdb *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &Cache{rdb: redis.NewClient(opts)}, nil
}

func (c *Cache) Client() *redis.Client {
	return c.rdb
}

func linkKey(code string) string {
	return fmt.Sprintf("link:%s", code)
}

func clicksKey(code string) string {
	return fmt.Sprintf("clicks:%s", code)
}

func (c *Cache) GetLink(ctx context.Context, code string) (*LinkCacheEntry, error) {
	raw, err := c.rdb.Get(ctx, linkKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var entry LinkCacheEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (c *Cache) SetLink(ctx context.Context, code string, entry LinkCacheEntry) error {
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, linkKey(code), raw, linkCacheTTL).Err()
}

func (c *Cache) InvalidateLink(ctx context.Context, code string) error {
	return c.rdb.Del(ctx, linkKey(code)).Err()
}

// SeedClicks initializes the live click counter for a short code from its
// current Postgres-backed count. Called once when the link cache entry is
// (re)populated; a no-op if the counter already exists, so it never clobbers
// clicks recorded since the last population.
func (c *Cache) SeedClicks(ctx context.Context, code string, currentCount int) error {
	return c.rdb.SetNX(ctx, clicksKey(code), currentCount, linkCacheTTL).Err()
}

// IncrClicks atomically increments the live click counter for a short code.
// The counter must already be seeded (see SeedClicks) so this stays a single
// Redis round trip on the redirect hot path, with no Postgres read involved.
func (c *Cache) IncrClicks(ctx context.Context, code string) (int64, error) {
	return c.rdb.Incr(ctx, clicksKey(code)).Result()
}

// GetClicks reads the live click counter for a short code, without touching
// Postgres. Used to surface up-to-the-second counts to readers (e.g. the
// dashboard) even though the durable click_count column only catches up a
// couple seconds later via the batched click writer.
func (c *Cache) GetClicks(ctx context.Context, code string) (int64, error) {
	count, err := c.rdb.Get(ctx, clicksKey(code)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, ErrCacheMiss
	}
	return count, err
}

func refreshKey(userID, jti string) string {
	return fmt.Sprintf("refresh:%s:%s", userID, jti)
}

// AllowRefreshToken registers a freshly-issued refresh token's jti in an
// allowlist, TTL-matched to the token's own lifetime. /refresh and /logout
// check/delete this key, giving rotation + reuse detection "for free": a
// refresh token is only honored while its key still exists.
func (c *Cache) AllowRefreshToken(ctx context.Context, userID, jti string, ttl time.Duration) error {
	return c.rdb.Set(ctx, refreshKey(userID, jti), 1, ttl).Err()
}

func (c *Cache) IsRefreshTokenAllowed(ctx context.Context, userID, jti string) (bool, error) {
	err := c.rdb.Get(ctx, refreshKey(userID, jti)).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) RevokeRefreshToken(ctx context.Context, userID, jti string) error {
	return c.rdb.Del(ctx, refreshKey(userID, jti)).Err()
}

// One-time, TTL-limited tokens keyed by their (hashed) value within a
// namespace, mapping to whatever value they were issued for (typically a
// user id). Shared plumbing for password reset and email verification —
// both are "mail someone a link with a token, consume it once" flows.
func opaqueTokenKey(namespace, tokenHash string) string {
	return fmt.Sprintf("%s:%s", namespace, tokenHash)
}

func (c *Cache) setOpaqueToken(ctx context.Context, namespace, tokenHash, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, opaqueTokenKey(namespace, tokenHash), value, ttl).Err()
}

func (c *Cache) getOpaqueToken(ctx context.Context, namespace, tokenHash string) (string, error) {
	value, err := c.rdb.Get(ctx, opaqueTokenKey(namespace, tokenHash)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	return value, err
}

func (c *Cache) deleteOpaqueToken(ctx context.Context, namespace, tokenHash string) error {
	return c.rdb.Del(ctx, opaqueTokenKey(namespace, tokenHash)).Err()
}

const (
	passwordResetNamespace     = "pwreset"
	emailVerificationNamespace = "verify-email"
)

// SetPasswordResetToken maps a (hashed) reset token to the user it was
// issued for, TTL-limited so an unused token stops working on its own.
func (c *Cache) SetPasswordResetToken(ctx context.Context, tokenHash, userID string, ttl time.Duration) error {
	return c.setOpaqueToken(ctx, passwordResetNamespace, tokenHash, userID, ttl)
}

// GetPasswordResetUserID resolves a (hashed) reset token to the user id it
// was issued for, or ErrCacheMiss if the token is unknown/expired/used.
func (c *Cache) GetPasswordResetUserID(ctx context.Context, tokenHash string) (string, error) {
	return c.getOpaqueToken(ctx, passwordResetNamespace, tokenHash)
}

// DeletePasswordResetToken burns a reset token so it can't be replayed.
func (c *Cache) DeletePasswordResetToken(ctx context.Context, tokenHash string) error {
	return c.deleteOpaqueToken(ctx, passwordResetNamespace, tokenHash)
}

// SetEmailVerificationToken maps a (hashed) verification token to the user
// it was issued for, TTL-limited so an unused token stops working on its
// own.
func (c *Cache) SetEmailVerificationToken(ctx context.Context, tokenHash, userID string, ttl time.Duration) error {
	return c.setOpaqueToken(ctx, emailVerificationNamespace, tokenHash, userID, ttl)
}

// GetEmailVerificationUserID resolves a (hashed) verification token to the
// user id it was issued for, or ErrCacheMiss if unknown/expired/used.
func (c *Cache) GetEmailVerificationUserID(ctx context.Context, tokenHash string) (string, error) {
	return c.getOpaqueToken(ctx, emailVerificationNamespace, tokenHash)
}

// DeleteEmailVerificationToken burns a verification token so it can't be
// replayed.
func (c *Cache) DeleteEmailVerificationToken(ctx context.Context, tokenHash string) error {
	return c.deleteOpaqueToken(ctx, emailVerificationNamespace, tokenHash)
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}
