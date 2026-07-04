package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shortking-api/internal/middleware"
)

func userIDFromContext(c *gin.Context) (uuid.UUID, error) {
	raw, _ := c.Get(middleware.ContextUserIDKey)
	return uuid.Parse(raw.(string))
}
