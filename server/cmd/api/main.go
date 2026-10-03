package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	platformredis "shijibu/internal/cache/redis"
	"shijibu/internal/config"
	"shijibu/internal/handler"
	"shijibu/internal/model/pgsql"
	platformauth "shijibu/internal/platform/auth"
	"shijibu/internal/platform/bailian"
	"shijibu/internal/platform/llm"
	"shijibu/internal/platform/wechat"
	"shijibu/internal/router"
	"shijibu/internal/service"

	"gorm.io/gorm"
)

type dependencies struct {
	database *gorm.DB
	redis    *platformredis.Client
	wechat   *wechat.Client
	bailian  *bailian.Client
	llm      *llm.Client
	tokens   *platformauth.SessionManager
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, config.Current()); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("application stopped: %v", err)
	}
}

// run 明确选择并初始化当前 API 进程需要的内部组件，退出时负责释放资源，
// 随后再装配外部客户端与 HTTP 服务。
func run(ctx context.Context, cfg config.Config) (runErr error) {
	if err := pgsql.Init(); err != nil {
		return err
	}
	defer func() {
		if err := pgsql.Close(); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()
	if err := platformredis.Init(); err != nil {
		return err
	}
	defer func() {
		if err := platformredis.Close(); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()

	deps, err := initializedDependencies(ctx, cfg)
	if err != nil {
		return err
	}
	server := newHTTPServer(cfg, newHandler(cfg, deps), deps.tokens)
	return serve(ctx, server)
}

// initializedDependencies 在 main 完成数据库和 Redis 初始化后装配外部平台客户端。
func initializedDependencies(ctx context.Context, cfg config.Config) (*dependencies, error) {
	deps := &dependencies{database: pgsql.DB(), redis: platformredis.Default()}
	deps.tokens = platformauth.NewSessionManager(cfg.JWTSecret, cfg.JWTExpire, deps.redis)
	deps.wechat = wechat.NewClient(cfg.WeChatAppID, cfg.WeChatAppSecret, cfg.AllowMockLogin)
	if err := deps.wechat.Init(ctx); err != nil {
		return nil, fmt.Errorf("initialize WeChat: %w", err)
	}
	deps.bailian = bailian.NewClient(cfg.BailianAPIKey, cfg.BailianBaseURL, cfg.BailianASRModel)
	if cfg.BailianAPIKey == "" {
		log.Print("speech recognition disabled: configure BAILIAN_API_KEY or bailian.api_key")
	}
	if err := deps.bailian.Init(ctx); err != nil {
		return nil, fmt.Errorf("initialize Bailian: %w", err)
	}
	deps.llm = llm.NewClient(cfg.LLMAPIKey, cfg.LLMBaseURL, cfg.LLMModel, cfg.LLMThinkingMode)
	if cfg.LLMAPIKey == "" {
		log.Print("nutrition recognition disabled: configure LLM_API_KEY or llm.api_key")
	}
	if err := deps.llm.Init(); err != nil {
		return nil, fmt.Errorf("initialize LLM: %w", err)
	}
	return deps, nil
}

// newHandler 集中装配业务服务，使依赖关系保持清晰，并避免基础设施包反向依赖 Handler。
func newHandler(cfg config.Config, deps *dependencies) *handler.Handler {
	accounts := service.NewAccountService(deps.database, deps.wechat, deps.tokens, cfg.UploadDir)
	return handler.New(
		accounts,
		service.NewDiscoveryService(deps.database),
		service.NewRecordService(deps.database),
		service.NewReviewService(deps.database, deps.wechat),
		service.NewContributionService(deps.database),
		service.NewSpeechService(deps.bailian),
		service.NewEngagementService(deps.database),
		service.NewNutritionService(deps.database, deps.llm),
		service.NewMediaService(deps.database, cfg.UploadDir, cfg.PublicBaseURL),
		service.NewVerificationService(deps.database, deps.wechat, cfg.DataEncryptionKey),
	)
}

func newHTTPServer(cfg config.Config, h *handler.Handler, tokens *platformauth.SessionManager) *http.Server {
	return &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           router.New(h, tokens, cfg.CORSAllowOrigins, cfg.UploadDir),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// serve 负责报告监听错误，并在收到取消或进程信号后给予在途请求十秒钟完成。
func serve(ctx context.Context, server *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("NutriLens API listening on %s", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve API: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve API: %w", err)
	}
	return nil
}
