package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/berndmarcel860-byte/ongame/pam/internal/api"
	"github.com/berndmarcel860-byte/ongame/pam/internal/auth"
	"github.com/berndmarcel860-byte/ongame/pam/internal/cache"
	"github.com/berndmarcel860-byte/ongame/pam/internal/db"
	"github.com/berndmarcel860-byte/ongame/pam/internal/service"
)

func main() {
	// Database
	database, err := db.NewPostgres(getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ongame?sslmode=disable"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database.DB, getEnv("MIGRATIONS_DIR", "./migrations")); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Redis
	redisClient, err := cache.NewRedis(getEnv("REDIS_URL", "redis://localhost:6379"))
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// Auth
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	pamAPIToken := getEnv("PAM_API_TOKEN", "change-me-pam-token")

	// Services
	userSvc := service.NewUserService(database, redisClient, jwtManager)
	balanceSvc := service.NewBalanceService(database)
	gameSvc := service.NewGameService(database, redisClient)

	// Router
	router := api.NewRouter(api.RouterConfig{
		UserService:    userSvc,
		BalanceService: balanceSvc,
		GameService:    gameSvc,
		JWTManager:     jwtManager,
		PAMAPIToken:    pamAPIToken,
	})

	addr := ":" + getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("PAM service listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server exited")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
