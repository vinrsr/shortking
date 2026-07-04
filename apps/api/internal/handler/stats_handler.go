package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shortking-api/internal/repository"
	"shortking-api/internal/service"
)

type StatsHandler struct {
	links  *service.LinkService
	auth   *service.AuthService
	clicks repository.ClickRepository
	stats  *service.StatsService
}

func NewStatsHandler(
	links *service.LinkService,
	auth *service.AuthService,
	clicks repository.ClickRepository,
	stats *service.StatsService,
) *StatsHandler {
	return &StatsHandler{links: links, auth: auth, clicks: clicks, stats: stats}
}

// Get returns public, aggregate stats for landing-page social proof. No auth
// required, nothing user-specific.
func (h *StatsHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	totalLinks, err := h.links.CountAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	activeLinks, err := h.links.CountActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	totalUsers, err := h.auth.CountUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	totalClicks, err := h.clicks.CountAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	totalQRCodes, err := h.stats.TotalQRGenerations(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalLinks":   totalLinks,
		"activeLinks":  activeLinks,
		"totalUsers":   totalUsers,
		"totalClicks":  totalClicks,
		"totalQrCodes": totalQRCodes,
	})
}
