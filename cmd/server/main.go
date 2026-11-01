package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"user-age-api/config"
	"user-age-api/db/sqlc"
	"user-age-api/internal/handler"
	"user-age-api/internal/logger"
	"user-age-api/internal/middleware"
	"user-age-api/internal/repository"
	"user-age-api/internal/routes"
	"user-age-api/internal/service"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.Env, cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	db, err := connectDB(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("could not connect to database", zap.Error(err))
	}
	defer db.Close()

	// Wiring: sqlc -> repository -> service -> handler -> routes
	queries := sqlc.New(db)
	repo := repository.NewUserRepository(queries)
	svc := service.NewUserService(repo, log)
	userHandler := handler.NewUserHandler(svc, log)

	app := fiber.New(fiber.Config{
		AppName:      "user-age-api",
		ErrorHandler: handler.ErrorHandler(log),
	})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))
	routes.Register(app, userHandler)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info("shutting down server")
		_ = app.ShutdownWithTimeout(10 * time.Second)
	}()

	log.Info("starting server", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

// connectDB opens the pool and retries the initial ping (useful when Postgres is still starting in Docker).
func connectDB(url string, log *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	var lastErr error
	for i := 1; i <= 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		lastErr = db.PingContext(ctx)
		cancel()
		if lastErr == nil {
			log.Info("connected to database")
			return db, nil
		}
		log.Warn("database not ready, retrying", zap.Int("attempt", i), zap.Error(lastErr))
		time.Sleep(2 * time.Second)
	}
	return nil, errors.Join(errors.New("database unreachable after retries"), lastErr)
}
