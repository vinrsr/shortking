package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"shortking-api/internal/cache"
	"shortking-api/internal/config"
	"shortking-api/internal/db"
	"shortking-api/internal/handler"
	"shortking-api/internal/mailer"
	"shortking-api/internal/repository"
	"shortking-api/internal/router"
	"shortking-api/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gormDB, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	redisCache, err := cache.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func() { _ = redisCache.Close() }()

	userRepo := repository.NewUserRepository(gormDB)
	linkRepo := repository.NewLinkRepository(gormDB)
	clickRepo := repository.NewClickRepository(gormDB)
	statsRepo := repository.NewStatsRepository(gormDB)

	authService := service.NewAuthService(
		userRepo, redisCache,
		cfg.JWTAccessSecret, cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL, cfg.JWTRefreshTTL,
	)
	linkService := service.NewLinkService(linkRepo, redisCache, cfg.BaseShortURL)
	clickRecorder := service.NewClickRecorder(clickRepo, linkRepo, cfg.IPHashPepper)
	statsService := service.NewStatsService(statsRepo)
	mailerSvc := mailer.New(mailer.Config{
		APIKey: cfg.ResendAPIKey,
		From:   cfg.EmailFrom,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clickRecorder.Start(ctx, 0)

	engine, err := router.New(router.Deps{
		Redis:          redisCache.Client(),
		AllowedOrigins: cfg.CORSAllowedOrigins,
		Auth:           handler.NewAuthHandler(authService, mailerSvc, cfg.WebBaseURL),
		Link:           handler.NewLinkHandler(linkService, clickRepo, statsService),
		Redirect:       handler.NewRedirectHandler(linkService, clickRecorder, cfg.WebBaseURL),
		Stats:          handler.NewStatsHandler(linkService, authService, clickRepo, statsService),
		AuthSvc:        authService,
	})
	if err != nil {
		log.Fatalf("router: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	go func() {
		log.Printf("shortking-api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	clickRecorder.Shutdown(shutdownTimeout)
}
