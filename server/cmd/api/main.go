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
	database "shijibu/internal/model/pgsql"
	platformredis "shijibu/internal/model/redis"
	platformauth "shijibu/internal/platform/auth"
	"shijibu/internal/platform/bailian"
	platformemail "shijibu/internal/platform/email"
	"shijibu/internal/platform/wechat"
	"shijibu/internal/router"
	"shijibu/internal/service"
)

// main 是后端进程入口，负责装配依赖、启动服务和退出时的收尾；具体业务在 service 层处理。
func main() {
	// 1. 加载并校验配置，默认读取工作目录下的 config/config.yaml。
	cfg, err := config.Load()
	if err != nil {
		panic("load configuration failed")
	}

	// 2. 通过 GORM 连接 PostgreSQL，并用 Ping 确认数据库可访问。
	db, err := database.New(cfg.DatabaseDSN)
	if err != nil {
		panic("connect PostgreSQL failed")
	}
	if err := database.Init(context.Background(), db); err != nil {
		panic("initialize PostgreSQL failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("close PostgreSQL failed")
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Print("close PostgreSQL failed")
		}
	}()

	if err := database.AutoMigrate(db); err != nil {
		panic("migrate PostgreSQL failed")
	}

	redis, err := platformredis.NewRedis(cfg.RedisDSN)
	if err != nil {
		panic("create Redis client failed")
	}
	defer func() {
		if err := redis.Close(); err != nil {
			log.Print("close Redis failed")
		}
	}()
	if err := redis.Init(context.Background()); err != nil {
		panic("initialize Redis failed")
	}
	tokens := platformauth.NewSessionManager(cfg.JWTSecret, cfg.JWTExpire, redis)
	wechatClient := wechat.NewClient(cfg.WeChatAppID, cfg.WeChatAppSecret, cfg.AllowMockLogin)
	if err := wechatClient.Init(context.Background()); err != nil {
		panic("initialize WeChat failed")
	}
	bailianClient := bailian.NewClient(cfg.BailianAPIKey, cfg.BailianBaseURL, cfg.BailianASRModel)
	if cfg.BailianAPIKey == "" {
		log.Print("speech recognition disabled: configure BAILIAN_API_KEY or bailian.api_key")
	}
	if err := bailianClient.Init(context.Background()); err != nil {
		panic("initialize Bailian failed")
	}

	// 4. 创建业务服务并传入它们需要的依赖，避免业务代码自行连接数据库或使用全局变量。
	accounts := service.NewAccountService(db, wechatClient, tokens, cfg.UploadDir)
	if cfg.SMTP.Host != "" {
		mailer, err := platformemail.NewClient(platformemail.Config{Host: cfg.SMTP.Host, Port: cfg.SMTP.Port, Username: cfg.SMTP.Username, Password: cfg.SMTP.Password, From: cfg.SMTP.From, TLSMode: cfg.SMTP.TLSMode})
		if err != nil {
			panic("invalid SMTP configuration")
		}
		if err := mailer.Init(context.Background()); err != nil {
			panic("initialize SMTP failed")
		}
		accounts.WithEmailLogin(redis, mailer, cfg.JWTSecret) // 复用 Redis，以用途隔离的 HMAC 保护验证码。
	} else {
		log.Print("email login disabled: configure SMTP settings")
	}

	speech := service.NewSpeechService(bailianClient)

	// 5. Handler 持有各业务服务，负责接收接口参数、调用业务逻辑并返回响应。
	h := handler.New(accounts, service.NewDiscoveryService(db), service.NewRecordService(db, wechatClient), service.NewContributionService(db), speech,
		service.NewEngagementService(db), service.NewNutritionService(db), service.NewMediaService(db, cfg.UploadDir, cfg.PublicBaseURL),
		service.NewVerificationService(db, wechatClient, cfg.DataEncryptionKey))

	// 6. Gin 负责路由和中间件，标准库 HTTP 服务器负责监听端口并接收请求。
	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           router.New(h, tokens, cfg.CORSAllowOrigins, cfg.UploadDir),
		ReadHeaderTimeout: 5 * time.Second,  // 读取请求头的最长时间。
		ReadTimeout:       60 * time.Second, // 读取整个请求（包括上传内容）的最长时间。
		WriteTimeout:      60 * time.Second, // 响应写入的超时限制。
		IdleTimeout:       60 * time.Second, // 保持连接时等待下一次请求的最长时间。
	}
	// 7. 监听 Ctrl+C（SIGINT）和容器停止等场景的 SIGTERM，收到信号后取消 ctx。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer 在 main 返回时执行，用于释放信号监听资源。
	defer stop()

	// 8. ListenAndServe 会持续阻塞，放到独立 goroutine 中，让主 goroutine 能等待退出信号。
	go func() {
		log.Printf("NutriLens API listening on %s", cfg.ServerAddr)
		// 正常关闭也会返回 ErrServerClosed，不应把它当作运行故障。
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve API: %v", err)
		}
	}()
	// 主 goroutine 在此等待退出信号；如果 main 提前返回，整个进程也会退出。
	<-ctx.Done()

	// 9. 退出前最多留 10 秒给正在处理的请求收尾。
	// 必须使用新的上下文，因为上面的 ctx 已取消，不能用它等待收尾。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 即使提前完成关闭，也及时释放计时器资源。
	defer cancel()
	// 停止接收新连接，等待现有请求完成；如果超时，记录错误后 main 返回。
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
