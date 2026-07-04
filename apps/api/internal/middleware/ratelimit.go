package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

const defaultRateLimitMessage = "rate limit exceeded"

// RateLimit builds Gin middleware that rate limits by client IP, using a
// Redis-backed store so limits are shared across API instances.
// formatted is a limiter.Rate spec, e.g. "10-M" for 10 requests per minute.
// name must be unique across all rate limiters in the app — it namespaces
// the Redis keys so two limiters checking the same key (e.g. two IP-keyed
// limits on different routes) don't silently share one counter.
func RateLimit(rdb *redis.Client, name, formatted string) (gin.HandlerFunc, error) {
	return rateLimit(rdb, name, formatted, defaultRateLimitMessage, ginlimiter.DefaultKeyGetter)
}

// RateLimitWithMessage is like RateLimit but replies with a custom message
// when the limit is hit — for limits that are a product decision (e.g. a
// daily quota meant to nudge signup) rather than generic abuse protection,
// where the default "rate limit exceeded" would be a confusing thing to see.
func RateLimitWithMessage(rdb *redis.Client, name, formatted, message string) (gin.HandlerFunc, error) {
	return rateLimit(rdb, name, formatted, message, ginlimiter.DefaultKeyGetter)
}

// RateLimitPerUser is like RateLimit but keys on the authenticated user id
// (set by AuthRequired) instead of client IP, falling back to IP if the
// request has no authenticated user.
func RateLimitPerUser(rdb *redis.Client, name, formatted string) (gin.HandlerFunc, error) {
	return rateLimit(rdb, name, formatted, defaultRateLimitMessage, func(c *gin.Context) string {
		if userID, ok := c.Get(ContextUserIDKey); ok {
			return userID.(string)
		}
		return ginlimiter.DefaultKeyGetter(c)
	})
}

func rateLimit(rdb *redis.Client, name, formatted, message string, keyGetter ginlimiter.KeyGetter) (gin.HandlerFunc, error) {
	rate, err := limiter.NewRateFromFormatted(formatted)
	if err != nil {
		return nil, err
	}

	store, err := redisstore.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:          "ratelimit:" + name,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
		MaxRetry:        limiter.DefaultMaxRetry,
	})
	if err != nil {
		return nil, err
	}

	lim := limiter.New(store, rate)

	return ginlimiter.NewMiddleware(
		lim,
		ginlimiter.WithKeyGetter(keyGetter),
		ginlimiter.WithErrorHandler(func(c *gin.Context, err error) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limiter unavailable"})
		}),
		ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": message})
		}),
	), nil
}
