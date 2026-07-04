package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortking-api/internal/service"
)

type RedirectHandler struct {
	links  *service.LinkService
	clicks *service.ClickRecorder
	// webBaseURL points at the frontend, used to bounce expired/missing
	// codes to a branded page instead of returning raw JSON.
	webBaseURL string
}

func NewRedirectHandler(links *service.LinkService, clicks *service.ClickRecorder, webBaseURL string) *RedirectHandler {
	return &RedirectHandler{links: links, clicks: clicks, webBaseURL: webBaseURL}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	resolved, err := h.links.ResolveRedirect(c.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			c.Redirect(http.StatusFound, h.webBaseURL+"/link-not-found")
		case errors.Is(err, service.ErrLinkExpired):
			c.Redirect(http.StatusFound, h.webBaseURL+"/link-expired")
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	// Fire-and-forget: never let analytics slow down the redirect response.
	h.clicks.Record(resolved.LinkID, c.Request.Referer(), c.Request.UserAgent(), c.ClientIP())

	c.Redirect(http.StatusFound, resolved.Destination)
}
