package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shortking-api/internal/service"
)

const ContextUserIDKey = "userID"

// AuthRequired verifies the Authorization: Bearer <access token> header and
// sets the authenticated user's id in the Gin context for downstream
// handlers (and for RateLimitPerUser).
func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := auth.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID.String())
		c.Next()
	}
}

// EmailVerified requires the authenticated user (set by AuthRequired) to
// have a verified email. Gates actions where a disposable/fake email would
// let someone create links they can immediately disown, e.g. shortening
// for abuse — must run after AuthRequired.
func EmailVerified(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.GetString(ContextUserIDKey))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		user, err := auth.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		if user.EmailVerifiedAt == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "verify your email before creating links"})
			return
		}

		c.Next()
	}
}
