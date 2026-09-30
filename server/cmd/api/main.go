package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"shijibu/config"
	"shijibu/internal/handler"
	platformauth "shijibu/internal/platform/auth"
	"shijibu/internal/platform/bailian"
	"shijibu/internal/platform/database"
	"shijibu/internal/platform/wechat"
	"shijibu/internal/router"
	"shijibu/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	tokens := platformauth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpire)
	wechatClient := wechat.NewClient(cfg.WeChatAppID, cfg.WeChatAppSecret, cfg.AllowMockLogin)
	accounts := service.NewAccountService(db, wechatClient, tokens, cfg.UploadDir)
	speech := service.NewSpeechService(bailian.NewClient(cfg.BailianAPIKey, cfg.BailianBaseURL, cfg.BailianASRModel))
	h := handler.New(accounts, service.NewDiscoveryService(db), service.NewRecordService(db, wechatClient), service.NewContributionService(db), speech,
		service.NewEngagementService(db), service.NewNutritionService(db), service.NewMediaService(db, cfg.UploadDir, cfg.PublicBaseURL),
		service.NewVerificationService(db, wechatClient, cfg.DataEncryptionKey))

	server := &http.Server{Addr: cfg.ServerAddr, Handler: router.New(h, tokens, cfg.CORSAllowOrigins, cfg.UploadDir), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("NutriLens API listening on %s", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve API: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
