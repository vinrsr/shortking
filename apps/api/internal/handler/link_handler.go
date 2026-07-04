package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"shortking-api/internal/models"
	"shortking-api/internal/repository"
	"shortking-api/internal/service"
	"shortking-api/pkg/qrcode"
)

type LinkHandler struct {
	links  *service.LinkService
	clicks repository.ClickRepository
	stats  *service.StatsService
}

func NewLinkHandler(links *service.LinkService, clicks repository.ClickRepository, stats *service.StatsService) *LinkHandler {
	return &LinkHandler{links: links, clicks: clicks, stats: stats}
}

type createLinkRequest struct {
	Destination string     `json:"destination" binding:"required,url"`
	CustomAlias string     `json:"customAlias"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	MaxClicks   *int       `json:"maxClicks"`
}

type linkResponse struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"shortCode"`
	ShortURL    string     `json:"shortUrl"`
	Destination string     `json:"destination"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	MaxClicks   *int       `json:"maxClicks,omitempty"`
	ClickCount  int        `json:"clickCount"`
	IsActive    bool       `json:"isActive"`
	QRGenerated bool       `json:"qrGenerated"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func (h *LinkHandler) toResponse(l *models.Link) linkResponse {
	return linkResponse{
		ID:          l.ID.String(),
		ShortCode:   l.ShortCode,
		ShortURL:    h.links.ShortURL(l.ShortCode),
		Destination: l.Destination,
		ExpiresAt:   l.ExpiresAt,
		MaxClicks:   l.MaxClicks,
		ClickCount:  l.ClickCount,
		IsActive:    l.IsActive,
		QRGenerated: l.QRGeneratedAt != nil,
		CreatedAt:   l.CreatedAt,
	}
}

func (h *LinkHandler) Create(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.links.Create(c.Request.Context(), service.CreateLinkInput{
		UserID:      &userID,
		Destination: req.Destination,
		CustomAlias: req.CustomAlias,
		ExpiresAt:   req.ExpiresAt,
		MaxClicks:   req.MaxClicks,
	})
	if err != nil {
		writeLinkError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(link))
}

type updateLinkRequest struct {
	Destination string     `json:"destination" binding:"required,url"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	MaxClicks   *int       `json:"maxClicks"`
	IsActive    bool       `json:"isActive"`
}

func (h *LinkHandler) Update(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.links.Update(c.Request.Context(), userID, linkID, service.UpdateLinkInput{
		Destination: req.Destination,
		ExpiresAt:   req.ExpiresAt,
		MaxClicks:   req.MaxClicks,
		IsActive:    req.IsActive,
	})
	if err != nil {
		writeLinkError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(link))
}

type createAnonymousLinkRequest struct {
	Destination string `json:"destination" binding:"required,url"`
}

// CreateAnonymous is the public, no-login shorten flow used by the landing
// page: destination URL only, no custom alias, no custom expiry, no
// max-clicks. Those require an account (see Create above).
func (h *LinkHandler) CreateAnonymous(c *gin.Context) {
	var req createAnonymousLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.links.CreateAnonymous(c.Request.Context(), req.Destination)
	if err != nil {
		writeLinkError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(link))
}

func (h *LinkHandler) List(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	links, err := h.links.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}

	resp := make([]linkResponse, 0, len(links))
	for i := range links {
		resp = append(resp, h.toResponse(&links[i]))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *LinkHandler) Get(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	link, err := h.links.GetOwned(c.Request.Context(), userID, linkID)
	if err != nil {
		writeLinkError(c, err)
		return
	}

	clicks, err := h.clicks.ListByLink(c.Request.Context(), link.ID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load click history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link":   h.toResponse(link),
		"clicks": clicks,
	})
}

func (h *LinkHandler) Delete(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	if err := h.links.Delete(c.Request.Context(), userID, linkID); err != nil {
		writeLinkError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) QRCode(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	link, err := h.links.GetOwned(c.Request.Context(), userID, linkID)
	if err != nil {
		writeLinkError(c, err)
		return
	}

	size := qrcode.DefaultSize
	if raw := c.Query("size"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			size = parsed
		}
	}

	png, err := qrcode.PNG(h.links.ShortURL(link.ShortCode), size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

// RecordQRCodeGeneration records that the user generated a QR code for this
// link, both persisted on the link (so the dashboard keeps showing it after
// a refresh) and counted toward the landing-page stat. It's a separate call
// from QRCode (GET) because that endpoint is hit once to preview the code
// and again to download it — counting there double-counts every generation.
func (h *LinkHandler) RecordQRCodeGeneration(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	firstTime, err := h.links.MarkQRGenerated(c.Request.Context(), userID, linkID)
	if err != nil {
		writeLinkError(c, err)
		return
	}

	if firstTime {
		if err := h.stats.RecordQRGeneration(c.Request.Context()); err != nil {
			log.Printf("stats: failed to record QR generation: %v", err)
		}
	}

	c.Status(http.StatusNoContent)
}

func writeLinkError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrLinkNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrAliasTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
