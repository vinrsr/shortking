package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"shortking-api/internal/handler"
	"shortking-api/internal/middleware"
	"shortking-api/internal/service"
)

type Deps struct {
	Redis          *redis.Client
	AllowedOrigins []string
	Auth           *handler.AuthHandler
	Link           *handler.LinkHandler
	Redirect       *handler.RedirectHandler
	Stats          *handler.StatsHandler
	AuthSvc        *service.AuthService
}

func New(deps Deps) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(middleware.CORS(deps.AllowedOrigins))

	authRateLimit, err := middleware.RateLimit(deps.Redis, "auth", "10-M")
	if err != nil {
		return nil, err
	}
	createLinkRateLimit, err := middleware.RateLimitPerUser(deps.Redis, "create-link", "60-M")
	if err != nil {
		return nil, err
	}
	// Two limits on anonymous shortening: a burst guard against abuse, and a
	// much tighter daily cap that's a product decision, not an anti-abuse
	// one — it's what actually pushes a habitual anonymous user to sign up.
	anonymousShortenBurstLimit, err := middleware.RateLimit(deps.Redis, "anon-shorten-burst", "10-M")
	if err != nil {
		return nil, err
	}
	anonymousShortenDailyLimit, err := middleware.RateLimitWithMessage(
		deps.Redis, "anon-shorten-daily", "5-D",
		"you've reached today's free shorten limit — sign up for unlimited links",
	)
	if err != nil {
		return nil, err
	}
	redirectRateLimit, err := middleware.RateLimit(deps.Redis, "redirect", "300-M")
	if err != nil {
		return nil, err
	}

	authRequired := middleware.AuthRequired(deps.AuthSvc)
	emailVerifiedRequired := middleware.EmailVerified(deps.AuthSvc)

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		authGroup.Use(authRateLimit)
		{
			authGroup.POST("/signup", deps.Auth.Signup)
			authGroup.POST("/login", deps.Auth.Login)
			authGroup.POST("/refresh", deps.Auth.Refresh)
			authGroup.POST("/logout", deps.Auth.Logout)
			authGroup.POST("/forgot-password", deps.Auth.ForgotPassword)
			authGroup.POST("/reset-password", deps.Auth.ResetPassword)
			authGroup.POST("/verify-email", deps.Auth.VerifyEmail)
			authGroup.POST("/resend-verification", deps.Auth.ResendVerification)
		}

		api.POST("/shorten", anonymousShortenBurstLimit, anonymousShortenDailyLimit, deps.Link.CreateAnonymous)
		api.GET("/stats", deps.Stats.Get)
		api.GET("/me", authRequired, deps.Auth.Me)

		links := api.Group("/links")
		links.Use(authRequired)
		{
			links.POST("", emailVerifiedRequired, createLinkRateLimit, deps.Link.Create)
			links.GET("", deps.Link.List)
			links.GET("/:id", deps.Link.Get)
			links.PATCH("/:id", deps.Link.Update)
			links.DELETE("/:id", deps.Link.Delete)
			links.GET("/:id/qrcode", deps.Link.QRCode)
			links.POST("/:id/qrcode/generations", deps.Link.RecordQRCodeGeneration)
		}
	}

	r.GET("/:code", redirectRateLimit, deps.Redirect.Redirect)

	return r, nil
}
